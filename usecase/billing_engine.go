package usecase

import (
	"amartha-test/entity"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

func (uc *billingEngineUsecase) GetStatements(loanID int64, until time.Time, limit, offset int) ([]entity.Statement, error) {
	statementsDB, err := uc.repo.GetStatements(loanID, until, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get statements: %w", err)
	}

	statements := make([]entity.Statement, 0, len(statementsDB))
	for _, s := range statementsDB {
		var toPay decimal.Decimal
		if s.Status != entity.StatementStatusPaid {
			toPay = s.InstallmentAmount.Add(s.CarryOverAmount)
		}
		statements = append(statements, entity.Statement{
			LoanID:        s.LoanID,
			StatementDate: s.StatementDate.Format("2006-01-02"),
			ToPayAmount:   toPay,
			Status:        entity.StatementStatus[s.Status],
		})
	}

	return statements, nil
}

func (uc *billingEngineUsecase) GetOutstandingAmount(loanID int64) (entity.OutstandingAmount, error) {
	outstanding, _, err := uc.repo.GetOutstandingAmount(loanID)
	if err != nil {
		return entity.OutstandingAmount{}, fmt.Errorf("get outstanding amount: %w", err)
	}

	return entity.OutstandingAmount{LoanID: loanID, OutstandingAmount: outstanding}, nil
}

func (uc *billingEngineUsecase) GetLatestStatement(loanID int64, now time.Time) (entity.LatestStatement, error) {
	statement, err := uc.repo.GetLatestStatement(loanID, now)
	if err != nil {
		return entity.LatestStatement{}, fmt.Errorf("get latest statement: %w", err)
	}

	outstanding, borrowerID, err := uc.repo.GetOutstandingAmount(loanID)
	if err != nil {
		return entity.LatestStatement{}, fmt.Errorf("get outstanding amount: %w", err)
	}

	isDelinquent, err := uc.repo.IsLoanDelinquent(borrowerID, loanID)
	if err != nil {
		return entity.LatestStatement{}, fmt.Errorf("check is delinquent: %w", err)
	}

	var toPay decimal.Decimal
	if statement.Status != entity.StatementStatusPaid {
		toPay = statement.InstallmentAmount.Add(statement.CarryOverAmount)
	}

	return entity.LatestStatement{
		LoanID:            loanID,
		StatementDate:     statement.StatementDate.Format("2006-01-02"),
		CarryOverAmount:   statement.CarryOverAmount,
		InstallmentAmount: statement.InstallmentAmount,
		TotalToPay:        toPay,
		Status:            entity.StatementStatus[statement.Status],
		Deadline:          statement.Deadline.Format("2006-01-02"),
		OutstandingAmount: outstanding,
		IsDelinquent:      isDelinquent,
	}, nil
}
