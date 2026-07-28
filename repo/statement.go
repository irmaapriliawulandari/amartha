package repo

import (
	"amartha-test/entity"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (r *loanRepo) GetStatements(loanID int64, until time.Time, limit, offset int) ([]entity.StatementDB, error) {
	const query = `
		select statement_id, installment_amount, carry_over_amount, paid_amount, statement_date, status
		from statement
		where loan_id = $1 and statement_date < $2
		order by statement_date desc
		limit $3 offset $4
	`

	rows, err := r.db.Query(query, loanID, until, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query statement: %w", err)
	}
	defer rows.Close()

	var statements []entity.StatementDB
	for rows.Next() {
		s := entity.StatementDB{LoanID: loanID}
		if err := rows.Scan(&s.StatementID, &s.InstallmentAmount, &s.CarryOverAmount, &s.PaidAmount, &s.StatementDate, &s.Status); err != nil {
			return nil, fmt.Errorf("scan statement row: %w", err)
		}
		statements = append(statements, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate statement rows: %w", err)
	}

	return statements, nil
}

func (r *loanRepo) GetLatestStatement(loanID int64, before time.Time) (entity.StatementDB, error) {
	const query = `
		select statement_id, installment_amount, carry_over_amount, paid_amount, statement_date, deadline, status
		from statement
		where loan_id = $1 and statement_date < $2
		order by statement_date desc
		limit 1
	`

	s := entity.StatementDB{LoanID: loanID}
	err := r.db.QueryRow(query, loanID, before).Scan(
		&s.StatementID, &s.InstallmentAmount, &s.CarryOverAmount, &s.PaidAmount, &s.StatementDate, &s.Deadline, &s.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.StatementDB{}, entity.ErrStatementNotFound
		}
		return entity.StatementDB{}, fmt.Errorf("query latest statement: %w", err)
	}

	return s, nil
}

func insertStatement(q queryRower, statement entity.StatementDB) (int64, error) {
	const query = `
		insert into statement (
			loan_id, installment_amount, carry_over_amount, paid_amount,
			statement_date, deadline, status, paid_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		on conflict (loan_id, statement_date) do nothing
		returning statement_id
	`

	var statementID int64
	err := q.QueryRow(query,
		statement.LoanID,
		statement.InstallmentAmount,
		statement.CarryOverAmount,
		statement.PaidAmount,
		statement.StatementDate,
		statement.Deadline,
		statement.Status,
		statement.PaidAt,
		statement.UpdatedAt,
	).Scan(&statementID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("insert statement: %w", err)
	}

	return statementID, nil
}

func (r *loanRepo) InsertStatement(statement entity.StatementDB) (int64, error) {
	return insertStatement(r.db, statement)
}

// MakePayment pays off a loan's latest statement (installment + carry over),
// marks any of its still-overdue prior statements as paid late, and clears
// the loan's open delinquency records — all in one transaction, with the
// latest statement row locked for the duration to prevent a concurrent
// double payment.
func (r *loanRepo) MakePayment(loanID int64, now time.Time) (entity.StatementDB, error) {
	const selectLatestForUpdate = `
		select statement_id, installment_amount, carry_over_amount, paid_amount, statement_date, deadline, status
		from statement
		where loan_id = $1 and statement_date < $2
		order by statement_date desc
		limit 1
		for update
	`
	const updateStatement = `
		update statement
		set paid_amount = $1, status = $2, paid_at = $3, updated_at = $3
		where statement_id = $4
	`
	const updateOverdueStatements = `
		update statement
		set status = $1, paid_at = $2, updated_at = $2
		where loan_id = $3 and status = $4
	`
	const clearDelinquency = `
		update delinquency_hist
		set cleared_at = $1, updated_at = $1
		where loan_id = $2 and cleared_at is null
	`

	tx, err := r.db.Begin()
	if err != nil {
		return entity.StatementDB{}, fmt.Errorf("begin transaction: %w", err)
	}

	s := entity.StatementDB{LoanID: loanID}
	err = tx.QueryRow(selectLatestForUpdate, loanID, now).Scan(
		&s.StatementID, &s.InstallmentAmount, &s.CarryOverAmount, &s.PaidAmount, &s.StatementDate, &s.Deadline, &s.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.StatementDB{}, rollback(tx, entity.ErrStatementNotFound)
		}
		return entity.StatementDB{}, rollback(tx, fmt.Errorf("query latest statement: %w", err))
	}

	if s.Status == entity.StatementStatusPaid || s.Status == entity.StatementStatusPaidLate {
		return entity.StatementDB{}, rollback(tx, entity.ErrStatementAlreadyPaid)
	}

	paidAmount := s.InstallmentAmount.Add(s.CarryOverAmount)
	if _, err := tx.Exec(updateStatement, paidAmount, entity.StatementStatusPaid, now, s.StatementID); err != nil {
		return entity.StatementDB{}, rollback(tx, fmt.Errorf("update statement: %w", err))
	}

	if _, err := tx.Exec(updateOverdueStatements, entity.StatementStatusPaidLate, now, loanID, entity.StatementStatusOverdue); err != nil {
		return entity.StatementDB{}, rollback(tx, fmt.Errorf("update overdue statements: %w", err))
	}

	if _, err := tx.Exec(clearDelinquency, now, loanID); err != nil {
		return entity.StatementDB{}, rollback(tx, fmt.Errorf("clear delinquency_hist: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return entity.StatementDB{}, fmt.Errorf("commit transaction: %w", err)
	}

	s.PaidAmount = paidAmount
	s.Status = entity.StatementStatusPaid
	s.PaidAt = sql.NullTime{Time: now, Valid: true}

	return s, nil
}
