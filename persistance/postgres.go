package persistance

import (
	"context"
	"database/sql"
	"fmt"
	"gong/.gen/gong/public/model"
	"gong/.gen/gong/public/table"
	"gong/common"
	"log"
	"log/slog"
	"os"
	"slices"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/oklog/ulid/v2"
)

const (
	envDatabaseHostname = "DB_HOSTNAME"
	envDatabasePort     = "DB_PORT"
	envDatabaseName     = "DB_NAME"
	envDatabaseUsername = "DB_USERNAME"
	envDatabasePassword = "DB_PASSWORD"
)

type postgresDatabase struct {
	conn *sql.DB
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

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", hostname, username, password, name, port)
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres database: %s", err.Error())
	}

	conn.SetMaxOpenConns(5) // requred to not overwhelm db server

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	SetQueryLogger(func(ctx context.Context, info QueryInfo) {
		// TO-DO
		//log.Printf("%+v\n%s\n", info, info.Statement.DebugSql())
	})

	log.Println("Database connection was successfully opened")
	return &postgresDatabase{
		conn: conn,
	}, nil
}

func sqlError(stmt Statement, err error) error {
	return fmt.Errorf("%s - %s", err.Error(), stmt.DebugSql())
}

func (db *postgresDatabase) CreateEvents(ctx context.Context, events []common.Event) error {
	m := make([]model.Event, len(events))
	for i := range len(events) {
		m[i].ID = events[i].ID.String()
		m[i].ProjectID = events[i].ProjectID
		m[i].ActorPid = events[i].PID
		m[i].Index = events[i].Index
		m[i].Timestamp = events[i].Timestamp
		m[i].Sender = events[i].Sender
		m[i].Type = events[i].Type
		m[i].Level = int32(events[i].Level)
		m[i].Msg = events[i].Msg
		m[i].Attrs = events[i].Attrs
	}

	if len(m) > 0 {
		tx, err := db.conn.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		stmt, err := tx.PrepareContext(ctx, "COPY event (project_id, timestamp, id, actor_pid, index, type, level, sender, msg, attrs) FROM STDIN")
		if err != nil {
			return err
		}

		for _, e := range m {
			_, err = stmt.ExecContext(ctx, e.ProjectID, e.Timestamp, e.ID, e.ActorPid, e.Index, e.Type, e.Level, e.Sender, e.Msg, e.Attrs)
			if err != nil {
				return err
			}
		}

		if _, err := stmt.ExecContext(ctx); err != nil {
			return err
		}

		if err := stmt.Close(); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

func (db *postgresDatabase) QueryEvents(ctx context.Context, query EventsQuery) ([]common.Event, error) {
	var m []model.Event

	whereExpression := table.Event.ProjectID.EQ(UUID(query.ProjectID)).AND(table.Event.ActorPid.EQ(String(query.ActorPID)))
	orderClause := table.Event.ID.DESC()

	if query.EventID.Compare(ulid.ULID{}) != 0 {
		if query.NumEvents > 0 {
			whereExpression = whereExpression.AND(table.Event.ID.GT(String(query.EventID.String())))
			orderClause = table.Event.ID.ASC()
		} else {
			whereExpression = whereExpression.AND(table.Event.ID.LT(String(query.EventID.String())))
		}
	}

	if query.Filter.ID != nil {
		whereExpression = whereExpression.AND(table.Event.ID.EQ(String(*query.Filter.ID)))
	}
	if query.Filter.Type != nil {
		whereExpression = whereExpression.AND(table.Event.Type.EQ(String(*query.Filter.Type)))
	}
	if query.Filter.Sender != nil {
		whereExpression = whereExpression.AND(table.Event.Sender.EQ(String(*query.Filter.Sender)))
	}
	if query.Filter.Level != nil {
		whereExpression = whereExpression.AND(table.Event.Level.GT_EQ(Int(int64(*query.Filter.Level))))
	}

	stmt := table.Event.SELECT(table.Event.AllColumns).WHERE(
		whereExpression).
		ORDER_BY(orderClause).
		LIMIT(int64(max(query.NumEvents, -query.NumEvents)))

	if err := stmt.QueryContext(ctx, db.conn, &m); err != nil {
		return nil, sqlError(stmt, err)
	}

	if query.NumEvents > 0 {
		slices.Reverse(m)
	}

	var err error
	events := make([]common.Event, len(m))
	for idx := range m {
		events[idx].ID, err = ulid.Parse(m[idx].ID)
		if err != nil {
			return nil, err
		}
		events[idx].ProjectID = m[idx].ProjectID
		events[idx].Index = m[idx].Index
		events[idx].PID = m[idx].ActorPid
		events[idx].Sender = m[idx].Sender
		events[idx].Timestamp = m[idx].Timestamp
		events[idx].Type = m[idx].Type
		events[idx].Level = slog.Level(m[idx].Level)
		events[idx].Msg = m[idx].Msg
		events[idx].Attrs = m[idx].Attrs
	}

	return events, nil
}

func (db *postgresDatabase) QueryActors(ctx context.Context, projectID uuid.UUID) ([]common.Actor, error) {
	var m []model.Actor

	stmt := table.Actor.SELECT(table.Actor.AllColumns).WHERE(table.Actor.ProjectID.EQ(UUID(projectID)))
	if err := stmt.QueryContext(ctx, db.conn, &m); err != nil {
		return nil, sqlError(stmt, err)
	}

	var err error
	actors := make([]common.Actor, len(m))
	for idx, mItem := range m {
		actors[idx].PID = mItem.Pid
		actors[idx].ProjectID = mItem.ProjectID
		actors[idx].NumEvents = mItem.NumEvents
		actors[idx].Alive = mItem.Alive
		actors[idx].LastEventID, err = ulid.Parse(mItem.LastEventID)
		if err != nil {
			return nil, err
		}
	}

	return actors, nil
}

func (db *postgresDatabase) CreateActor(ctx context.Context, actor common.Actor) error {
	var m model.Actor

	m.Pid = actor.PID
	m.ProjectID = actor.ProjectID
	m.NumEvents = actor.NumEvents
	m.Alive = actor.Alive
	m.LastEventID = actor.LastEventID.String()

	stmt := table.Actor.INSERT(table.Actor.AllColumns).MODEL(&m)
	if _, err := stmt.ExecContext(ctx, db.conn); err != nil {
		return sqlError(stmt, err)
	}

	return nil
}

func (db *postgresDatabase) UpdateActor(ctx context.Context, actor common.Actor) error {
	var m model.Actor

	m.Pid = actor.PID
	m.ProjectID = actor.ProjectID
	m.NumEvents = actor.NumEvents
	m.Alive = actor.Alive
	m.LastEventID = actor.LastEventID.String()

	stmt := table.Actor.UPDATE(table.Actor.AllColumns).MODEL(&m).WHERE(table.Actor.ProjectID.EQ(UUID(actor.ProjectID)).AND(table.Actor.Pid.EQ(String(actor.PID))))
	if _, err := stmt.ExecContext(ctx, db.conn); err != nil {
		return sqlError(stmt, err)
	}

	return nil
}

func (db *postgresDatabase) QueryActorSourcesSinks(ctx context.Context, query ActorSourcesSinksQuery) ([]string, error) {
	var m []struct {
		Pid string
	}

	var stmt SelectStatement
	if query.Sources {
		condition := table.Event.ProjectID.EQ(UUID(query.ProjectID)).AND(table.Event.ActorPid.EQ(String(query.ActorPID))).AND(table.Event.Timestamp.BETWEEN(TimestampzT(query.MinTime), TimestampzT(query.MaxTime)))
		stmt = SELECT(table.Event.Sender.AS("pid")).DISTINCT(table.Event.Sender).FROM(table.Event).WHERE(condition).ORDER_BY(table.Event.Sender)
	} else {
		condition := table.Event.ProjectID.EQ(UUID(query.ProjectID)).AND(table.Event.Sender.EQ(String(query.ActorPID))).AND(table.Event.Timestamp.BETWEEN(TimestampzT(query.MinTime), TimestampzT(query.MaxTime)))
		stmt = SELECT(table.Event.ActorPid.AS("pid")).DISTINCT(table.Event.ActorPid).FROM(table.Event).WHERE(condition).ORDER_BY(table.Event.ActorPid)
	}

	if err := stmt.QueryContext(ctx, db.conn, &m); err != nil {
		return nil, sqlError(stmt, err)
	}

	pids := make([]string, 0, len(m))
	for _, mItem := range m {
		if mItem.Pid != "" {
			pids = append(pids, mItem.Pid)
		}
	}

	return pids, nil
}
