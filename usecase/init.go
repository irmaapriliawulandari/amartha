package usecase

import (
	"amartha-test/entity"
	"time"
)

type repo interface {
	GetStatements(loanID int64, until time.Time, limit, offset int) ([]entity.StatementDB, error)
	InsertLoan(loan entity.LoanDB) error
	InsertStatement(statement entity.StatementDB) (int64, error)
	InsertDelinquencyHist(dh entity.DelinquencyHistDB) (int64, error)
	InsertLoanWithStatements(loan entity.LoanDB, statements []entity.StatementDB) error
}

type billingEngineUsecase struct {
	repo repo
}

func NewBillingEngineUsecase(repo repo) *billingEngineUsecase {
	return &billingEngineUsecase{
		repo: repo,
	}
}
