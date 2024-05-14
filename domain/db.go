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

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/oklog/ulid/v2"
)

type Database interface {
	CreateEvent(event *types.Event) error
	QueryEvents(query eventsQuery) ([]types.Event, error)
}

type eventsQuery struct {
	projectID uuid.UUID
	actorPID  string
	startIdx  int32
	endIdx    int32
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
		ID:        uuid.UUID(event.ID),
		ProjectID: event.ProjectID,
		ActorPid:  event.PID,
		Index:     event.Index,
		Timestamp: event.Timestamp,
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

	stmt := table.Event.SELECT(table.Event.AllColumns).WHERE(
		table.Event.ProjectID.EQ(UUID(query.projectID)).
			AND(table.Event.ActorPid.EQ(String(query.actorPID))).
			AND(table.Event.Index.BETWEEN(Int32(query.startIdx), Int32(query.endIdx)))).
		ORDER_BY(table.Event.ID.DESC())

	if err := stmt.Query(db.conn, &m); err != nil {
		return nil, err
	}

	events := make([]types.Event, len(m))
	for idx := range m {
		events[idx].ID = ulid.ULID(m[idx].ID)
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
