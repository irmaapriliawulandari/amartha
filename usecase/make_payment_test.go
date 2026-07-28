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

func TestBillingEngineUsecase_MakePayment(t *testing.T) {
	loanID := int64(1)
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)

	t.Run("success pays statement, marks overdue paid late, clears delinquency, commits", func(t *testing.T) {
		repo := new(mockRepo)
		tx := new(mockTx)

		repo.On("BeginTx").Return(tx, nil)
		repo.On("GetLatestStatementForUpdate", tx, loanID, now).Return(entity.StatementDB{
			StatementID:       5,
			InstallmentAmount: decimal.NewFromInt(110000),
			CarryOverAmount:   decimal.NewFromInt(10000),
			Status:            entity.StatementStatusUnpaid,
		}, nil)
		repo.On("UpdateStatementPaid", tx, int64(5), decimal.NewFromInt(120000), now).Return(nil)
		repo.On("MarkPriorOverdueAsPaidLate", tx, loanID, now).Return(nil)
		repo.On("ClearDelinquency", tx, loanID, now).Return(nil)
		tx.On("Commit").Return(nil)
		repo.On("GetOutstandingAmount", loanID).Return(decimal.NewFromInt(4380000), int64(2), nil)

		uc := NewBillingEngineUsecase(repo)
		got, err := uc.MakePayment(loanID, now)

		assert.NoError(t, err)
		assert.Equal(t, entity.MakePaymentResponse{
			LoanID:            loanID,
			StatementID:       5,
			PaidAmount:        decimal.NewFromInt(120000),
			PaidAt:            "2026-08-25 00:00:00",
			OutstandingAmount: decimal.NewFromInt(4380000),
		}, got)
		repo.AssertExpectations(t)
		tx.AssertExpectations(t)
		tx.AssertNotCalled(t, "Rollback")
	})

	t.Run("no statement to pay rolls back and is passed through", func(t *testing.T) {
		repo := new(mockRepo)
		tx := new(mockTx)

		repo.On("BeginTx").Return(tx, nil)
		repo.On("GetLatestStatementForUpdate", tx, loanID, now).Return(entity.StatementDB{}, entity.ErrStatementNotFound)
		tx.On("Rollback").Return(nil)

		uc := NewBillingEngineUsecase(repo)
		_, err := uc.MakePayment(loanID, now)

		assert.ErrorIs(t, err, entity.ErrStatementNotFound)
		tx.AssertExpectations(t)
		tx.AssertNotCalled(t, "Commit")
	})

	t.Run("already paid rolls back and is passed through", func(t *testing.T) {
		repo := new(mockRepo)
		tx := new(mockTx)

		repo.On("BeginTx").Return(tx, nil)
		repo.On("GetLatestStatementForUpdate", tx, loanID, now).Return(entity.StatementDB{
			StatementID: 5, Status: entity.StatementStatusPaid,
		}, nil)
		tx.On("Rollback").Return(nil)

		uc := NewBillingEngineUsecase(repo)
		_, err := uc.MakePayment(loanID, now)

		assert.ErrorIs(t, err, entity.ErrStatementAlreadyPaid)
		tx.AssertExpectations(t)
		repo.AssertNotCalled(t, "UpdateStatementPaid", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("update failure rolls back the whole transaction", func(t *testing.T) {
		repo := new(mockRepo)
		tx := new(mockTx)

		repo.On("BeginTx").Return(tx, nil)
		repo.On("GetLatestStatementForUpdate", tx, loanID, now).Return(entity.StatementDB{
			StatementID: 5, Status: entity.StatementStatusOverdue,
		}, nil)
		repoErr := errors.New("connection refused")
		repo.On("UpdateStatementPaid", tx, int64(5), mock.Anything, now).Return(repoErr)
		tx.On("Rollback").Return(nil)

		uc := NewBillingEngineUsecase(repo)
		_, err := uc.MakePayment(loanID, now)

		assert.ErrorIs(t, err, repoErr)
		tx.AssertExpectations(t)
		repo.AssertNotCalled(t, "MarkPriorOverdueAsPaidLate", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("outstanding amount repo error is wrapped after a successful commit", func(t *testing.T) {
		repo := new(mockRepo)
		tx := new(mockTx)

		repo.On("BeginTx").Return(tx, nil)
		repo.On("GetLatestStatementForUpdate", tx, loanID, now).Return(entity.StatementDB{
			StatementID: 5, Status: entity.StatementStatusUnpaid,
		}, nil)
		repo.On("UpdateStatementPaid", tx, int64(5), mock.Anything, now).Return(nil)
		repo.On("MarkPriorOverdueAsPaidLate", tx, loanID, now).Return(nil)
		repo.On("ClearDelinquency", tx, loanID, now).Return(nil)
		tx.On("Commit").Return(nil)
		repoErr := errors.New("connection lost")
		repo.On("GetOutstandingAmount", loanID).Return(decimal.Decimal{}, int64(0), repoErr)

		uc := NewBillingEngineUsecase(repo)
		_, err := uc.MakePayment(loanID, now)

		assert.ErrorIs(t, err, repoErr)
	})
}
