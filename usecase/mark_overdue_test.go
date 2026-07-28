package usecase

import (
	"errors"
	"testing"
	"time"

	"amartha-test/entity"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBillingEngineUsecase_MarkOverdue(t *testing.T) {
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	yesterday := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)

	t.Run("no candidates returns zero marked, no error", func(t *testing.T) {
		repo := new(mockRepo)
		repo.On("ListOverdueCandidates", yesterday).Return([]entity.OverdueCandidate{}, nil)

		uc := NewBillingEngineUsecase(repo)
		marked, err := uc.MarkOverdue(now)

		assert.NoError(t, err)
		assert.Equal(t, 0, marked)
	})

	t.Run("candidate listing error is wrapped", func(t *testing.T) {
		repo := new(mockRepo)
		repoErr := errors.New("connection lost")
		repo.On("ListOverdueCandidates", yesterday).Return(nil, repoErr)

		uc := NewBillingEngineUsecase(repo)
		_, err := uc.MarkOverdue(now)

		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("marks statement overdue, rolls carry over into next cycle, no delinquency when previous cycle wasn't overdue", func(t *testing.T) {
		c := entity.OverdueCandidate{StatementID: 10, LoanID: 1, BorrowerID: 2}
		statementDate := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)

		repo := new(mockRepo)
		tx := new(mockTx)

		repo.On("ListOverdueCandidates", yesterday).Return([]entity.OverdueCandidate{c}, nil)
		repo.On("BeginTx").Return(tx, nil)
		repo.On("GetStatementForUpdate", tx, c.StatementID).Return(entity.StatementDB{
			StatementID: 10, LoanID: 1,
			InstallmentAmount: decimal.NewFromInt(110000), CarryOverAmount: decimal.NewFromInt(0),
			StatementDate: statementDate, Status: entity.StatementStatusPublished,
		}, nil)
		repo.On("MarkStatementOverdue", tx, c.StatementID, now).Return(nil)
		repo.On("GetNextStatementForUpdate", tx, c.LoanID, statementDate).Return(int64(11), true, nil)
		repo.On("UpdateCarryOver", tx, int64(11), decimal.NewFromInt(110000), now).Return(nil)
		repo.On("GetPreviousStatementStatus", tx, c.LoanID, statementDate).Return(entity.StatementStatusPaid, true, nil)
		tx.On("Commit").Return(nil)

		uc := NewBillingEngineUsecase(repo)
		marked, err := uc.MarkOverdue(now)

		assert.NoError(t, err)
		assert.Equal(t, 1, marked)
		repo.AssertExpectations(t)
		tx.AssertExpectations(t)
		repo.AssertNotCalled(t, "InsertDelinquencyHistTx", mock.Anything, mock.Anything)
	})

	t.Run("records delinquency when previous cycle was also overdue", func(t *testing.T) {
		c := entity.OverdueCandidate{StatementID: 10, LoanID: 1, BorrowerID: 2}
		statementDate := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)

		repo := new(mockRepo)
		tx := new(mockTx)

		repo.On("ListOverdueCandidates", yesterday).Return([]entity.OverdueCandidate{c}, nil)
		repo.On("BeginTx").Return(tx, nil)
		repo.On("GetStatementForUpdate", tx, c.StatementID).Return(entity.StatementDB{
			StatementID: 10, LoanID: 1,
			InstallmentAmount: decimal.NewFromInt(50000), CarryOverAmount: decimal.NewFromInt(0),
			StatementDate: statementDate, Status: entity.StatementStatusPublished,
		}, nil)
		repo.On("MarkStatementOverdue", tx, c.StatementID, now).Return(nil)
		repo.On("GetNextStatementForUpdate", tx, c.LoanID, statementDate).Return(int64(0), false, nil)
		repo.On("GetPreviousStatementStatus", tx, c.LoanID, statementDate).Return(entity.StatementStatusOverdue, true, nil)
		repo.On("InsertDelinquencyHistTx", tx, entity.DelinquencyHistDB{
			BorrowerID: c.BorrowerID, LoanID: c.LoanID, StatementID: c.StatementID,
		}).Return(nil)
		tx.On("Commit").Return(nil)

		uc := NewBillingEngineUsecase(repo)
		marked, err := uc.MarkOverdue(now)

		assert.NoError(t, err)
		assert.Equal(t, 1, marked)
		repo.AssertExpectations(t)
		tx.AssertExpectations(t)
		repo.AssertNotCalled(t, "UpdateCarryOver", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("already resolved candidate is skipped without error", func(t *testing.T) {
		c := entity.OverdueCandidate{StatementID: 10, LoanID: 1, BorrowerID: 2}

		repo := new(mockRepo)
		tx := new(mockTx)

		repo.On("ListOverdueCandidates", yesterday).Return([]entity.OverdueCandidate{c}, nil)
		repo.On("BeginTx").Return(tx, nil)
		repo.On("GetStatementForUpdate", tx, c.StatementID).Return(entity.StatementDB{
			StatementID: 10, LoanID: 1, Status: entity.StatementStatusPaid,
		}, nil)
		tx.On("Rollback").Return(nil)

		uc := NewBillingEngineUsecase(repo)
		marked, err := uc.MarkOverdue(now)

		assert.NoError(t, err)
		assert.Equal(t, 0, marked)
		tx.AssertExpectations(t)
		tx.AssertNotCalled(t, "Commit")
	})

	t.Run("a failing candidate is aggregated but doesn't block the rest of the batch", func(t *testing.T) {
		c1 := entity.OverdueCandidate{StatementID: 10, LoanID: 1, BorrowerID: 2}
		c2 := entity.OverdueCandidate{StatementID: 20, LoanID: 3, BorrowerID: 4}
		statementDate := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)

		repo := new(mockRepo)
		tx1 := new(mockTx)
		tx2 := new(mockTx)

		repo.On("ListOverdueCandidates", yesterday).Return([]entity.OverdueCandidate{c1, c2}, nil)

		repo.On("BeginTx").Return(tx1, nil).Once()
		repoErr := errors.New("connection refused")
		repo.On("GetStatementForUpdate", tx1, c1.StatementID).Return(entity.StatementDB{}, repoErr)
		tx1.On("Rollback").Return(nil)

		repo.On("BeginTx").Return(tx2, nil).Once()
		repo.On("GetStatementForUpdate", tx2, c2.StatementID).Return(entity.StatementDB{
			StatementID: 20, LoanID: 3,
			InstallmentAmount: decimal.NewFromInt(50000), CarryOverAmount: decimal.NewFromInt(0),
			StatementDate: statementDate, Status: entity.StatementStatusPublished,
		}, nil)
		repo.On("MarkStatementOverdue", tx2, c2.StatementID, now).Return(nil)
		repo.On("GetNextStatementForUpdate", tx2, c2.LoanID, statementDate).Return(int64(0), false, nil)
		repo.On("GetPreviousStatementStatus", tx2, c2.LoanID, statementDate).Return(0, false, nil)
		tx2.On("Commit").Return(nil)

		uc := NewBillingEngineUsecase(repo)
		marked, err := uc.MarkOverdue(now)

		assert.ErrorContains(t, err, "connection refused")
		assert.Equal(t, 1, marked)
		repo.AssertExpectations(t)
		tx1.AssertExpectations(t)
		tx2.AssertExpectations(t)
	})
}
