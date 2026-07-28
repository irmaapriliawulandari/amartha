package usecase

import (
	"amartha-test/entity"
	"errors"
	"fmt"
	"time"
)

// MarkOverdue finds every still-unpaid statement of an active loan whose
// deadline was yesterday and marks it overdue. For each one it also rolls
// the missed amount (installment + carry over) into the next cycle's
// carry_over_amount, and — if the cycle before it was also overdue,
// meaning two consecutive misses — records a new delinquency_hist entry.
// Each statement is processed in its own transaction so one failure
// doesn't block the rest of the batch; it returns how many were
// successfully marked, even alongside a wrapped error for any that failed.
func (uc *billingEngineUsecase) MarkOverdue(now time.Time) (int, error) {
	yesterday := now.AddDate(0, 0, -1)

	candidates, err := uc.repo.ListOverdueCandidates(yesterday)
	if err != nil {
		return 0, fmt.Errorf("mark overdue: %w", err)
	}

	var marked int
	var errs []error
	for _, c := range candidates {
		didMark, err := uc.markStatementOverdue(c, now)
		if err != nil {
			errs = append(errs, fmt.Errorf("statement %d: %w", c.StatementID, err))
			continue
		}
		if didMark {
			marked++
		}
	}

	if len(errs) > 0 {
		return marked, fmt.Errorf("mark overdue: %w", errors.Join(errs...))
	}

	return marked, nil
}

func (uc *billingEngineUsecase) markStatementOverdue(c entity.OverdueCandidate, now time.Time) (bool, error) {
	tx, err := uc.repo.BeginTx()
	if err != nil {
		return false, err
	}

	statement, err := uc.repo.GetStatementForUpdate(tx, c.StatementID)
	if err != nil {
		return false, rollback(tx, err)
	}

	if statement.Status != entity.StatementStatusUnpaid {
		// already resolved (paid, or already marked overdue by a previous run) since it was listed
		if err := tx.Rollback(); err != nil {
			return false, fmt.Errorf("rollback: %w", err)
		}
		return false, nil
	}

	if err := uc.repo.MarkStatementOverdue(tx, c.StatementID, now); err != nil {
		return false, rollback(tx, err)
	}

	nextStatementID, found, err := uc.repo.GetNextStatementForUpdate(tx, c.LoanID, statement.StatementDate)
	if err != nil {
		return false, rollback(tx, err)
	}
	if found {
		newCarryOver := statement.InstallmentAmount.Add(statement.CarryOverAmount)
		if err := uc.repo.UpdateCarryOver(tx, nextStatementID, newCarryOver, now); err != nil {
			return false, rollback(tx, err)
		}
	}

	previousStatus, found, err := uc.repo.GetPreviousStatementStatus(tx, c.LoanID, statement.StatementDate)
	if err != nil {
		return false, rollback(tx, err)
	}
	if found && previousStatus == entity.StatementStatusOverdue {
		if err := uc.repo.InsertDelinquencyHistTx(tx, entity.DelinquencyHistDB{
			BorrowerID:  c.BorrowerID,
			LoanID:      c.LoanID,
			StatementID: c.StatementID,
		}); err != nil {
			return false, rollback(tx, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit transaction: %w", err)
	}

	return true, nil
}
