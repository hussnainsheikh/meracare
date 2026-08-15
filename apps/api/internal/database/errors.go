package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// PostgreSQL SQLSTATE codes the API distinguishes.
const (
	codeUniqueViolation     = "23505"
	codeForeignKeyViolation = "23503"
	codeCheckViolation      = "23514"
)

// IsUniqueViolation reports whether err is a unique-constraint violation for
// the named constraint or index.
//
// Matching the constraint by name keeps the caller specific: "this email is
// taken" and "this account already has a profile" are different outcomes and
// must not be conflated.
func IsUniqueViolation(err error, constraint string) bool {
	pgErr, ok := asPgError(err)
	return ok && pgErr.Code == codeUniqueViolation && pgErr.ConstraintName == constraint
}

// IsForeignKeyViolation reports whether err references a row that does not exist.
func IsForeignKeyViolation(err error) bool {
	pgErr, ok := asPgError(err)
	return ok && pgErr.Code == codeForeignKeyViolation
}

// IsCheckViolation reports whether err breaks a CHECK constraint, which means
// the API tried to store a value the schema does not recognise.
func IsCheckViolation(err error, constraint string) bool {
	pgErr, ok := asPgError(err)
	return ok && pgErr.Code == codeCheckViolation && pgErr.ConstraintName == constraint
}

func asPgError(err error) (*pgconn.PgError, bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr, true
	}
	return nil, false
}
