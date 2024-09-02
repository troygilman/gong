package persistance

import (
	"context"
	"database/sql"
	"fmt"
	"stagebase/internal/common"
	"io/fs"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/questdb/go-questdb-client/v3"
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

type questDatabase struct {
	bunDB  *bun.DB
	client questdb.LineSender
}

func NewQuestDatabase(sqlMigrations fs.FS) (common.Database, error) {
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
		// bundebug.FromEnv("BUNDEBUG"),
		bundebug.WithVerbose(false),
	))

	bunDB.SetMaxOpenConns(5) // requred to not overwhelm db server

	if err := bunDB.Ping(); err != nil {
		return nil, err
	}

	if sqlMigrations != nil {
		dbDriver, err := NewQuestDriver(sqlDB, Config{
			MigrationsTableName: "migrations",
			Verbose:             true,
		})
		if err != nil {
			return nil, fmt.Errorf("error while creating db driver: %s", err.Error())
		}

		sourceDriver, err := iofs.New(sqlMigrations, "migrations/quest")
		if err != nil {
			return nil, fmt.Errorf("error while creating source driver: %s", err.Error())
		}

		migrations, err := migrate.NewWithInstance("sqlMigrations", sourceDriver, "postgres", dbDriver)
		if err != nil {
			return nil, fmt.Errorf("error while creating db migrations: %s", err.Error())
		}

		if err := migrations.Up(); err != nil && err != migrate.ErrNoChange {
			return nil, fmt.Errorf("error while applying db migrations: %s", err.Error())
		}
	}

	client, err := questdb.NewLineSender(context.Background(), questdb.WithTcp(), questdb.WithAddress(hostname+":9009"), questdb.WithBasicAuth(username, password))
	if err != nil {
		return nil, err
	}

	log.Println("Database connection was successfully opened")
	return &questDatabase{
		bunDB:  bunDB,
		client: client,
	}, nil
}

func (db *questDatabase) QueryEvents(
	ctx context.Context,
	query common.EventsQuery,
) ([]common.Event, int, error) {
	pageSize := 100

	subq := db.bunDB.NewSelect().Model(&common.Event{}).
		Where("project_id = ?", query.ProjectID).
		Where("pid = ?", query.ActorPID).
		Where("timestamp <= ?", query.Timestamp).
		Order("timestamp DESC")

	if query.Filter.Type != nil {
		subq = subq.Where("type ~ ?", *query.Filter.Type)
	}

	if query.Filter.Sender != nil {
		subq = subq.Where("sender ~ ?", *query.Filter.Sender)
	}

	subq = subq.Where("level >= ?", strconv.Itoa(int(query.Filter.Level)))

	count, err := subq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	maxPageIndex := count / pageSize

	lowerLimit := query.PageIndex * pageSize
	upperLimit := (query.PageIndex + 1) * pageSize

	var events []common.Event
	if err := db.bunDB.NewRaw("SELECT * FROM (?) LIMIT ?, ?", subq, lowerLimit, upperLimit).Scan(ctx, &events); err != nil {
		return nil, 0, err
	}

	return events, maxPageIndex, nil
}

func (db *questDatabase) CreateEvents(
	ctx context.Context,
	events []common.Event,
) error {
	if len(events) == 0 {
		return nil
	}

	for _, e := range events {
		err := db.client.Table("events").
			Symbol("project_id", e.ProjectID).
			Symbol("pid", e.PID).
			Symbol("sender", e.Sender).
			Symbol("type", e.Type).
			Symbol("level", e.Level).
			StringColumn("msg", e.Msg).
			StringColumn("attrs", e.Attrs).
			At(ctx, e.Timestamp)
		if err != nil {
			return err
		}
	}

	if err := db.client.Flush(ctx); err != nil {
		return err
	}

	return nil
}

func (db *questDatabase) QueryActors(
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

func (db *questDatabase) CreateActor(
	ctx context.Context, actor common.Actor,
) error {
	_, err := db.bunDB.NewInsert().Model(&actor).Exec(ctx)
	return err
}

func (db *questDatabase) UpdateActor(
	ctx context.Context,
	actor common.Actor,
) error {
	_, err := db.bunDB.NewUpdate().Model(&actor).Where("project_id = ? AND pid = ?", actor.ProjectID, actor.PID).Exec(ctx)
	return err
}

func (db *questDatabase) QueryActorSourcesSinks(
	ctx context.Context,
	query common.ActorSourcesSinksQuery,
) ([]string, error) {
	var actors []struct {
		bun.BaseModel `bun:"table:events"`
		PID           string
		Sender        string
	}

	stmt := db.bunDB.NewSelect().
		Model(&actors).
		Distinct().
		Where("project_id = ?", query.ProjectID.String()).
		Where("timestamp BETWEEN ? AND ?", query.MinTime, query.MaxTime)

	if query.Sources {
		stmt = stmt.Where("pid = ?", query.ActorPID).Order("sender").ExcludeColumn("pid")
	} else {
		stmt = stmt.Where("sender = ?", query.ActorPID).Order("pid").ExcludeColumn("sender")
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

func (db *questDatabase) QuerySourcesSinksNodeInfo(
	ctx context.Context,
	query common.SourcesSinksNodeInfoQuery,
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
