package usecase

import (
	"amartha-test/entity"
	"fmt"
	"time"
)

func (uc *billingEngineUsecase) GetStatements(loanID int64, until time.Time, limit, offset int) ([]entity.Statement, error) {
	statementsDB, err := uc.repo.GetStatements(loanID, until, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get statements: %w", err)
	}

	statements := make([]entity.Statement, 0, len(statementsDB))
	for _, s := range statementsDB {
		toPay := s.InstallmentAmount.Add(s.CarryOverAmount)
		statements = append(statements, entity.Statement{
			LoanID:        s.LoanID,
			StatementDate: s.StatementDate.Format("2006-01-02"),
			ToPayAmount:   toPay,
			Status:        entity.StatementStatus[s.Status],
		})
	}

	return statements, nil
}
