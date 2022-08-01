package storage

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type PostgresRepository struct {
	database       *sql.DB
}

func NewPostgresRepository(
	ctx context.Context,
	databaseDSN string,
) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}
	var p = &PostgresRepository{
		database:       db,
	}
	err = p.createSchema(ctx)
	return p, err
}

func (p *PostgresRepository) Ping(ctx context.Context) error {
	if err := p.database.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func (p *PostgresRepository) Close() []error {
	var errs []error

	err := p.database.Close()
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

func (p *PostgresRepository) createSchema(ctx context.Context) error {
	var tables = []string{
		"CREATE EXTENSION IF NOT EXISTS pgcrypto",
		"CREATE TABLE IF NOT EXISTS users (" +
			"id SERIAL PRIMARY KEY, " +
			"login TEXT NOT NULL UNIQUE, " +
			"password TEXT NOT NULL)",
		"CREATE TABLE IF NOT EXISTS orders (" +
			"id INTEGER PRIMARY KEY, " +
			"user_id INTEGER NOT NULL, " +
			"uploaded_at TIMESTAMP NOT NULL DEFAULT Now(), " +
			"status TEXT NOT NULL DEFAULT \"NEW\", " +
			"accrual INTEGER, " +
			"withdrawal INTEGER, " +
			"FOREIGN KEY (user_id) REFERENCES users (id))",
	}
	for _, table := range tables {
		_, err := p.database.ExecContext(ctx, table)
		if err != nil {
			return err
		}
	}
	return nil
}
