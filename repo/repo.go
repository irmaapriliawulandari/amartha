package repo

import (
	"database/sql"
	"fmt"
)

type loanRepo struct {
	db *sql.DB
}

func NewLoanRepo(db *sql.DB) *loanRepo {
	return &loanRepo{
		db: db,
	}
}

// execer and queryRower are satisfied by both *sql.DB and *sql.Tx, letting
// insert helpers run either standalone or inside a transaction.
type execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

type queryRower interface {
	QueryRow(query string, args ...any) *sql.Row
}

// Tx is an opaque database transaction handle. The usecase layer holds one
// across a sequence of repo calls to run them atomically and to decide,
// between calls, whether to keep going — without needing to know anything
// about database/sql. Repo owns all SQL and transaction mechanics; callers
// only ever see Commit/Rollback.
type Tx interface {
	Commit() error
	Rollback() error
}

func (r *loanRepo) BeginTx() (Tx, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	return tx, nil
}

// sqlTx recovers the concrete *sql.Tx backing a Tx handle. Every repo
// method that accepts a Tx parameter uses this to run its query within it.
func sqlTx(tx Tx) (*sql.Tx, error) {
	t, ok := tx.(*sql.Tx)
	if !ok {
		return nil, fmt.Errorf("invalid transaction handle")
	}

	return t, nil
}
