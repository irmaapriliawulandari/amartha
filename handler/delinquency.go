package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

type borrowerRequest struct {
	BorrowerID int64 `json:"borrower_id"`
}

// IsDelinquent reports whether the borrower currently has an uncleared delinquency record.
func (h *httpHandler) IsDelinquent(w http.ResponseWriter, r *http.Request) {
	var req borrowerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BorrowerID <= 0 {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.uc.IsDelinquent(req.BorrowerID)
	if err != nil {
		http.Error(w, "failed to get data", http.StatusInternalServerError)
		log.Printf("[IsDelinquent] failed to get data, err: %s", err.Error())
		return
	}

	writeResponse(w, res)
}

// IsEverDelinquent reports whether the borrower has ever had a delinquency record, cleared or not.
func (h *httpHandler) IsEverDelinquent(w http.ResponseWriter, r *http.Request) {
	var req borrowerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BorrowerID <= 0 {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.uc.IsEverDelinquent(req.BorrowerID)
	if err != nil {
		http.Error(w, "failed to get data", http.StatusInternalServerError)
		log.Printf("[IsEverDelinquent] failed to get data, err: %s", err.Error())
		return
	}

	writeResponse(w, res)
}
