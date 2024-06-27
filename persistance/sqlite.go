package persistance

import (
	"context"
	"database/sql"
	"fmt"
	"gong/common"
	"io/fs"
	"log"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
)

type DatabaseProducer func() (*sql.DB, error)

func WithSQLiteFile(filePath string) DatabaseProducer {
	return func() (*sql.DB, error) {
		return sql.Open(sqliteshim.ShimName, fmt.Sprintf("file:%s?cache=shared", filePath))
	}
}

func WithTurso(name string, token string) DatabaseProducer {
	return func() (*sql.DB, error) {
		return sql.Open("libsql", fmt.Sprintf("libsql://%s.turso.io?authToken=%s", name, token))
	}
}

type sqliteDatabase struct {
	bunDB *bun.DB
}

func NewSQLiteDatabase(sqlMigrations fs.FS, producer DatabaseProducer) (Database, error) {
	sqlDB, err := producer()
	if err != nil {
		return nil, err
	}

	bunDB := bun.NewDB(sqlDB, sqlitedialect.New())

	bunDB.AddQueryHook(bundebug.NewQueryHook(
		// bundebug.FromEnv("BUNDEBUG"),
		bundebug.WithVerbose(false),
	))

	bunDB.SetMaxOpenConns(5) // requred to not overwhelm db server

	if err := bunDB.Ping(); err != nil {
		return nil, err
	}

	if sqlMigrations != nil {
		dbDriver, err := sqlite.WithInstance(sqlDB, &sqlite.Config{})
		if err != nil {
			return nil, fmt.Errorf("error while creating db driver: %s", err.Error())
		}

		sourceDriver, err := iofs.New(sqlMigrations, "migration/sqlite")
		if err != nil {
			return nil, fmt.Errorf("error while creating source driver: %s", err.Error())
		}

		migrations, err := migrate.NewWithInstance("sqlMigrations", sourceDriver, "sqlite", dbDriver)
		if err != nil {
			return nil, fmt.Errorf("error while creating db migrations: %s", err.Error())
		}

		if err := migrations.Up(); err != nil && err != migrate.ErrNoChange {
			return nil, fmt.Errorf("error while applying db migrations: %s", err.Error())
		}
	}

	log.Println("Database connection was successfully opened")
	return &sqliteDatabase{
		bunDB: bunDB,
	}, nil
}

type sqliteEvent struct {
	bun.BaseModel `bun:"table:events"`
	ProjectID     string
	PID           string
	Timestamp     int64
	Type          string
	Level         slog.Level
	Sender        string
	Msg           string
	Attrs         string
}

func (e *sqliteEvent) Parse(o common.Event) {
	e.ProjectID = o.ProjectID
	e.PID = o.PID
	e.Timestamp = o.Timestamp.UnixMicro()
	e.Type = o.Type
	e.Level = o.Level
	e.Sender = o.Sender
	e.Msg = o.Msg
	e.Attrs = o.Attrs
}

func (e sqliteEvent) Scan(o *common.Event) (err error) {
	o.ProjectID = e.ProjectID
	o.PID = e.PID
	o.Timestamp = time.UnixMicro(e.Timestamp)
	o.Type = e.Type
	o.Level = e.Level
	o.Sender = e.Sender
	o.Msg = e.Msg
	o.Attrs = e.Attrs
	return nil
}

func (db *sqliteDatabase) QueryEvents(
	ctx context.Context,
	query EventsQuery,
) ([]common.Event, error) {

	var events []sqliteEvent

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

	outEvents := make([]common.Event, len(events))
	for i, e := range events {
		if err := e.Scan(&outEvents[i]); err != nil {
			return nil, err
		}
	}

	return outEvents, nil
}

func (db *sqliteDatabase) CreateEvents(
	ctx context.Context,
	events []common.Event,
) error {
	if len(events) == 0 {
		return nil
	}

	sqlEvents := make([]sqliteEvent, len(events))
	for i, e := range events {
		sqlEvents[i].Parse(e)
	}

	_, err := db.bunDB.NewInsert().Model(&sqlEvents).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (db *sqliteDatabase) QueryActors(
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

func (db *sqliteDatabase) CreateActor(
	ctx context.Context, actor common.Actor,
) error {
	_, err := db.bunDB.NewInsert().Model(&actor).Exec(ctx)
	return err
}

func (db *sqliteDatabase) UpdateActor(
	ctx context.Context,
	actor common.Actor,
) error {
	_, err := db.bunDB.NewUpdate().Model(&actor).Where("project_id = ? AND pid = ?", actor.ProjectID, actor.PID).Exec(ctx)
	return err
}

func (db *sqliteDatabase) QueryActorSourcesSinks(
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
		Distinct().
		Where("project_id = ?", query.ProjectID.String()).
		Where("timestamp BETWEEN ? AND ?", query.MinTime.UnixMicro(), query.MaxTime.UnixMicro())

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

func (db *sqliteDatabase) QuerySourcesSinksNodeInfo(
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
		Where("timestamp BETWEEN ? AND ?", query.MinTime.UnixMicro(), query.MaxTime.UnixMicro()).
		Scan(ctx, &info)

	info.PID = query.NodePID
	return info, err
}
