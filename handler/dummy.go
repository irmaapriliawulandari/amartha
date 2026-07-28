package handler

import (
	"io"
	"log"
	"net/http"
)

func (h *httpHandler) Publish(w http.ResponseWriter, r *http.Request) {
	msg, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = h.publisher.Publish("disburse_loan", msg)
	if err != nil {
		http.Error(w, "failed to publish", http.StatusInternalServerError)
		log.Printf("[httpHandler][Publish] failed to publish msg: %s to %s, err: %s", string(msg), "disburse_loan", err.Error())
		return
	}

	writeResponse(w, "ok")
}
