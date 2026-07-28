package usecase

import (
	"amartha-test/entity"
	"errors"
	"fmt"
	"time"
)

// PublishStatement makes every statement due today (statement_date == now)
// visible and payable by moving it from Created to Published. Each
// statement is processed in its own transaction so one failure doesn't
// block the rest of the batch; it returns how many were successfully
// published, even alongside a wrapped error for any that failed.
func (uc *billingEngineUsecase) PublishStatement(now time.Time) (int, error) {
	candidates, err := uc.repo.ListPublishableCandidates(now)
	if err != nil {
		return 0, fmt.Errorf("publish statement: %w", err)
	}

	var published int
	var errs []error
	for _, c := range candidates {
		didPublish, err := uc.publishStatement(c, now)
		if err != nil {
			errs = append(errs, fmt.Errorf("statement %d: %w", c.StatementID, err))
			continue
		}
		if didPublish {
			published++
		}
	}

	if len(errs) > 0 {
		return published, fmt.Errorf("publish statement: %w", errors.Join(errs...))
	}

	return published, nil
}

func (uc *billingEngineUsecase) publishStatement(c entity.PublishCandidate, now time.Time) (bool, error) {
	tx, err := uc.repo.BeginTx()
	if err != nil {
		return false, err
	}

	statement, err := uc.repo.GetStatementForUpdate(tx, c.StatementID)
	if err != nil {
		return false, rollback(tx, err)
	}

	if statement.Status != entity.StatementStatusCreated {
		// already published since it was listed
		if err := tx.Rollback(); err != nil {
			return false, fmt.Errorf("rollback: %w", err)
		}
		return false, nil
	}

	if err := uc.repo.MarkStatementPublished(tx, c.StatementID, now); err != nil {
		return false, rollback(tx, err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit transaction: %w", err)
	}

	return true, nil
}
