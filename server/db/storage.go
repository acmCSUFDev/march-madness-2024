package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "embed"

	"libdb.so/lazymigrate"

	_ "modernc.org/sqlite"
)

//go:generate sqlc generate

//go:embed sql_schema.sql
var schema string

const pragma = `
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
PRAGMA strict = ON;
`

// Database provides methods for interacting with the database.
// For now, it just wraps around sqlc's Queries because I'm lazy.
type Database struct {
	*Queries
	db *sql.DB
}

// NewDatabase creates a new database at the given path.
func NewDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	return newDatabase(db)
}

// NewDatabaseInMemory creates a new in-memory database.
func NewDatabaseInMemory() (*Database, error) {
	db, _ := sql.Open("sqlite", ":memory:")
	return newDatabase(db)
}

func newDatabase(db *sql.DB) (*Database, error) {
	if _, err := db.Exec(pragma); err != nil {
		return nil, err
	}

	schema := lazymigrate.NewSchema(schema)
	if err := schema.Migrate(context.Background(), db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &Database{
		Queries: New(db),
		db:      db,
	}, nil
}
