package persistance

import (
	"context"
	"database/sql"
	"fmt"
	"gong/common"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

const (
	envDatabaseHostname = "DB_HOSTNAME"
	envDatabasePort     = "DB_PORT"
	envDatabaseName     = "DB_NAME"
	envDatabaseUsername = "DB_USERNAME"
	envDatabasePassword = "DB_PASSWORD"
)

type postgresDatabase struct {
	bunDB *bun.DB
}

func NewPostgresDatabase() (Database, error) {
	hostname, ok := os.LookupEnv(envDatabaseHostname)
	if !ok {
		return nil, fmt.Errorf("%s not set", envDatabaseHostname)
	}

	port, ok := os.LookupEnv(envDatabasePort)
	if !ok {
		return nil, fmt.Errorf("%s not set", envDatabasePort)
	}

	name, ok := os.LookupEnv(envDatabaseName)
	if !ok {
		return nil, fmt.Errorf("%s not set", envDatabaseName)
	}

	username, ok := os.LookupEnv(envDatabaseUsername)
	if !ok {
		return nil, fmt.Errorf("%s not set", envDatabaseUsername)
	}

	password, ok := os.LookupEnv(envDatabasePassword)
	if !ok {
		return nil, fmt.Errorf("%s not set", envDatabasePassword)
	}

	connector := pgdriver.NewConnector(
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(hostname+":"+port),
		pgdriver.WithUser(username),
		pgdriver.WithPassword(password),
		pgdriver.WithDatabase(name),
		pgdriver.WithTLSConfig(nil),
	)

	sqlDB := sql.OpenDB(connector)

	bunDB := bun.NewDB(sqlDB, pgdialect.New())

	bunDB.AddQueryHook(bundebug.NewQueryHook(
		bundebug.FromEnv("BUNDEBUG"),
		bundebug.WithVerbose(true),
	))

	bunDB.SetMaxOpenConns(5) // requred to not overwhelm db server

	if err := bunDB.Ping(); err != nil {
		return nil, err
	}

	// if sqlMigrations != nil {
	// 	dbDriver, err := sqlite.WithInstance(db.DB, &sqlite.Config{})
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error while creating db driver: %s", err.Error())
	// 	}

	// 	sourceDriver, err := iofs.New(sqlMigrations, "migration")
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error while creating source driver: %s", err.Error())
	// 	}

	// 	migrations, err := migrate.NewWithInstance("sqlMigrations", sourceDriver, "sqlite", dbDriver)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error while creating db migrations: %s", err.Error())
	// 	}

	// 	if err := migrations.Up(); err != nil && err != migrate.ErrNoChange {
	// 		return nil, fmt.Errorf("error while applying db migrations: %s", err.Error())
	// 	}
	// }

	log.Println("Database connection was successfully opened")
	return &postgresDatabase{
		bunDB: bunDB,
	}, nil
}

func (db *postgresDatabase) QueryEvents(
	ctx context.Context,
	query EventsQuery,
) ([]common.Event, error) {

	var events []common.Event

	stmt := db.bunDB.NewSelect().Model(&events).Where("project_id = ?", query.ProjectID).Where("pid = ?", query.ActorPID)

	order := "id DESC"
	if query.EventID.Compare(ulid.ULID{}) != 0 {
		if query.NumEvents > 0 {
			stmt = stmt.Where("id > ?", query.EventID.String())
			order = "id ASC"
		} else {
			stmt = stmt.Where("id < ?", query.EventID.String())
		}
	}

	if query.Filter.ID != nil {
		stmt = stmt.Where("id = ?", *query.Filter.ID)
	}

	if query.Filter.Type != nil {
		stmt = stmt.Where("type = ?", *query.Filter.Type)
	}

	if query.Filter.Sender != nil {
		stmt = stmt.Where("sender = ?", *query.Filter.Sender)
	}

	if query.Filter.Level != nil {
		stmt = stmt.Where("level = ?", *query.Filter.Level)
	}

	err := stmt.Order(order).Limit(max(query.NumEvents, -query.NumEvents)).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (db *postgresDatabase) CreateEvents(
	ctx context.Context,
	events []common.Event,
) error {
	if len(events) == 0 {
		return nil
	}

	_, err := db.bunDB.NewInsert().Model(&events).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (db *postgresDatabase) QueryActors(
	ctx context.Context,
	projectID uuid.UUID,
) ([]common.Actor, error) {
	var actors []common.Actor

	err := db.bunDB.NewSelect().Model(&actors).Where("project_id = ?", projectID.String()).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return actors, nil
}

func (db *postgresDatabase) CreateActor(
	ctx context.Context, actor common.Actor,
) error {
	_, err := db.bunDB.NewInsert().Model(&actor).Exec(ctx)
	return err
}

func (db *postgresDatabase) UpdateActor(
	ctx context.Context,
	actor common.Actor,
) error {
	_, err := db.bunDB.NewUpdate().Model(&actor).Where("project_id = ? AND pid = ?", actor.ProjectID, actor.PID).Exec(ctx)
	return err
}

func (db *postgresDatabase) QueryActorSourcesSinks(
	ctx context.Context,
	query ActorSourcesSinksQuery,
) ([]string, error) {
	var actors []struct {
		bun.BaseModel `bun:"table:events"`
		PID           string
		Sender        string
	}

	stmt := db.bunDB.NewSelect().
		Model(&actors).
		Where("project_id = ?", query.ProjectID.String()).
		Where("timestamp BETWEEN ? AND ?", query.MinTime, query.MaxTime)

	if query.Sources {
		stmt = stmt.Where("pid = ?", query.ActorPID).DistinctOn("sender").Order("sender").ExcludeColumn("pid")
	} else {
		stmt = stmt.Where("sender = ?", query.ActorPID).DistinctOn("pid").Order("pid").ExcludeColumn("sender")
	}

	if err := stmt.Scan(ctx); err != nil {
		return nil, err
	}

	pids := make([]string, 0, len(actors))
	for _, actor := range actors {
		if query.Sources {
			if actor.Sender != "" {
				pids = append(pids, actor.Sender)
			}
		} else {
			if actor.PID != "" {
				pids = append(pids, actor.PID)
			}
		}
	}

	return pids, nil
}

func (db *postgresDatabase) QuerySourcesSinksNodeInfo(
	ctx context.Context,
	query SourcesSinksNodeInfoQuery,
) (info common.SourcesSinksNodeInfo, err error) {
	var actorPid string
	var sender string

	if query.Source {
		actorPid = query.ActorPID
		sender = query.NodePID
	} else {
		actorPid = query.NodePID
		sender = query.ActorPID
	}

	err = db.bunDB.NewSelect().
		Table("events").
		ColumnExpr("count(*) AS msg_count").
		Where("project_id = ?", query.ProjectID.String()).
		Where("pid = ?", actorPid).
		Where("sender = ?", sender).
		Where("timestamp BETWEEN ? AND ?", query.MinTime, query.MaxTime).
		Scan(ctx, &info)

	info.PID = query.NodePID
	return info, err
}
