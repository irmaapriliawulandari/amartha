package usecase

import (
	"amartha-test/entity"
	"fmt"

	"github.com/shopspring/decimal"
)

func (uc *billingEngineUsecase) InitBilling(loan entity.DisburseLoanMsg) error {
	loanDB, err := loan.ConvertToDB()
	if err != nil {
		return fmt.Errorf("convert disburse loan message: %w", err)
	}

	statements, err := buildStatements(loanDB)
	if err != nil {
		return fmt.Errorf("build statements: %w", err)
	}

	if err := uc.repo.InsertLoanWithStatements(loanDB, statements); err != nil {
		return fmt.Errorf("insert loan with statements: %w", err)
	}

	return nil
}

func buildStatements(loanDB entity.LoanDB) ([]entity.StatementDB, error) {
	count := int64(loanDB.InstallmentCount)
	baseAmount := loanDB.TotalAmount.Div(decimal.NewFromInt(count)).Round(2)
	lastAmount := loanDB.TotalAmount.Sub(baseAmount.Mul(decimal.NewFromInt(count - 1)))

	statements := make([]entity.StatementDB, loanDB.InstallmentCount)
	for i := 0; i < loanDB.InstallmentCount; i++ {
		statementDate, err := entity.NextStatementDate(loanDB.StartDate, loanDB.InstallmentType, i)
		if err != nil {
			return nil, err
		}

		deadline, err := entity.StatementDeadline(statementDate, loanDB.InstallmentType)
		if err != nil {
			return nil, err
		}

		amount := baseAmount
		if i == loanDB.InstallmentCount-1 {
			amount = lastAmount
		}

		statements[i] = entity.StatementDB{
			LoanID:            loanDB.LoanID,
			InstallmentAmount: amount,
			StatementDate:     statementDate,
			Deadline:          deadline,
			Status:            entity.StatementStatusCreated,
		}
	}

	return statements, nil
}
