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
