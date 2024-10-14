package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/romanp1989/gophermart/internal/domain"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{
		db: db,
	}
}

func (d *DBStorage) CreateUser(ctx context.Context, user domain.User) (*domain.User, error) {
	var id domain.UserID
	var pgErr *pgconn.PgError

	insertQuery := `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING (id)`
	err := d.db.QueryRowContext(ctx, insertQuery, user.Login, user.Password).Scan(&id)
	if err != nil {
		if errors.As(err, pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, domain.ErrLoginExists
		}

		return nil, err
	}

	user.ID = id

	return &user, nil
}

func (d *DBStorage) FindByLogin(ctx context.Context, login string) (*domain.User, error) {
	var foundUser *domain.User

	searchUserQuery := "SELECT id, login, password FROM users WHERE login = $1"
	row := d.db.QueryRowContext(ctx, searchUserQuery, login)
	if err := row.Scan(&foundUser.ID, &foundUser.Login, &foundUser.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("cannot scan row: %w", err)
	}

	return foundUser, nil
}
