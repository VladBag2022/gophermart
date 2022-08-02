package storage

import (
	"context"
	"database/sql"

	"github.com/georgysavva/scany/sqlscan"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type PostgresRepository struct {
	database *sql.DB
}

func NewPostgresRepository(
	ctx context.Context,
	databaseDSN string,
) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}
	p := &PostgresRepository{
		database: db,
	}
	err = p.createSchema(ctx)
	return p, err
}

func (p *PostgresRepository) Ping(ctx context.Context) error {
	return p.database.PingContext(ctx)
}

func (p *PostgresRepository) Close() error {
	return p.database.Close()
}

func (p *PostgresRepository) createSchema(ctx context.Context) error {
	tables := []string{
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
			"accrual REAL, " +
			"withdrawal REAL, " +
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

func (p *PostgresRepository) IsLoginAvailable(
	ctx context.Context,
	login string,
) (available bool, err error) {
	var count int64
	row := p.database.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE login = $1", login)
	err = row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, err
}

func (p *PostgresRepository) Register(
	ctx context.Context,
	login, password string,
) error {
	_, err := p.database.ExecContext(ctx,
		"INSERT INTO users (login, password) VALUES ($1, crypt($2, gen_salt('bf'))",
		login, password)
	return err
}

func (p *PostgresRepository) Login(
	ctx context.Context,
	login, password string,
) (success bool, err error) {
	var count int64
	row := p.database.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE login = $1 AND password = crypt($2, gen_salt('bf')",
		login, password)
	err = row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, err
}

func (p *PostgresRepository) OrderOwner(
	ctx context.Context,
	order int,
) (login string, err error) {
	row := p.database.QueryRowContext(ctx,
		"SELECT login FROM users JOIN orders ON users.id = orders.user_id AND orders.id = $1",
		order)
	err = row.Scan(&login)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return login, err
}

func (p *PostgresRepository) UploadOrder(
	ctx context.Context,
	login string,
	order int,
) error {
	_, err := p.database.ExecContext(ctx,
		"INSERT INTO orders (id, user_id) SELECT $1, id FROM users WHERE login = $2",
		order, login)
	return err
}

func (p *PostgresRepository) Orders(
	ctx context.Context,
	login string,
) (orders []OrderInfo, err error) {
	err = sqlscan.Select(ctx, p.database, &orders,
		"SELECT orders.id, orders.status, orders.uploaded_at FROM orders "+
			"JOIN users ON orders.user_id = users.id AND users.login = $1", login)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (p *PostgresRepository) Balance(
	ctx context.Context,
	login string,
) (current, withdrawn float64, err error) {
	row := p.database.QueryRowContext(ctx,
		"SELECT SUM(accrual), SUM(withdrawal) FROM orders JOIN users ON users.id = orders.user_id AND users.login = $1",
		login)
	err = row.Scan(&current, &withdrawn)
	if err != nil {
		return 0.0, 0.0, err
	}
	return current - withdrawn, withdrawn, nil
}

func (p *PostgresRepository) Withdraw(
	ctx context.Context,
	login string,
	order int,
	sum float64,
) error {
	_, err := p.database.ExecContext(ctx,
		"INSERT INTO orders (id, user_id, withdrawal) SELECT $1, id, $2 FROM users WHERE login = $3",
		order, sum, login)
	return err
}

func (p *PostgresRepository) Withdrawals(
	ctx context.Context,
	login string,
) (withdrawals []WithdrawalInfo, err error) {
	err = sqlscan.Select(ctx, p.database, &withdrawals,
		"SELECT orders.id, orders.withdrawal, orders.uploaded_at FROM orders "+
			"JOIN users ON orders.user_id = users.id AND users.login = $1", login)
	if err != nil {
		return nil, err
	}
	return withdrawals, nil
}
