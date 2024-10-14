package balance

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/romanp1989/gophermart/internal/domain"
	"strings"
	"time"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{
		db: db,
	}
}

func (d *DBStorage) GetUserBalance(ctx context.Context, userID domain.UserID) ([]*domain.BalanceSum, error) {
	var (
		placeholders []string
		vals         []interface{}
	)

	types := []domain.BalanceType{domain.BalanceTypeAdded, domain.BalanceTypeWithdrawn}

	vals = append(vals, userID)
	for i, v := range types {
		placeholders = append(placeholders, fmt.Sprintf("$%d",
			i+2,
		))

		vals = append(vals, v)
	}

	q := fmt.Sprintf(`SELECT b.type, b.user_id, SUM(b.sum) 
							FROM balance b 
							WHERE b.type IN (%s) AND user_id = $1 GROUP BY b.type, b.user_id;`, strings.Join(placeholders, ","))
	rows, err := d.db.QueryContext(ctx, q, vals...)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var sums []*domain.BalanceSum
	for rows.Next() {
		sb := &domain.BalanceSum{}
		var sum float64

		err = rows.Scan(&sb.Type, &sb.UserID, &sum)
		if err != nil {
			return nil, err
		}

		sb.Sum = sum
		sums = append(sums, sb)
	}

	return sums, nil
}

func (d *DBStorage) Withdraw(ctx context.Context, userID *domain.UserID, orderNumber string, sum float64) (*domain.Balance, error) {
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var currentBalance sql.NullFloat64
	err = tx.QueryRowContext(
		ctx,
		`SELECT sum(b.sum) - (
			SELECT sum(b2.sum) FROM balance b2 WHERE b2.user_id = $1 and b2.type = $3
			) as sum FROM balance b
		WHERE b.user_id = $1 and b.type = $2;`,
		userID, domain.BalanceTypeAdded, domain.BalanceTypeWithdrawn,
	).Scan(&currentBalance)
	if err != nil {
		return nil, err
	}

	if currentBalance.Valid && (currentBalance.Float64-sum) < 0 {
		return nil, domain.ErrBalanceInsufficientFunds
	}

	var id int64
	createdAt := time.Now()

	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO balance (order_number, sum, type, user_id, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id;`,
		orderNumber, sum, domain.BalanceTypeWithdrawn, userID, createdAt,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &domain.Balance{
		ID:          id,
		CreatedAt:   createdAt,
		OrderNumber: orderNumber,
		UserID:      *userID,
		Sum:         sum,
		Type:        domain.BalanceTypeWithdrawn,
	}, nil
}

func (d *DBStorage) GetUserWithdrawals(ctx context.Context, userID domain.UserID) ([]*domain.Balance, error) {
	rows, err := d.db.QueryContext(
		ctx,
		`SELECT id, created_at, order_number, user_id, sum, type
			FROM balance b 
			WHERE b.user_id = $1 AND b.type = $2 
			ORDER BY b.created_at DESC;`,
		userID,
		domain.BalanceTypeWithdrawn,
	)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	var userWithdrawals []*domain.Balance
	for rows.Next() {
		balance := &domain.Balance{}
		var sum float64

		err = rows.Scan(&balance.ID, &balance.CreatedAt, &balance.OrderNumber, &balance.UserID, &sum, &balance.Type)
		if err != nil {
			return nil, err
		}

		balance.Sum = sum
		userWithdrawals = append(userWithdrawals, balance)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return userWithdrawals, nil

}