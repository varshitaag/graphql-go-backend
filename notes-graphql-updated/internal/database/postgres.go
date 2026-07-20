package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// NewConnection opens a connection pool to Postgres and verifies it's reachable.
func NewConnection(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
