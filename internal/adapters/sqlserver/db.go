package sqlserver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	if err := ensureDatabase(ctx, dsn); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func ensureDatabase(ctx context.Context, dsn string) error {
	u, err := url.Parse(dsn)
	if err != nil {
		return err
	}
	q := u.Query()
	name := q.Get("database")
	if name == "" {
		return errors.New("SQLSERVER_URL no trae el parametro database")
	}
	q.Del("database")
	u.RawQuery = q.Encode()
	master, err := sql.Open("sqlserver", u.String())
	if err != nil {
		return err
	}
	defer master.Close()
	var lastErr error
	for i := 0; i < 15; i++ {
		if lastErr = master.PingContext(ctx); lastErr == nil {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	if lastErr != nil {
		return fmt.Errorf("sql server no responde: %w", lastErr)
	}
	_, err = master.ExecContext(ctx, `DECLARE @n sysname = @p1;
IF DB_ID(@n) IS NULL
BEGIN
    DECLARE @s nvarchar(300) = N'CREATE DATABASE ' + QUOTENAME(@n);
    EXEC(@s);
END`, name)
	return err
}
