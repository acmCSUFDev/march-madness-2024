package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

//go:generate sqlc generate

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

	return &Database{
		Queries: New(db),
		db:      db,
	}, nil
}

// NewDatabaseInMemory creates a new in-memory database.
func NewDatabaseInMemory() *Database {
	db, _ := sql.Open("sqlite", ":memory:")
	return &Database{
		Queries: New(db),
		db:      db,
	}
}
