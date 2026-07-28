package usecase

import (
	"amartha-test/entity"
	"time"

	"github.com/shopspring/decimal"
)

type repo interface {
	GetStatements(loanID int64, until time.Time, limit, offset int) ([]entity.StatementDB, error)
	InsertLoan(loan entity.LoanDB) error
	InsertStatement(statement entity.StatementDB) (int64, error)
	InsertDelinquencyHist(dh entity.DelinquencyHistDB) (int64, error)
	InsertLoanWithStatements(loan entity.LoanDB, statements []entity.StatementDB) error
	GetOutstandingAmount(loanID int64) (outstanding decimal.Decimal, borrowerID int64, err error)
	GetLatestStatement(loanID int64, before time.Time) (entity.StatementDB, error)
	IsDelinquent(borrowerID int64) (bool, error)
	IsEverDelinquent(borrowerID int64) (bool, error)
	IsLoanDelinquent(borrowerID, loanID int64) (bool, error)
	MakePayment(loanID int64, now time.Time) (entity.StatementDB, error)
}

type billingEngineUsecase struct {
	repo repo
}

func NewBillingEngineUsecase(repo repo) *billingEngineUsecase {
	return &billingEngineUsecase{
		repo: repo,
	}
}
