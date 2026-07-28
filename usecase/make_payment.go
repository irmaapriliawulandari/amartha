package usecase

import (
	"amartha-test/entity"
	"fmt"
	"time"
)

// MakePayment pays off a loan's latest statement (installment + carry
// over), marks any of its still-overdue prior statements as paid late, and
// clears the loan's open delinquency records — all in one transaction,
// with the latest statement row locked for the duration to prevent a
// concurrent double payment. If the paid statement was the loan's last
// cycle, the loan's status is updated to closed.
func (uc *billingEngineUsecase) MakePayment(loanID int64, now time.Time) (entity.MakePaymentResponse, error) {
	tx, err := uc.repo.BeginTx()
	if err != nil {
		return entity.MakePaymentResponse{}, fmt.Errorf("make payment: %w", err)
	}

	statement, err := uc.repo.GetLatestStatementForUpdate(tx, loanID, now)
	if err != nil {
		return entity.MakePaymentResponse{}, rollback(tx, fmt.Errorf("make payment: %w", err))
	}

	if statement.Status == entity.StatementStatusPaid || statement.Status == entity.StatementStatusPaidLate {
		return entity.MakePaymentResponse{}, rollback(tx, entity.ErrStatementAlreadyPaid)
	}

	paidAmount := statement.InstallmentAmount.Add(statement.CarryOverAmount)
	if err := uc.repo.UpdateStatementPaid(tx, statement.StatementID, paidAmount, now); err != nil {
		return entity.MakePaymentResponse{}, rollback(tx, fmt.Errorf("make payment: %w", err))
	}

	_, hasNextStatement, err := uc.repo.GetNextStatementForUpdate(tx, loanID, statement.StatementDate)
	if err != nil {
		return entity.MakePaymentResponse{}, rollback(tx, fmt.Errorf("make payment: %w", err))
	}
	if !hasNextStatement {
		if err := uc.repo.UpdateLoanStatus(tx, loanID, entity.LoanStatusClosed, now); err != nil {
			return entity.MakePaymentResponse{}, rollback(tx, fmt.Errorf("make payment: %w", err))
		}
	}

	if err := uc.repo.MarkPriorOverdueAsPaidLate(tx, loanID, now); err != nil {
		return entity.MakePaymentResponse{}, rollback(tx, fmt.Errorf("make payment: %w", err))
	}

	if err := uc.repo.ClearDelinquency(tx, loanID, now); err != nil {
		return entity.MakePaymentResponse{}, rollback(tx, fmt.Errorf("make payment: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return entity.MakePaymentResponse{}, fmt.Errorf("make payment: commit transaction: %w", err)
	}

	outstanding, _, err := uc.repo.GetOutstandingAmount(loanID)
	if err != nil {
		return entity.MakePaymentResponse{}, fmt.Errorf("get outstanding amount: %w", err)
	}

	return entity.MakePaymentResponse{
		LoanID:            loanID,
		StatementID:       statement.StatementID,
		PaidAmount:        paidAmount,
		PaidAt:            now.Format("2006-01-02 15:04:05"),
		OutstandingAmount: outstanding,
	}, nil
}
