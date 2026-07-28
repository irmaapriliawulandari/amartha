package repo

import (
	"amartha-test/entity"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
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

// GetLatestStatementForUpdate locks and returns a loan's most recent
// statement dated before `before`, for the caller to inspect and mutate
// within the same transaction.
func (r *loanRepo) GetLatestStatementForUpdate(tx Tx, loanID int64, before time.Time) (entity.StatementDB, error) {
	t, err := sqlTx(tx)
	if err != nil {
		return entity.StatementDB{}, err
	}

	const query = `
		select statement_id, installment_amount, carry_over_amount, paid_amount, statement_date, deadline, status
		from statement
		where loan_id = $1 and statement_date < $2
		order by statement_date desc
		limit 1
		for update
	`

	s := entity.StatementDB{LoanID: loanID}
	err = t.QueryRow(query, loanID, before).Scan(
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

// UpdateStatementPaid marks a statement fully paid.
func (r *loanRepo) UpdateStatementPaid(tx Tx, statementID int64, paidAmount decimal.Decimal, now time.Time) error {
	t, err := sqlTx(tx)
	if err != nil {
		return err
	}

	const query = `
		update statement
		set paid_amount = $1, status = $2, paid_at = $3, updated_at = $3
		where statement_id = $4
	`

	if _, err := t.Exec(query, paidAmount, entity.StatementStatusPaid, now, statementID); err != nil {
		return fmt.Errorf("update statement: %w", err)
	}

	return nil
}

// MarkPriorOverdueAsPaidLate marks every currently-overdue statement of a
// loan as paid late.
func (r *loanRepo) MarkPriorOverdueAsPaidLate(tx Tx, loanID int64, now time.Time) error {
	t, err := sqlTx(tx)
	if err != nil {
		return err
	}

	const query = `
		update statement
		set status = $1, paid_at = $2, updated_at = $2
		where loan_id = $3 and status = $4
	`

	if _, err := t.Exec(query, entity.StatementStatusPaidLate, now, loanID, entity.StatementStatusOverdue); err != nil {
		return fmt.Errorf("update overdue statements: %w", err)
	}

	return nil
}

// ListOverdueCandidates returns every statement of an active loan that is
// still unpaid with the given deadline.
func (r *loanRepo) ListOverdueCandidates(deadline time.Time) ([]entity.OverdueCandidate, error) {
	const query = `
		select s.statement_id, s.loan_id, l.borrower_id
		from statement s
		join loan l on l.loan_id = s.loan_id
		where s.status = $1 and s.deadline = $2 and l.status = $3
	`

	rows, err := r.db.Query(query, entity.StatementStatusUnpaid, deadline, entity.LoanStatusActive)
	if err != nil {
		return nil, fmt.Errorf("query overdue candidates: %w", err)
	}
	defer rows.Close()

	var candidates []entity.OverdueCandidate
	for rows.Next() {
		var c entity.OverdueCandidate
		if err := rows.Scan(&c.StatementID, &c.LoanID, &c.BorrowerID); err != nil {
			return nil, fmt.Errorf("scan overdue candidate: %w", err)
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate overdue candidates: %w", err)
	}

	return candidates, nil
}

// GetStatementForUpdate locks and returns a single statement by id.
func (r *loanRepo) GetStatementForUpdate(tx Tx, statementID int64) (entity.StatementDB, error) {
	t, err := sqlTx(tx)
	if err != nil {
		return entity.StatementDB{}, err
	}

	const query = `
		select loan_id, installment_amount, carry_over_amount, paid_amount, statement_date, deadline, status
		from statement
		where statement_id = $1
		for update
	`

	s := entity.StatementDB{StatementID: statementID}
	err = t.QueryRow(query, statementID).Scan(
		&s.LoanID, &s.InstallmentAmount, &s.CarryOverAmount, &s.PaidAmount, &s.StatementDate, &s.Deadline, &s.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.StatementDB{}, entity.ErrStatementNotFound
		}
		return entity.StatementDB{}, fmt.Errorf("query statement: %w", err)
	}

	return s, nil
}

// MarkStatementOverdue marks a single statement overdue.
func (r *loanRepo) MarkStatementOverdue(tx Tx, statementID int64, now time.Time) error {
	t, err := sqlTx(tx)
	if err != nil {
		return err
	}

	const query = `
		update statement
		set status = $1, updated_at = $2
		where statement_id = $3 and status = $4
	`

	if _, err := t.Exec(query, entity.StatementStatusOverdue, now, statementID, entity.StatementStatusUnpaid); err != nil {
		return fmt.Errorf("update statement: %w", err)
	}

	return nil
}

// GetNextStatementForUpdate locks and returns the id of the statement
// immediately following afterDate for a loan. found is false if this was
// the loan's last installment.
func (r *loanRepo) GetNextStatementForUpdate(tx Tx, loanID int64, afterDate time.Time) (statementID int64, found bool, err error) {
	t, err := sqlTx(tx)
	if err != nil {
		return 0, false, err
	}

	const query = `
		select statement_id
		from statement
		where loan_id = $1 and statement_date > $2
		order by statement_date asc
		limit 1
		for update
	`

	err = t.QueryRow(query, loanID, afterDate).Scan(&statementID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("query next statement: %w", err)
	}

	return statementID, true, nil
}

// UpdateCarryOver sets a statement's carry_over_amount.
func (r *loanRepo) UpdateCarryOver(tx Tx, statementID int64, carryOverAmount decimal.Decimal, now time.Time) error {
	t, err := sqlTx(tx)
	if err != nil {
		return err
	}

	const query = `
		update statement
		set carry_over_amount = $1, updated_at = $2
		where statement_id = $3
	`

	if _, err := t.Exec(query, carryOverAmount, now, statementID); err != nil {
		return fmt.Errorf("update statement carry over: %w", err)
	}

	return nil
}

// GetPreviousStatementStatus returns the status of the statement
// immediately before beforeDate for a loan. found is false if this was the
// loan's first installment.
func (r *loanRepo) GetPreviousStatementStatus(tx Tx, loanID int64, beforeDate time.Time) (status int, found bool, err error) {
	t, err := sqlTx(tx)
	if err != nil {
		return 0, false, err
	}

	const query = `
		select status
		from statement
		where loan_id = $1 and statement_date < $2
		order by statement_date desc
		limit 1
	`

	err = t.QueryRow(query, loanID, beforeDate).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("query previous statement: %w", err)
	}

	return status, true, nil
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
