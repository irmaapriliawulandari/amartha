package handler

import (
	auth "amartha-test/helper"
	"encoding/json"
	"net/http"
)

// RegisterRoutes wires the billing engine routes onto mux, protected by AuthMiddleware.
func RegisterRoutes(mux *http.ServeMux, h *billingEngine) {
	mux.HandleFunc("/ping", auth.AuthMiddleware(Ping))
	mux.HandleFunc("/billing-engine/get-statements", auth.AuthMiddleware(h.GetStatements))
}

// Ping check service healthy
func Ping(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, map[string]string{"status": "ok", "msg": "pong"})
}

func writeResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
