package persistance

import (
	"context"
	"database/sql"
	"fmt"
	"gong/common"
	"io/fs"
	"log"
	"os"
	"slices"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/questdb/go-questdb-client/v3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
)

type questDatabase struct {
	postgresDatabase
	client questdb.LineSender
}

func NewQuestDatabase(sqlMigrations fs.FS) (Database, error) {
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

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", hostname, username, password, name, port)
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	client, err := questdb.NewLineSender(context.Background(), questdb.WithTcp(), questdb.WithAddress(hostname+":9009"), questdb.WithBasicAuth(username, password))
	if err != nil {
		return nil, err
	}

	bunDB := bun.NewDB(sqlDB, pgdialect.New())

	bunDB.AddQueryHook(bundebug.NewQueryHook(
		// bundebug.FromEnv("BUNDEBUG"),
		bundebug.WithVerbose(true),
	))

	bunDB.SetMaxOpenConns(5) // requred to not overwhelm db server

	if err := bunDB.Ping(); err != nil {
		return nil, err
	}

	if sqlMigrations != nil {
		dbDriver, err := NewQuestDriver(bunDB, Config{
			MigrationsTableName: "migrations",
		})
		if err != nil {
			return nil, fmt.Errorf("error while creating db driver: %s", err.Error())
		}

		sourceDriver, err := iofs.New(sqlMigrations, "migration/quest")
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

	// 	_, err = bunDB.NewRaw(`CREATE TABLE IF NOT EXISTS 'events' (
	//   project_id SYMBOL INDEX,
	//   timestamp TIMESTAMP,
	//   pid SYMBOL INDEX,
	//   sender SYMBOL INDEX,
	//   type SYMBOL INDEX,
	//   level LONG,
	//   msg VARCHAR,
	//   attrs VARCHAR
	// ) TIMESTAMP(timestamp) PARTITION BY day;`).Exec(context.Background())
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	_, err = bunDB.NewCreateTable().IfNotExists().Model((*common.Actor)(nil)).Exec(context.Background())
	// 	if err != nil {
	// 		return nil, err
	// 	}

	log.Println("Database connection was successfully opened")
	return &questDatabase{
		postgresDatabase: postgresDatabase{
			bunDB: bunDB,
		},
		client: client,
	}, nil
}

func (db *questDatabase) QueryEvents(
	ctx context.Context,
	query EventsQuery,
) ([]common.Event, error) {

	var events []common.Event

	stmt := db.bunDB.NewSelect().Model(&events).Where("project_id = ?", query.ProjectID).Where("pid = ?", query.ActorPID)

	order := "timestamp DESC"
	if query.NumEvents > 0 {
		stmt = stmt.Where("timestamp > ?", query.Timestamp)
		order = "timestamp ASC"
	} else {
		stmt = stmt.Where("timestamp < ?", query.Timestamp)
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

	slices.SortFunc(events, func(a common.Event, b common.Event) int {
		return a.Timestamp.Compare(b.Timestamp)
	})

	return events, nil
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
			Int64Column("level", int64(e.Level)).
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
