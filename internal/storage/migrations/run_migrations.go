// Package migrations хранит SQL-миграции и применяет их к базе через goose.
package migrations

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var migrationFS embed.FS

// RunMigrations применяет все неприменённые миграции к базе данных db.
func RunMigrations(db *sql.DB) error {
	goose.SetBaseFS(migrationFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, ".")
}
