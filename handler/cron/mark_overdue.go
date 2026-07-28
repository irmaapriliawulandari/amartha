package cron

import (
	"log"
	"time"
)

type usecases interface {
	MarkOverdue(now time.Time) (int, error)
}

type cronHandler struct {
	uc usecases
}

func NewCronHandler(uc usecases) *cronHandler {
	return &cronHandler{
		uc: uc,
	}
}

// Run marks every eligible statement overdue and logs the outcome. It's the
// function scheduled by the cron entrypoint.
func (h *cronHandler) MarkOverdue() {
	// now := time.Now()
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.Now().Location())

	marked, err := h.uc.MarkOverdue(now)
	if err != nil {
		log.Printf("[MarkOverdue] completed with errors, marked=%d, err: %s", marked, err.Error())
		return
	}
	log.Printf("[MarkOverdue] marked %d statement(s) overdue", marked)
}
