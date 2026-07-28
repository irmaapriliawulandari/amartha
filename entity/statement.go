package entity

import (
	"database/sql"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

const (
	StatementStatusOverdue  = -1
	StatementStatusUnpaid   = 0
	StatementStatusPaid     = 1
	StatementStatusPaidLate = 2
)

var (
	StatementStatus = map[int]string{
		StatementStatusOverdue:  "Overdue",
		StatementStatusUnpaid:   "Unpaid",
		StatementStatusPaid:     "Paid",
		StatementStatusPaidLate: "Paid Late",
	}
)

var ErrStatementNotFound = errors.New("statement not found")
var ErrStatementAlreadyPaid = errors.New("statement already paid")

type Statement struct {
	LoanID        int64           `json:"loan_id"`
	StatementDate string          `json:"statement_date"`
	ToPayAmount   decimal.Decimal `json:"to_pay_amount"`
	Status        string          `json:"status"`
}

// LatestStatement is the currently-due statement for a loan, i.e. the most
// recent unpaid statement dated before the reference date, plus the loan's
// overall outstanding balance.
type LatestStatement struct {
	LoanID            int64           `json:"loan_id"`
	StatementDate     string          `json:"statement_date"`
	CarryOverAmount   decimal.Decimal `json:"carry_over_amount"`
	InstallmentAmount decimal.Decimal `json:"installment_amount"`
	TotalToPay        decimal.Decimal `json:"total_to_pay"`
	Status            string          `json:"status"`
	Deadline          string          `json:"deadline"`
	OutstandingAmount decimal.Decimal `json:"outstanding_amount"`
	IsDelinquent      bool            `json:"is_delinquent"`
}

// MakePaymentResponse confirms a payment applied to a loan's latest statement.
type MakePaymentResponse struct {
	LoanID            int64           `json:"loan_id"`
	StatementID       int64           `json:"statement_id"`
	PaidAmount        decimal.Decimal `json:"paid_amount"`
	PaidAt            string          `json:"paid_at"`
	OutstandingAmount decimal.Decimal `json:"outstanding_amount"`
}

// StatementDB mirrors every column of the statement table.
type StatementDB struct {
	StatementID       int64           `json:"statement_id"`
	LoanID            int64           `json:"loan_id"`
	InstallmentAmount decimal.Decimal `json:"installment_amount"`
	CarryOverAmount   decimal.Decimal `json:"carry_over_amount"`
	PaidAmount        decimal.Decimal `json:"paid_amount"`
	StatementDate     time.Time       `json:"statement_date"`
	Deadline          time.Time       `json:"deadline"`
	Status            int             `json:"status"`
	PaidAt            sql.NullTime    `json:"paid_at"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         sql.NullTime    `json:"updated_at"`
}
