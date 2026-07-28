package repo

import "database/sql"

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
