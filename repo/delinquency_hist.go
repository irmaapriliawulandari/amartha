package repo

import (
	"amartha-test/entity"
	"database/sql"
	"errors"
	"fmt"
)

func (r *loanRepo) InsertDelinquencyHist(dh entity.DelinquencyHistDB) (int64, error) {
	const query = `
		insert into delinquency_hist (
			borrower_id, loan_id, statement_id, cleared_at, updated_at
		)
		values ($1, $2, $3, $4, $5)
		on conflict (statement_id) do nothing
		returning dh_id
	`

	var dhID int64
	err := r.db.QueryRow(query,
		dh.BorrowerID,
		dh.LoanID,
		dh.StatementID,
		dh.ClearedAt,
		dh.UpdatedAt,
	).Scan(&dhID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("insert delinquency_hist: %w", err)
	}

	return dhID, nil
}

func (r *loanRepo) IsDelinquent(borrowerID int64) (bool, error) {
	const query = `
		select exists(
			select 1 from delinquency_hist
			where borrower_id = $1 and cleared_at is null
		)
	`

	var exists bool
	if err := r.db.QueryRow(query, borrowerID).Scan(&exists); err != nil {
		return false, fmt.Errorf("query is delinquent: %w", err)
	}

	return exists, nil
}

func (r *loanRepo) IsLoanDelinquent(borrowerID, loanID int64) (bool, error) {
	const query = `
		select exists(
			select 1 from delinquency_hist
			where borrower_id = $1 and loan_id = $2 and cleared_at is null
		)
	`

	var exists bool
	if err := r.db.QueryRow(query, borrowerID, loanID).Scan(&exists); err != nil {
		return false, fmt.Errorf("query is loan delinquent: %w", err)
	}

	return exists, nil
}

func (r *loanRepo) IsEverDelinquent(borrowerID int64) (bool, error) {
	const query = `
		select exists(
			select 1 from delinquency_hist
			where borrower_id = $1
		)
	`

	var exists bool
	if err := r.db.QueryRow(query, borrowerID).Scan(&exists); err != nil {
		return false, fmt.Errorf("query is ever delinquent: %w", err)
	}

	return exists, nil
}
