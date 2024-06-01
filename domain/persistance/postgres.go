package persistance

import (
	"context"
	"database/sql"
	"fmt"
	"gong/.gen/gong/public/model"
	"gong/.gen/gong/public/table"
	"gong/types"
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

func NewPostgresDatabase() Database {
	hostname, ok := os.LookupEnv(envDatabaseHostname)
	if !ok {
		log.Fatalf("%s not set", envDatabaseHostname)
	}

	port, ok := os.LookupEnv(envDatabasePort)
	if !ok {
		log.Fatalf("%s not set", envDatabasePort)
	}

	name, ok := os.LookupEnv(envDatabaseName)
	if !ok {
		log.Fatalf("%s not set", envDatabaseName)
	}

	username, ok := os.LookupEnv(envDatabaseUsername)
	if !ok {
		log.Fatalf("%s not set", envDatabaseUsername)
	}

	password, ok := os.LookupEnv(envDatabasePassword)
	if !ok {
		log.Fatalf("%s not set", envDatabasePassword)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", hostname, username, password, name, port)
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		panic("failed to connect to postgres database: " + err.Error())
	}

	conn.SetMaxOpenConns(5) // requred to not overwhelm db server

	SetQueryLogger(func(ctx context.Context, info QueryInfo) {
		// TO-DO
	})

	log.Println("Database connection was successfully opened")
	return &postgresDatabase{
		conn: conn,
	}
}

func sqlError(stmt Statement, err error) error {
	return fmt.Errorf("%s - %s", err.Error(), stmt.DebugSql())
}

func (db *postgresDatabase) CreateEvents(ctx context.Context, events []types.Event) error {
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

		stmt, err := tx.PrepareContext(ctx, "COPY event (id, project_id, actor_pid, index, timestamp, sender, type, level, msg, attrs) FROM STDIN")
		if err != nil {
			return err
		}

		for _, e := range m {
			_, err = stmt.ExecContext(ctx, e.ID, e.ProjectID, e.ActorPid, e.Index, e.Timestamp, e.Sender, e.Type, e.Level, e.Msg, e.Attrs)
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

func (db *postgresDatabase) QueryEvents(ctx context.Context, query EventsQuery) ([]types.Event, error) {
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

	if query.Filter.ID != "" {
		whereExpression = whereExpression.AND(table.Event.ID.EQ(String(query.Filter.ID)))
	}
	if query.Filter.Type != "" {
		whereExpression = whereExpression.AND(table.Event.Type.EQ(String(query.Filter.Type)))
	}
	if query.Filter.Sender != "" {
		whereExpression = whereExpression.AND(table.Event.Sender.EQ(String(query.Filter.Sender)))
	}
	if query.Filter.Level > slog.LevelDebug {
		whereExpression = whereExpression.AND(table.Event.Level.GT_EQ(Int(int64(query.Filter.Level))))
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
	events := make([]types.Event, len(m))
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

func (db *postgresDatabase) QueryActors(ctx context.Context, projectID uuid.UUID) ([]types.Actor, error) {
	var m []model.Actor

	stmt := table.Actor.SELECT(table.Actor.AllColumns).WHERE(table.Actor.ProjectID.EQ(UUID(projectID)))
	if err := stmt.QueryContext(ctx, db.conn, &m); err != nil {
		return nil, sqlError(stmt, err)
	}

	var err error
	actors := make([]types.Actor, len(m))
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

func (db *postgresDatabase) CreateActor(ctx context.Context, actor types.Actor) error {
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

func (db *postgresDatabase) UpdateActor(ctx context.Context, actor types.Actor) error {
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

func buildDistinctColumnValueQuery(t ReadableTable, c ColumnString, alias string, condition BoolExpression) Statement {
	cte := CTE("cte")
	cteVal := c.From(cte)

	l := LATERAL(
		SELECT(c).FROM(t).WHERE(
			c.GT(cteVal).AND(condition),
		).ORDER_BY(c).LIMIT(1),
	).AS("l")

	sub := SELECT(c.From(l)).FROM(cte.CROSS_JOIN(l))

	stmt := WITH_RECURSIVE(
		cte.AS(
			SELECT(c).FROM(t).WHERE(condition).ORDER_BY(c).LIMIT(1).
				UNION_ALL(sub)))(SELECT(cteVal.AS(alias)).FROM(cte))

	return stmt
}

func (db *postgresDatabase) QueryActorSources(ctx context.Context, query ActorSourcesSinksQuery) ([]string, error) {
	var m []struct {
		Source string
	}

	condition := table.Event.ActorPid.EQ(String(query.ActorPID)).AND(table.Event.Timestamp.BETWEEN(TimestampzT(query.MinTime), TimestampzT(query.MaxTime)))
	stmt := buildDistinctColumnValueQuery(table.Event, table.Event.Sender, "source", condition)

	if err := stmt.QueryContext(ctx, db.conn, &m); err != nil {
		return nil, sqlError(stmt, err)
	}

	sources := make([]string, 0, len(m))
	for _, source := range m {
		if source.Source != "" {
			sources = append(sources, source.Source)
		}
	}

	return sources, nil
}

func (db *postgresDatabase) QueryActorSinks(ctx context.Context, query ActorSourcesSinksQuery) ([]string, error) {
	var m []struct {
		Sink string
	}

	condition := table.Event.Sender.EQ(String(query.ActorPID)).AND(table.Event.Timestamp.BETWEEN(TimestampzT(query.MinTime), TimestampzT(query.MaxTime)))
	stmt := buildDistinctColumnValueQuery(table.Event, table.Event.ActorPid, "sink", condition)

	if err := stmt.QueryContext(ctx, db.conn, &m); err != nil {
		return nil, sqlError(stmt, err)
	}

	sinks := make([]string, 0, len(m))
	for _, sink := range m {
		if sink.Sink != "" {
			sinks = append(sinks, sink.Sink)
		}
	}

	return sinks, nil
}
