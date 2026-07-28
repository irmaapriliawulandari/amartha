package entity

import (
	"database/sql"
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

type Statement struct {
	LoanID        int64           `json:"loan_id"`
	StatementDate string          `json:"statement_date"`
	ToPayAmount   decimal.Decimal `json:"to_pay_amount"`
	Status        string          `json:"status"`
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
