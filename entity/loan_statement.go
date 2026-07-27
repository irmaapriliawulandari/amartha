package entity

import "time"

var (
	LoanStatementStatus = map[int]string{
		0: "Unpaid",
		1: "Paid",
		2: "Overdue",
	}
)

type LoanStatement struct {
	LoanID        int64  `json:"loan_id"`
	StatementDate string `json:"statement_date"`
	Amount        string `json:"amount"`
	Status        string `json:"status"`
}

type LoanStatementDB struct {
	LoanID        int64     `json:"loan_id"`
	StatementDate time.Time `json:"statement_date"`
	Amount        int64     `json:"amount"`
	Status        int       `json:"status"`
}
