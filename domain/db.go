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

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/oklog/ulid/v2"
)

type Database interface {
	CreateEvent(event *types.Event) error
	QueryEvents(query eventsQuery) ([]types.Event, error)
	QueryActorProps(projectID uuid.UUID) ([]types.ProjectActorProps, error)
}

type eventsQuery struct {
	projectID uuid.UUID
	actorPID  string
	eventID   ulid.ULID
	numEvents int
	filter    types.EventsFilter
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

func (db *postgresDatabase) CreateEvent(event *types.Event) error {
	m := model.Event{
		ID:        event.ID.String(),
		ProjectID: event.ProjectID,
		ActorPid:  event.PID,
		Index:     event.Index,
		Timestamp: event.Timestamp,
		Sender:    event.Sender,
		Type:      event.Type,
		Level:     int32(event.Level),
		Msg:       event.Msg,
	}

	stmt := table.Event.INSERT(table.Event.AllColumns).MODEL(&m)
	if _, err := stmt.Exec(db.conn); err != nil {
		return sqlError(stmt, err)
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

	if query.filter.Type != nil {
		whereExpression = whereExpression.AND(table.Event.Type.LIKE(String("%" + *query.filter.Type + "%")))
	}
	if query.filter.Sender != nil {
		whereExpression = whereExpression.AND(table.Event.Sender.LIKE(String("%" + *query.filter.Sender + "%")))
	}

	stmt := table.Event.SELECT(table.Event.AllColumns).WHERE(
		whereExpression).
		ORDER_BY(orderClause).
		LIMIT(int64(max(query.numEvents, -query.numEvents)))

	if err := stmt.Query(db.conn, &m); err != nil {
		return nil, err
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
	}

	return events, nil
}

func (db *postgresDatabase) QueryActorProps(projectID uuid.UUID) ([]types.ProjectActorProps, error) {
	var m []struct {
		ActorPID      string
		NumEvents     int32
		FirstEventID  string
		LastEventID   string
		LastEventType string
	}

	subQuery := SELECT(
		table.Event.ActorPid.AS("actor_pid"),
		COUNT(table.Event.ID).AS("num_events"),
		MIN(table.Event.ID).AS("first_event_id"),
		MAX(table.Event.ID).AS("last_event_id"),
	).FROM(table.Event).GROUP_BY(table.Event.ProjectID, table.Event.ActorPid).AsTable("sub")

	stmt := SELECT(subQuery.AllColumns(), table.Event.Type.AS("last_event_type")).FROM(table.Event.RIGHT_JOIN(
		subQuery,
		table.Event.ID.EQ(StringColumn("last_event_id"))))

	if err := stmt.Query(db.conn, &m); err != nil {
		return nil, sqlError(stmt, err)
	}

	props := make([]types.ProjectActorProps, len(m))
	for idx, mItem := range m {
		firstEventID, err := ulid.Parse(mItem.FirstEventID)
		if err != nil {
			return nil, err
		}
		lastEventID, err := ulid.Parse(mItem.LastEventID)
		if err != nil {
			return nil, err
		}
		props[idx] = types.ProjectActorProps{
			PID:          mItem.ActorPID,
			NumEvents:    mItem.NumEvents,
			Alive:        isActorAlive(mItem.LastEventType),
			FirstEventID: firstEventID,
			LastEventID:  lastEventID,
		}
	}

	return props, nil
}
