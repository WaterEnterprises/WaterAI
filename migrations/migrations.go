package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

// RunMigrations brings the database up to the latest version.
func RunMigrations(db *sql.DB) error {
	// Set the dialect (change to "postgres" if not using SQLite)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	// We register all migrations here in order.
	// Migration 1: Initial Schema
	goose.AddMigrationContext(func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			CREATE TABLE session (
				id VARCHAR(36) NOT NULL PRIMARY KEY,
				workspace_dir TEXT NOT NULL UNIQUE,
				created_at DATETIME,
				device_id TEXT
			);
			CREATE TABLE event (
				id VARCHAR(36) NOT NULL PRIMARY KEY,
				session_id VARCHAR(36) NOT NULL,
				timestamp DATETIME,
				event_type TEXT NOT NULL,
				event_payload JSON NOT NULL,
				FOREIGN KEY (session_id) REFERENCES session(id) ON DELETE CASCADE
			);
		`)
		return err
	}, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "DROP TABLE event; DROP TABLE session;")
		return err
	})

	// Migration 2: Add name column and populate from first user message
	goose.AddMigrationContext(func(ctx context.Context, tx *sql.Tx) error {
		// 1. Add column
		if _, err := tx.ExecContext(ctx, "ALTER TABLE session ADD COLUMN name TEXT;"); err != nil {
			return err
		}
		// 2. Data Migration: Extract text from JSON payload
		// This uses the SQLite/Postgres '->>' operator for JSON path navigation
		_, err := tx.ExecContext(ctx, `
			UPDATE session
			SET name = (
				SELECT event.event_payload->>'$.content.text'
				FROM event
				WHERE event.session_id = session.id
				AND event.event_type = 'user_message'
				ORDER BY event.timestamp ASC
				LIMIT 1
			) WHERE name IS NULL;
		`)
		return err
	}, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "ALTER TABLE session DROP COLUMN name;")
		return err
	})

	// Migration 3: Add sandbox column
	goose.AddMigrationContext(func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, "ALTER TABLE session ADD COLUMN sandbox_id TEXT;"); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, "UPDATE session SET sandbox_id = id WHERE sandbox_id IS NULL;")
		return err
	}, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "ALTER TABLE session DROP COLUMN sandbox_id;")
		return err
	})

	// Run migrations
	return goose.Up(db, ".")
}