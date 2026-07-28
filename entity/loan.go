package entity

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

const (
	DefaultCurrency = "IDR"

	InstallmentTypeDaily   = 1
	InstallmentTypeWeekly  = 2
	InstallmentTypeMonthly = 3

	LoanStatusActive = 1
	LoanStatusClosed = 2
)

var ErrLoanNotFound = errors.New("loan not found")

// OutstandingAmount is loan.total_amount minus everything paid so far across
// all of its statements.
type OutstandingAmount struct {
	LoanID            int64           `json:"loan_id"`
	OutstandingAmount decimal.Decimal `json:"outstanding_amount"`
}

type DisburseLoanMsg struct {
	LoanID           int64           `json:"loan_id"`
	BorrowerID       int64           `json:"borrower_id"`
	PrincipalAmount  decimal.Decimal `json:"principal_amount"`
	Rate             decimal.Decimal `json:"rate"`
	TotalAmount      decimal.Decimal `json:"total_amount"`
	InstallmentCount int             `json:"installment_count"`
	InstallmentType  int             `json:"installment_type"`
	Currency         string          `json:"currency"`
	DisbursedAt      time.Time       `json:"disbursed_at"`
}

func (dlm *DisburseLoanMsg) Validate() bool {
	return dlm.LoanID > 0 &&
		dlm.BorrowerID > 0 &&
		dlm.PrincipalAmount.IntPart() > 0 &&
		dlm.TotalAmount.IntPart() > 0 &&
		dlm.InstallmentCount > 0 &&
		!dlm.DisbursedAt.IsZero()
}

func (dlm *DisburseLoanMsg) ConvertToDB() (LoanDB, error) {
	currency := DefaultCurrency
	if dlm.Currency != "" {
		currency = dlm.Currency
	}

	startDate, err := getBillingStartDate(dlm.DisbursedAt, dlm.InstallmentType)
	if err != nil {
		return LoanDB{}, err
	}

	return LoanDB{
		LoanID:           dlm.LoanID,
		BorrowerID:       dlm.BorrowerID,
		PrincipalAmount:  dlm.PrincipalAmount,
		Rate:             dlm.Rate,
		TotalAmount:      dlm.TotalAmount,
		InstallmentCount: dlm.InstallmentCount,
		InstallmentType:  dlm.InstallmentType,
		Currency:         currency,
		StartDate:        startDate,
		DisbursedAt:      sql.NullTime{Time: dlm.DisbursedAt, Valid: !dlm.DisbursedAt.IsZero()},
		Status:           LoanStatusActive,
		UpdatedAt:        sql.NullTime{Time: time.Now(), Valid: true},
	}, nil
}

// getBillingStartDate returns the date (midnight, WIB) the first billing
// cycle starts on, one period after disbursement.
func getBillingStartDate(disburseAt time.Time, installmentType int) (time.Time, error) {
	next, err := NextStatementDate(disburseAt.In(DefaultLocation), installmentType, 1)
	if err != nil {
		return time.Time{}, err
	}

	return truncateToDate(next), nil
}

func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, DefaultLocation)
}

// NextStatementDate returns the date n billing periods after start.
func NextStatementDate(start time.Time, installmentType, n int) (time.Time, error) {
	switch installmentType {
	case InstallmentTypeDaily:
		return start.AddDate(0, 0, n), nil
	case InstallmentTypeWeekly:
		return start.AddDate(0, 0, 7*n), nil
	case InstallmentTypeMonthly:
		return start.AddDate(0, n, 0), nil
	default:
		return time.Time{}, fmt.Errorf("invalid installment type")
	}
}

// StatementDeadline returns the payment deadline for a statement dated
// statementDate: one day before the next billing period's statement date.
func StatementDeadline(statementDate time.Time, installmentType int) (time.Time, error) {
	next, err := NextStatementDate(statementDate, installmentType, 1)
	if err != nil {
		return time.Time{}, err
	}

	return next.AddDate(0, 0, -1), nil
}

// LoanDB mirrors every column of the loan table.
type LoanDB struct {
	LoanID           int64           `json:"loan_id"`
	BorrowerID       int64           `json:"borrower_id"`
	PrincipalAmount  decimal.Decimal `json:"principal_amount"`
	Rate             decimal.Decimal `json:"rate"`
	TotalAmount      decimal.Decimal `json:"total_amount"`
	InstallmentCount int             `json:"installment_count"`
	InstallmentType  int             `json:"installment_type"`
	Currency         string          `json:"currency"`
	StartDate        time.Time       `json:"start_date"`
	DisbursedAt      sql.NullTime    `json:"disbursed_at"`
	Metadata         []byte          `json:"metadata,omitempty"`
	Status           int             `json:"status"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        sql.NullTime    `json:"updated_at"`
}
