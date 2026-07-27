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
