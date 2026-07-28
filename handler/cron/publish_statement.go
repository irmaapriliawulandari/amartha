package cron

import (
	"log"
	"time"
)

// PublishStatement makes every statement due today visible and payable,
// and logs the outcome. It's the function scheduled by the cron entrypoint.
func (h *cronHandler) PublishStatement() {
	// now := time.Now()
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.Now().Location())

	published, err := h.uc.PublishStatement(now)
	if err != nil {
		log.Printf("[PublishStatement] completed with errors, published=%d, err: %s", published, err.Error())
		return
	}
	log.Printf("[PublishStatement] published %d statement(s)", published)
}
