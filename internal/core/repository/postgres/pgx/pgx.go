// Package pgx contains helpers shared by all pgx-backed repositories: a common
// Querier abstraction (so repos work with both a *pgxpool.Pool and a tx) and
// PostgreSQL error inspection helpers.
package pgx

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Querier is implemented by *pgxpool.Pool, pgx.Conn and pgx.Tx. Repositories
// depend on this interface so they can run inside a transaction transparently.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// PostgreSQL error codes we care about (see: https://www.postgresql.org/docs/current/errcodes-appendix.html).
const (
	codeUniqueViolation     = "23505"
	codeForeignKeyViolation = "23503"
)

// IsNoRows reports whether err is pgx.ErrNoRows.
func IsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

// UniqueViolation returns the offending constraint name and true when err is a
// unique-constraint violation.
func UniqueViolation(err error) (constraint string, ok bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == codeUniqueViolation {
		return pgErr.ConstraintName, true
	}
	return "", false
}

// ForeignKeyViolation reports whether err is a foreign-key violation.
func ForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == codeForeignKeyViolation
}
