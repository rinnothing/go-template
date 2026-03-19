package db

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rinnothing/avito-pr/internal/model"
	"github.com/rinnothing/avito-pr/pkg/transaction"
)

const (
	uniqueViolation = "23505"
)

func convertUniqueViolation(err error) (error, bool) {
	if err == nil {
		return nil, false
	}

	pgErr, ok := err.(*pgconn.PgError)
	if !ok {
		return nil, false
	}

	if pgErr.Code != uniqueViolation {
		return nil, false
	}
	return fmt.Errorf("%w: %w", model.ErrAlreadyExists, err), true
}

func convertNotFound(err error) (error, bool) {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %w", model.ErrNotFound, err), true
	}
	return nil, false
}

type postgresRepository struct {
	db *pgxpool.Pool
	t  transaction.Transactor
}

func NewPostgresRepository(db *pgxpool.Pool, t transaction.Transactor) *postgresRepository {
	return &postgresRepository{
		db: db,
		t:  t,
	}
}
