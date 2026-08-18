package migrate

import (
	"database/sql"
	"fmt"

	"phaseline/db/migrations"

	"github.com/pressly/goose/v3"
)

func Up(sqlDB *sql.DB) error {
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}
	if err := goose.Up(sqlDB, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
