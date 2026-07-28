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
