package usecase

import (
	txrepo "amartha-test/repo"
	"fmt"
)

// rollback rolls back tx and returns cause, or a wrapped error noting the
// rollback itself also failed.
func rollback(tx txrepo.Tx, cause error) error {
	if err := tx.Rollback(); err != nil {
		return fmt.Errorf("%w (rollback failed: %v)", cause, err)
	}

	return cause
}
