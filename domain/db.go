package domain

import (
	"database/sql"
	"fmt"
	"gong/.gen/gong/public/model"
	"gong/.gen/gong/public/table"
	"gong/types"
	"log"
	"os"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Database interface {
	CreateEvent(event *types.Event) error
	QuerySession(projectID uuid.UUID) (int32, error)
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
		Session:   event.Session,
		Index:     event.Index,
		Timestamp: event.Timestamp,
		Type:      event.MsgType,
		Level:     int32(event.Level),
		Msg:       event.Msg,
	}

	stmt := table.Event.INSERT(table.Event.AllColumns).MODEL(&m)
	if _, err := stmt.Exec(db.conn); err != nil {
		return sqlError(stmt, err)
	}

	return nil
}

func (db *postgresDatabase) QuerySession(projectID uuid.UUID) (int32, error) {
	stmt := SELECT(MAX(table.Event.Session).AS("event.session")).FROM(table.Event).WHERE(table.Event.ProjectID.EQ(UUID(projectID)))
	m := model.Event{}
	if err := stmt.Query(db.conn, &m); err != nil {
		return 0, sqlError(stmt, err)
	}
	return m.Session, nil
}
