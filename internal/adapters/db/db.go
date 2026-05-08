package db

import (
	"context"
	"database/sql"
	"fmt"
)

func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("db.Open: %w", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("db.Open ping: %w", err)
	}

	return db, nil
}

func RunMigrations(db *sql.DB, sql string) error {
	if _, err := db.Exec(sql); err != nil {
		return fmt.Errorf("db.RunMigrations: %w", err)
	}

	return nil
}
