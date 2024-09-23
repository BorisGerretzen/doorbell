package database

import (
	"database/sql"
	_ "github.com/glebarez/go-sqlite"
)

type Database struct {
	Db *sql.DB
}

func NewDatabase(connectionString string) (*Database, error) {
	db, err := sql.Open("sqlite", connectionString)
	if err != nil {
		return nil, err
	}

	return &Database{
		Db: db,
	}, nil
}

func (d *Database) Close() error {
	return d.Db.Close()
}
