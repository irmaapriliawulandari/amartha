package handler

import (
	"amartha-test/entity"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

const (
	defaultLimit  = 10
	defaultOffset = 0
)

type usecases interface {
	GetStatements(loanID int64, until time.Time, limit, offset int) ([]entity.Statement, error)
	GetOutstandingAmount(loanID int64) (entity.OutstandingAmount, error)
	GetLatestStatement(loanID int64, now time.Time) (entity.LatestStatement, error)
}

type getStatementsRequest struct {
	LoanID int64 `json:"loan_id"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

type getOutstandingAmountRequest struct {
	LoanID int64 `json:"loan_id"`
}

type getLatestStatementRequest struct {
	LoanID int64 `json:"loan_id"`
}

// GetStatement returns all previous and current statements
func (h *httpHandler) GetStatements(w http.ResponseWriter, r *http.Request) {
	var req getStatementsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.LoanID <= 0 {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// now := time.Now()
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.Now().Location())
	statetementList, err := h.uc.GetStatements(req.LoanID, now, req.Limit, req.Offset)
	if err != nil {
		http.Error(w, "failed to get data", http.StatusInternalServerError)
		log.Printf("[GetStatements] failed to get data, err: %s", err.Error())
		return
	}

	writeResponse(w, statetementList)
}

// GetOutstandingAmount returns loan.total_amount minus everything paid so far.
func (h *httpHandler) GetOutstandingAmount(w http.ResponseWriter, r *http.Request) {
	var req getOutstandingAmountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.LoanID <= 0 {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	outstanding, err := h.uc.GetOutstandingAmount(req.LoanID)
	if err != nil {
		if errors.Is(err, entity.ErrLoanNotFound) {
			http.Error(w, "loan not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get data", http.StatusInternalServerError)
		log.Printf("[GetOutstandingAmount] failed to get data, err: %s", err.Error())
		return
	}

	writeResponse(w, outstanding)
}

// GetLatestStatement returns the currently-due statement for a loan: the
// most recent unpaid statement dated before today, plus the loan's overall
// outstanding balance.
func (h *httpHandler) GetLatestStatement(w http.ResponseWriter, r *http.Request) {
	var req getLatestStatementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.LoanID <= 0 {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// now := time.Now()
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.Now().Location())
	latest, err := h.uc.GetLatestStatement(req.LoanID, now)
	if err != nil {
		if errors.Is(err, entity.ErrStatementNotFound) || errors.Is(err, entity.ErrLoanNotFound) {
			http.Error(w, "no active statement found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get data", http.StatusInternalServerError)
		log.Printf("[GetLatestStatement] failed to get data, err: %s", err.Error())
		return
	}

	writeResponse(w, latest)
}
