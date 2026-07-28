package handler

import (
	"amartha-test/entity"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

type makePaymentRequest struct {
	LoanID int64 `json:"loan_id"`
}

// MakePayment pays off a loan's latest statement, marks any still-overdue
// prior statements as paid late, and clears the loan's open delinquency records.
func (h *httpHandler) MakePayment(w http.ResponseWriter, r *http.Request) {
	var req makePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.LoanID <= 0 {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// now := time.Now()
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.Now().Location())
	res, err := h.uc.MakePayment(req.LoanID, now)
	if err != nil {
		if errors.Is(err, entity.ErrStatementNotFound) || errors.Is(err, entity.ErrLoanNotFound) {
			http.Error(w, "no active statement found", http.StatusNotFound)
			return
		}
		if errors.Is(err, entity.ErrStatementAlreadyPaid) {
			http.Error(w, "statement already paid", http.StatusConflict)
			return
		}
		http.Error(w, "failed to make payment", http.StatusInternalServerError)
		log.Printf("[MakePayment] failed to make payment, err: %s", err.Error())
		return
	}

	writeResponse(w, res)
}
