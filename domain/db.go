package domain

import (
	"database/sql"
	"fmt"
	"gong/.gen/gong/public/model"
	"gong/.gen/gong/public/table"
	"gong/types"
	"log"
	"log/slog"
	"os"
	"slices"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/oklog/ulid/v2"
)

type Database interface {
	QueryEvents(query eventsQuery) ([]types.Event, error)
	CreateEvents(events []types.Event) error
	QueryActors(projectID uuid.UUID) ([]types.Actor, error)
	CreateActor(types.Actor) error
	UpdateActor(types.Actor) error
	QueryActorSources(actorSourcesQuery) ([]string, error)
}

type eventsQuery struct {
	projectID uuid.UUID
	actorPID  string
	eventID   ulid.ULID
	numEvents int
	filter    types.EventsFilter
}

type actorSourcesQuery struct {
	projectID uuid.UUID
	actorPID  string
	minTime   time.Time
	maxTime   time.Time
}

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

	log.Println("Database connection was successfully opened")
	return &postgresDatabase{
		conn: conn,
	}
}

func sqlError(stmt Statement, err error) error {
	return fmt.Errorf("%s - %s", err.Error(), stmt.DebugSql())
}

func (db *postgresDatabase) CreateEvents(events []types.Event) error {
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
		tx, err := db.conn.Begin()
		if err != nil {
			return err
		}

		stmt, err := tx.Prepare("COPY event (id, project_id, actor_pid, index, timestamp, sender, type, level, msg, attrs) FROM STDIN")
		if err != nil {
			return err
		}

		for _, e := range m {
			_, err = stmt.Exec(e.ID, e.ProjectID, e.ActorPid, e.Index, e.Timestamp, e.Sender, e.Type, e.Level, e.Msg, e.Attrs)
			if err != nil {
				return err
			}
		}

		if _, err := stmt.Exec(); err != nil {
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

func (db *postgresDatabase) QueryEvents(query eventsQuery) ([]types.Event, error) {
	var m []model.Event

	whereExpression := table.Event.ProjectID.EQ(UUID(query.projectID)).AND(table.Event.ActorPid.EQ(String(query.actorPID)))
	orderClause := table.Event.ID.DESC()

	if query.eventID.Compare(ulid.ULID{}) != 0 {
		if query.numEvents > 0 {
			whereExpression = whereExpression.AND(table.Event.ID.GT(String(query.eventID.String())))
			orderClause = table.Event.ID.ASC()
		} else {
			whereExpression = whereExpression.AND(table.Event.ID.LT(String(query.eventID.String())))
		}
	}

	if query.filter.ID != "" {
		whereExpression = whereExpression.AND(table.Event.ID.LIKE(String("%" + query.filter.ID + "%")))
	}
	if query.filter.Type != "" {
		whereExpression = whereExpression.AND(table.Event.Type.LIKE(String("%" + query.filter.Type + "%")))
	}
	if query.filter.Sender != "" {
		whereExpression = whereExpression.AND(table.Event.Sender.LIKE(String("%" + query.filter.Sender + "%")))
	}
	if query.filter.Level > slog.LevelDebug {
		whereExpression = whereExpression.AND(table.Event.Level.GT_EQ(Int(int64(query.filter.Level))))
	}

	stmt := table.Event.SELECT(table.Event.AllColumns).WHERE(
		whereExpression).
		ORDER_BY(orderClause).
		LIMIT(int64(max(query.numEvents, -query.numEvents)))

	if err := stmt.Query(db.conn, &m); err != nil {
		return nil, sqlError(stmt, err)
	}

	if query.numEvents > 0 {
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

func (db *postgresDatabase) QueryActors(projectID uuid.UUID) ([]types.Actor, error) {
	var m []model.Actor

	stmt := table.Actor.SELECT(table.Actor.AllColumns).WHERE(table.Actor.ProjectID.EQ(UUID(projectID)))
	if err := stmt.Query(db.conn, &m); err != nil {
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

func (db *postgresDatabase) CreateActor(actor types.Actor) error {
	var m model.Actor

	m.Pid = actor.PID
	m.ProjectID = actor.ProjectID
	m.NumEvents = actor.NumEvents
	m.Alive = actor.Alive
	m.LastEventID = actor.LastEventID.String()

	stmt := table.Actor.INSERT(table.Actor.AllColumns).MODEL(&m)
	if _, err := stmt.Exec(db.conn); err != nil {
		return sqlError(stmt, err)
	}

	return nil
}

func (db *postgresDatabase) UpdateActor(actor types.Actor) error {
	var m model.Actor

	m.Pid = actor.PID
	m.ProjectID = actor.ProjectID
	m.NumEvents = actor.NumEvents
	m.Alive = actor.Alive
	m.LastEventID = actor.LastEventID.String()

	stmt := table.Actor.UPDATE(table.Actor.AllColumns).MODEL(&m).WHERE(table.Actor.ProjectID.EQ(UUID(actor.ProjectID)).AND(table.Actor.Pid.EQ(String(actor.PID))))
	if _, err := stmt.Exec(db.conn); err != nil {
		return sqlError(stmt, err)
	}

	return nil
}

func (db *postgresDatabase) QueryActorSources(query actorSourcesQuery) ([]string, error) {
	var m []struct {
		Sender string
	}

	stmt := table.Event.SELECT(table.Event.Sender.AS("sender")).WHERE(
		table.Event.ActorPid.EQ(String(query.actorPID)).
			AND(table.Event.Timestamp.BETWEEN(TimestampzT(query.minTime), TimestampzT(query.maxTime))).
			AND(NOT(table.Event.Sender.EQ(String(""))))).
		GROUP_BY(table.Event.Sender)

	if err := stmt.Query(db.conn, &m); err != nil {
		return nil, sqlError(stmt, err)
	}

	senders := make([]string, len(m))
	for i, sender := range m {
		senders[i] = sender.Sender
	}

	return senders, nil
}
