package repo

import (
	"amartha-test/entity"
	"fmt"
	"time"
)

func (r *loanRepo) GetStatements(loanID int64, until time.Time, limit, offset int) ([]entity.LoanStatementDB, error) {
	const query = `
		select statement_date, amount, status
		from loan_statement
		where loan_id = $1 and statement_date < $2
		order by statement_date desc
		limit $3 offset $4
	`

	rows, err := r.db.Query(query, loanID, until, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query loan_statement: %w", err)
	}
	defer rows.Close()

	var statements []entity.LoanStatementDB
	for rows.Next() {
		s := entity.LoanStatementDB{LoanID: loanID}
		if err := rows.Scan(&s.StatementDate, &s.Amount, &s.Status); err != nil {
			return nil, fmt.Errorf("scan loan_statement row: %w", err)
		}
		statements = append(statements, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate loan_statement rows: %w", err)
	}

	return statements, nil
}
