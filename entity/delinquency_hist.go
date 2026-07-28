package entity

import (
	"database/sql"
	"time"
)

// DelinquencyHistDB mirrors every column of the delinquency_hist table.
type DelinquencyHistDB struct {
	DHID        int64        `json:"dh_id"`
	BorrowerID  int64        `json:"borrower_id"`
	LoanID      int64        `json:"loan_id"`
	StatementID int64        `json:"statement_id"`
	MarkedAt    time.Time    `json:"marked_at"`
	ClearedAt   sql.NullTime `json:"cleared_at"`
	UpdatedAt   sql.NullTime `json:"updated_at"`
}

// IsDelinquentResponse reports whether a borrower currently has an
// uncleared delinquency record.
type IsDelinquentResponse struct {
	BorrowerID   int64 `json:"borrower_id"`
	IsDelinquent bool  `json:"is_delinquent"`
}

// IsEverDelinquentResponse reports whether a borrower has ever had a
// delinquency record, cleared or not.
type IsEverDelinquentResponse struct {
	BorrowerID       int64 `json:"borrower_id"`
	IsEverDelinquent bool  `json:"is_ever_delinquent"`
}
