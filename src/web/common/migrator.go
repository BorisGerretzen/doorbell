package common

import (
	"database/sql"
	"fmt"
	"log"
	"web/common/migrations"
)

type Migration interface {
	Version() int
	Description() string
	Up() []string
	Down() []string
}

type Migrator struct {
	migrations []Migration
}

func NewMigrator() *Migrator {
	return &Migrator{
		migrations: []Migration{
			&migrations.Migration202409182046{},
		},
	}
}

func (m *Migrator) Migrate(db *sql.DB) error {
	// create migrations table if note xists
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS migrations (version INTEGER PRIMARY KEY, description TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		return err
	}

	_, err = db.Exec("PRAGMA foreign_keys=ON")
	if err != nil {
		return err
	}

	transaction, err := db.Begin()
	if err != nil {
		return err
	}

	// get current version
	var currentVersion sql.NullInt64
	err = transaction.QueryRow("SELECT MAX(version) FROM migrations").Scan(&currentVersion)
	if err != nil {
		return err
	}

	version := 0
	if currentVersion.Valid {
		version = int(currentVersion.Int64)
	}

	log.Printf("Current version: %d", version)

	// apply migrations
	for _, migration := range m.migrations {
		if migration.Version() > version {
			log.Printf("Applying migration %d: %s", migration.Version(), migration.Description())
			for _, query := range migration.Up() {
				_, err = transaction.Exec(query)
				if err != nil {
					if err2 := transaction.Rollback(); err2 != nil {
						return fmt.Errorf("migration failed: %s, rollback failed: %s", err, err2)
					}

					return err
				}
			}

			_, err = transaction.Exec("INSERT INTO migrations (version, description) VALUES (?, ?)", migration.Version(), migration.Description())
			if err != nil {
				if err2 := transaction.Rollback(); err2 != nil {
					return fmt.Errorf("migration failed: %s, rollback failed: %s", err, err2)
				}
				return err
			}
		}
	}

	return transaction.Commit()
}
