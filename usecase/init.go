package usecase

import (
	"amartha-test/entity"
	txrepo "amartha-test/repo"
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

	BeginTx() (txrepo.Tx, error)
	GetLatestStatementForUpdate(tx txrepo.Tx, loanID int64, before time.Time) (entity.StatementDB, error)
	UpdateStatementPaid(tx txrepo.Tx, statementID int64, paidAmount decimal.Decimal, now time.Time) error
	MarkPriorOverdueAsPaidLate(tx txrepo.Tx, loanID int64, now time.Time) error
	ClearDelinquency(tx txrepo.Tx, loanID int64, now time.Time) error

	ListOverdueCandidates(deadline time.Time) ([]entity.OverdueCandidate, error)
	GetStatementForUpdate(tx txrepo.Tx, statementID int64) (entity.StatementDB, error)
	MarkStatementOverdue(tx txrepo.Tx, statementID int64, now time.Time) error
	GetNextStatementForUpdate(tx txrepo.Tx, loanID int64, afterDate time.Time) (statementID int64, found bool, err error)
	UpdateCarryOver(tx txrepo.Tx, statementID int64, carryOverAmount decimal.Decimal, now time.Time) error
	GetPreviousStatementStatus(tx txrepo.Tx, loanID int64, beforeDate time.Time) (status int, found bool, err error)
	InsertDelinquencyHistTx(tx txrepo.Tx, dh entity.DelinquencyHistDB) error
}

type billingEngineUsecase struct {
	repo repo
}

func NewBillingEngineUsecase(repo repo) *billingEngineUsecase {
	return &billingEngineUsecase{
		repo: repo,
	}
}
