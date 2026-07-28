package repo

import (
	"amartha-test/entity"
	"database/sql"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

func insertLoan(exec execer, loan entity.LoanDB) error {
	const query = `
		insert into loan (
			loan_id, borrower_id, principal_amount, rate, total_amount,
			installment_count, installment_type, currency,
			start_date, disbursed_at, metadata, status, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		on conflict (loan_id) do nothing
	`

	_, err := exec.Exec(query,
		loan.LoanID,
		loan.BorrowerID,
		loan.PrincipalAmount,
		loan.Rate,
		loan.TotalAmount,
		loan.InstallmentCount,
		loan.InstallmentType,
		loan.Currency,
		loan.StartDate,
		loan.DisbursedAt,
		loan.Metadata,
		loan.Status,
		loan.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert loan: %w", err)
	}

	return nil
}

func (r *loanRepo) InsertLoan(loan entity.LoanDB) error {
	return insertLoan(r.db, loan)
}

// InsertLoanWithStatements inserts a loan and its full billing schedule
// atomically: either every row commits, or none do. This makes retrying a
// failed disbursement message safe — a partial failure never leaves an
// orphaned loan with a missing or incomplete statement schedule, and a
// redelivered message that already fully committed is a safe no-op thanks
// to the ON CONFLICT DO NOTHING on both the loan and statement inserts.
func (r *loanRepo) InsertLoanWithStatements(loan entity.LoanDB, statements []entity.StatementDB) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := insertLoan(tx, loan); err != nil {
		return rollback(tx, err)
	}

	for _, statement := range statements {
		if _, err := insertStatement(tx, statement); err != nil {
			return rollback(tx, fmt.Errorf("insert statement %s: %w", statement.StatementDate.Format("2006-01-02"), err))
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *loanRepo) GetOutstandingAmount(loanID int64) (decimal.Decimal, error) {
	const query = `
		select l.total_amount - coalesce(sum(s.paid_amount), 0)
		from loan l
		left join statement s on s.loan_id = l.loan_id
		where l.loan_id = $1
		group by l.loan_id, l.total_amount
	`

	var outstanding decimal.Decimal
	err := r.db.QueryRow(query, loanID).Scan(&outstanding)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return decimal.Decimal{}, entity.ErrLoanNotFound
		}
		return decimal.Decimal{}, fmt.Errorf("query outstanding amount: %w", err)
	}

	return outstanding, nil
}

func rollback(tx *sql.Tx, cause error) error {
	if err := tx.Rollback(); err != nil {
		return fmt.Errorf("%w (rollback failed: %v)", cause, err)
	}

	return cause
}
