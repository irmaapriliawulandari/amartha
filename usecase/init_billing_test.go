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

func validDisburseLoanMsg() entity.DisburseLoanMsg {
	return entity.DisburseLoanMsg{
		LoanID:           1,
		BorrowerID:       2,
		PrincipalAmount:  decimal.NewFromInt(5000000),
		Rate:             decimal.NewFromFloat(0.1),
		TotalAmount:      decimal.NewFromInt(5500000),
		InstallmentCount: 50,
		InstallmentType:  entity.InstallmentTypeWeekly,
		Currency:         entity.DefaultCurrency,
		DisbursedAt:      time.Date(2026, 7, 25, 13, 30, 0, 0, entity.DefaultLocation),
	}
}

func TestBillingEngineUsecase_InitBilling(t *testing.T) {
	t.Run("success inserts loan with its full billing schedule", func(t *testing.T) {
		msg := validDisburseLoanMsg()

		repo := new(mockRepo)
		repo.On("InsertLoanWithStatements",
			mock.MatchedBy(func(loan entity.LoanDB) bool {
				return loan.LoanID == msg.LoanID &&
					loan.BorrowerID == msg.BorrowerID &&
					loan.Status == entity.LoanStatusActive
			}),
			mock.MatchedBy(func(statements []entity.StatementDB) bool {
				return len(statements) == msg.InstallmentCount
			}),
		).Return(nil)

		uc := NewBillingEngineUsecase(repo)
		err := uc.InitBilling(msg)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("invalid installment type is not inserted", func(t *testing.T) {
		msg := validDisburseLoanMsg()
		msg.InstallmentType = 99

		repo := new(mockRepo)

		uc := NewBillingEngineUsecase(repo)
		err := uc.InitBilling(msg)

		assert.Error(t, err)
		repo.AssertNotCalled(t, "InsertLoanWithStatements", mock.Anything, mock.Anything)
	})

	t.Run("repo error is returned", func(t *testing.T) {
		msg := validDisburseLoanMsg()

		repo := new(mockRepo)
		repo.On("InsertLoanWithStatements", mock.Anything, mock.Anything).Return(errors.New("connection refused"))

		uc := NewBillingEngineUsecase(repo)
		err := uc.InitBilling(msg)

		assert.ErrorContains(t, err, "connection refused")
		repo.AssertExpectations(t)
	})
}
