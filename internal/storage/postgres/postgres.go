package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/muneebasaleem65/omnistage/internal/storage"
)

type Postgres struct {
	db *sql.DB
}

var _ storage.Storage = (*Postgres)(nil)

func New(storagePath string) (*Postgres, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("pgx", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Postgres{db: db}, nil
}

func (p *Postgres) CreateUser(email, passwordHash, name string) (int64, error) {
	const op = "storage.postgres.CreateUser"

	var id int64

	err := p.db.QueryRow(
		`INSERT INTO USERS (email, password_hash, name)
		VALUES ($1, $2, $3)
		RETURNING id`,
		email, passwordHash, name,
	).Scan(&id)

	//check for duplicate email error
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
	}

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
