package usecase

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"amartha-test/entity"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestBillingEngineUsecase_MakePayment(t *testing.T) {
	loanID := int64(1)
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)

	t.Run("success returns payment confirmation with updated outstanding amount", func(t *testing.T) {
		repo := new(mockRepo)
		repo.On("MakePayment", loanID, now).Return(entity.StatementDB{
			StatementID: 5,
			LoanID:      loanID,
			PaidAmount:  decimal.NewFromInt(120000),
			PaidAt:      sql.NullTime{Time: now, Valid: true},
			Status:      entity.StatementStatusPaid,
		}, nil)
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
	})

	t.Run("no statement to pay is passed through", func(t *testing.T) {
		repo := new(mockRepo)
		repo.On("MakePayment", loanID, now).Return(entity.StatementDB{}, entity.ErrStatementNotFound)

		uc := NewBillingEngineUsecase(repo)
		_, err := uc.MakePayment(loanID, now)

		assert.ErrorIs(t, err, entity.ErrStatementNotFound)
	})

	t.Run("already paid is passed through", func(t *testing.T) {
		repo := new(mockRepo)
		repo.On("MakePayment", loanID, now).Return(entity.StatementDB{}, entity.ErrStatementAlreadyPaid)

		uc := NewBillingEngineUsecase(repo)
		_, err := uc.MakePayment(loanID, now)

		assert.ErrorIs(t, err, entity.ErrStatementAlreadyPaid)
	})

	t.Run("repo error is wrapped", func(t *testing.T) {
		repo := new(mockRepo)
		repoErr := errors.New("connection lost")
		repo.On("MakePayment", loanID, now).Return(entity.StatementDB{}, repoErr)

		uc := NewBillingEngineUsecase(repo)
		_, err := uc.MakePayment(loanID, now)

		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("outstanding amount repo error is wrapped", func(t *testing.T) {
		repo := new(mockRepo)
		repo.On("MakePayment", loanID, now).Return(entity.StatementDB{
			StatementID: 5, LoanID: loanID, PaidAt: sql.NullTime{Time: now, Valid: true},
		}, nil)
		repoErr := errors.New("connection lost")
		repo.On("GetOutstandingAmount", loanID).Return(decimal.Decimal{}, int64(0), repoErr)

		uc := NewBillingEngineUsecase(repo)
		_, err := uc.MakePayment(loanID, now)

		assert.ErrorIs(t, err, repoErr)
	})
}
