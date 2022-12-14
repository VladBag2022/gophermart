package storage

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/georgysavva/scany/sqlscan"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type PostgresRepository struct {
	database *sql.DB
}

type PostgresOrderInfo struct {
	Number     int64           `json:"number"`
	Status     string          `json:"status"`
	Accrual    sql.NullFloat64 `json:"accrual"`
	UploadedAt string          `json:"uploaded_at"`
}

type PostgresWithdrawalInfo struct {
	Order       string          `json:"order"`
	Sum         sql.NullFloat64 `json:"sum"`
	ProcessedAt string          `json:"processed_at"`
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
			"id BIGINT PRIMARY KEY, " +
			"user_id INTEGER NOT NULL, " +
			"uploaded_at TIMESTAMP NOT NULL DEFAULT Now(), " +
			"status TEXT NOT NULL DEFAULT 'NEW', " +
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
	var count int
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
		"INSERT INTO users (login, password) VALUES ($1, crypt($2, gen_salt('bf')))",
		login, password)
	return err
}

func (p *PostgresRepository) Login(
	ctx context.Context,
	login, password string,
) (success bool, err error) {
	var count int
	row := p.database.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE login = $1 AND password = crypt($2, password)",
		login, password)
	err = row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, err
}

func (p *PostgresRepository) OrderOwner(
	ctx context.Context,
	order int64,
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
	order int64,
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
	var pOrders []PostgresOrderInfo
	err = sqlscan.Select(ctx, p.database, &pOrders,
		"SELECT orders.id AS number, orders.status, orders.accrual, orders.uploaded_at FROM orders "+
			"JOIN users ON orders.user_id = users.id AND users.login = $1", login)
	if err != nil {
		return nil, err
	}
	for _, pOrder := range pOrders {
		accrual := 0.0
		if pOrder.Accrual.Valid {
			accrual = pOrder.Accrual.Float64
		}

		orders = append(orders, OrderInfo{
			Accrual:    accrual,
			Number:     strconv.FormatInt(pOrder.Number, 10),
			Status:     pOrder.Status,
			UploadedAt: pOrder.UploadedAt,
		})
	}
	return orders, nil
}

func (p *PostgresRepository) AccrualOrders(
	ctx context.Context,
) (orders []int64, err error) {
	err = sqlscan.Select(ctx, p.database, &orders,
		"SELECT id AS number FROM orders "+
			"WHERE status != 'INVALID' AND status != 'PROCESSED'")
	return
}

func (p *PostgresRepository) UpdateOrder(
	ctx context.Context,
	order int64,
	status string,
	accrual float64,
) error {
	_, err := p.database.ExecContext(ctx,
		"UPDATE orders SET status = $1, accrual = $2 WHERE id = $3",
		status, accrual, order)
	return err
}

func (p *PostgresRepository) Balance(
	ctx context.Context,
	login string,
) (balance BalanceInfo, err error) {
	row := p.database.QueryRowContext(ctx,
		"SELECT SUM(accrual), SUM(withdrawal) FROM orders JOIN users ON users.id = orders.user_id AND users.login = $1",
		login)
	var current, withdrawn sql.NullFloat64
	err = row.Scan(&current, &withdrawn)
	if err != nil {
		balance.Current = 0.0
		balance.Withdrawn = 0.0
	} else {
		if !current.Valid {
			current.Float64 = 0.0
		}
		if !withdrawn.Valid {
			withdrawn.Float64 = 0.0
		}
		balance.Withdrawn = withdrawn.Float64
		balance.Current = current.Float64 - withdrawn.Float64
	}
	return balance, err
}

func (p *PostgresRepository) Withdraw(
	ctx context.Context,
	login string,
	order int64,
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
	var pWithdrawals []PostgresWithdrawalInfo
	err = sqlscan.Select(ctx, p.database, &pWithdrawals,
		"SELECT orders.id AS order, orders.withdrawal AS sum, orders.uploaded_at AS processed_at FROM orders "+
			"JOIN users ON orders.user_id = users.id AND users.login = $1", login)
	if err != nil {
		return nil, err
	}
	for _, pWithdrawal := range pWithdrawals {
		if pWithdrawal.Sum.Valid {
			withdrawals = append(withdrawals, WithdrawalInfo{
				Sum:         pWithdrawal.Sum.Float64,
				Order:       pWithdrawal.Order,
				ProcessedAt: pWithdrawal.ProcessedAt,
			})
		}
	}
	return withdrawals, nil
}
