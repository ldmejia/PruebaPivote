package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

func Migrate(ctx context.Context, db *sql.DB, files fs.FS) error {
	_, err := db.ExecContext(ctx, `IF OBJECT_ID('schema_migrations') IS NULL
CREATE TABLE schema_migrations (
    version VARCHAR(100) NOT NULL PRIMARY KEY,
    applied_at DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME()
)`)
	if err != nil {
		return fmt.Errorf("creando schema_migrations: %w", err)
	}
	names, err := fs.Glob(files, "*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)
	for _, name := range names {
		var applied int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(1) FROM schema_migrations WHERE version = @p1", name).Scan(&applied); err != nil {
			return err
		}
		if applied > 0 {
			continue
		}
		content, err := fs.ReadFile(files, name)
		if err != nil {
			return err
		}
		if err := apply(ctx, db, name, string(content)); err != nil {
			return fmt.Errorf("migracion %s: %w", name, err)
		}
	}
	return nil
}

func apply(ctx context.Context, db *sql.DB, name, content string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, batch := range splitBatches(content) {
		if _, err := tx.ExecContext(ctx, batch); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (@p1)", name); err != nil {
		return err
	}
	return tx.Commit()
}

func splitBatches(content string) []string {
	var batches []string
	var current strings.Builder
	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		if strings.EqualFold(strings.TrimSpace(line), "GO") {
			if s := strings.TrimSpace(current.String()); s != "" {
				batches = append(batches, s)
			}
			current.Reset()
			continue
		}
		current.WriteString(line)
		current.WriteString("\n")
	}
	if s := strings.TrimSpace(current.String()); s != "" {
		batches = append(batches, s)
	}
	return batches
}
