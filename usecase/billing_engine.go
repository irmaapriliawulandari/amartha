package usecase

import (
	"amartha-test/entity"
	"fmt"
	"strconv"
	"time"
)

type repo interface {
	GetStatements(loanID int64, until time.Time, limit, offset int) ([]entity.LoanStatementDB, error)
}

type billingEngineUsecase struct {
	repo repo
}

func NewBillingEngineUsecase(repo repo) *billingEngineUsecase {
	return &billingEngineUsecase{
		repo: repo,
	}
}

func (uc *billingEngineUsecase) GetStatements(loanID int64, until time.Time, limit, offset int) ([]entity.LoanStatement, error) {
	statementsDB, err := uc.repo.GetStatements(loanID, until, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get statements: %w", err)
	}

	statements := make([]entity.LoanStatement, 0, len(statementsDB))
	for _, s := range statementsDB {
		statements = append(statements, entity.LoanStatement{
			LoanID:        s.LoanID,
			StatementDate: s.StatementDate.Format("2006-01-02"),
			Amount:        strconv.FormatInt(s.Amount, 10),
			Status:        entity.LoanStatementStatus[s.Status],
		})
	}

	return statements, nil
}
