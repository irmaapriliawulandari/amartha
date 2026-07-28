package handler

import (
	"amartha-test/entity"
	"encoding/json"
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
}

type getStatementsRequest struct {
	LoanID int64 `json:"loan_id"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
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
