package persistance

import (
	"context"
	"io"
	"strings"
	"time"

	"database/sql"

	"github.com/golang-migrate/migrate/v4/database"
	"github.com/uptrace/bun"
)

type Config struct {
	StatementTimeout    time.Duration
	MigrationsTableName string
}

type questDriver struct {
	config Config
	db     *bun.DB
	conn   bun.Conn
	tx     bun.Tx
}

func NewQuestDriver(db *bun.DB, config Config) (database.Driver, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}

	conn, err := db.Conn(context.Background())
	if err != nil {
		return nil, err
	}

	d := &questDriver{
		config: config,
		db:     db,
		conn:   conn,
	}

	return d, d.ensureVersionTable()
}

func (d *questDriver) Open(url string) (database.Driver, error) {
	// conn, err := pgx.Connect(context.Background(), url)
	// if err != nil {
	// 	return nil, err
	// }

	// if err := conn.Ping(context.Background()); err != nil {
	// 	return nil, err
	// }

	// d.conn = conn
	return d, d.ensureVersionTable()
}

func (d *questDriver) Close() error {
	return d.conn.Close()
}

func (d *questDriver) Lock() error {
	tx, err := d.conn.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	d.tx = tx
	return nil
}

func (d *questDriver) Unlock() error {
	if err := d.tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (d *questDriver) Run(migration io.Reader) error {
	migr, err := io.ReadAll(migration)
	if err != nil {
		return err
	}
	return d.runStatement(migr)
}

func (d *questDriver) runStatement(statement []byte) error {
	ctx := context.Background()
	if d.config.StatementTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d.config.StatementTimeout)
		defer cancel()
	}
	query := string(statement)
	if strings.TrimSpace(query) == "" {
		return nil
	}
	if _, err := d.tx.Exec(query); err != nil {
		return database.Error{OrigErr: err, Err: "migration failed", Query: statement}
	}
	return nil
}

func (d *questDriver) SetVersion(version int, dirty bool) error {
	// Also re-write the schema version for nil dirty versions to prevent
	// empty schema version for failed down migration on the first migration
	// See: https://github.com/golang-migrate/migrate/issues/330
	if version >= 0 || (version == database.NilVersion && dirty) {
		values := map[string]interface{}{
			"version": version,
			"dirty":   dirty,
		}

		updateQuery := d.db.NewUpdate().Model(&values).TableExpr(d.config.MigrationsTableName).Where("version = ?", version)
		result, err := updateQuery.Exec(context.Background())
		if err != nil {
			return &database.Error{OrigErr: err, Query: []byte(updateQuery.String())}
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			insertQuery := d.db.NewInsert().Model(&values).TableExpr(d.config.MigrationsTableName)
			if _, err := insertQuery.Exec(context.Background()); err != nil {
				return &database.Error{OrigErr: err, Query: []byte(insertQuery.String())}
			}
		}
	}

	return nil
}

func (d *questDriver) Version() (version int, dirty bool, err error) {
	query := `SELECT version, dirty FROM ` + d.config.MigrationsTableName + ` LIMIT 1`
	err = d.db.QueryRow(query).Scan(&version, &dirty)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return database.NilVersion, false, nil
		default:
			return 0, false, &database.Error{OrigErr: err, Query: []byte(query)}
		}
	}
	return version, dirty, nil
}

func (d *questDriver) Drop() error {
	return nil
}

func (d *questDriver) ensureVersionTable() error {
	query := `CREATE TABLE IF NOT EXISTS ` + d.config.MigrationsTableName + ` (version int, dirty boolean)`
	if _, err := d.db.Exec(query); err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}
	return nil
}
