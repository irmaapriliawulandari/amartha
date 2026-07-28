package usecase

import (
	"amartha-test/entity"
	"fmt"
	"time"
)

func (uc *billingEngineUsecase) MakePayment(loanID int64, now time.Time) (entity.MakePaymentResponse, error) {
	statement, err := uc.repo.MakePayment(loanID, now)
	if err != nil {
		return entity.MakePaymentResponse{}, fmt.Errorf("make payment: %w", err)
	}

	outstanding, _, err := uc.repo.GetOutstandingAmount(loanID)
	if err != nil {
		return entity.MakePaymentResponse{}, fmt.Errorf("get outstanding amount: %w", err)
	}

	return entity.MakePaymentResponse{
		LoanID:            loanID,
		StatementID:       statement.StatementID,
		PaidAmount:        statement.PaidAmount,
		PaidAt:            statement.PaidAt.Time.Format("2006-01-02 15:04:05"),
		OutstandingAmount: outstanding,
	}, nil
}
