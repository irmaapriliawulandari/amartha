package usecase

import (
	"errors"
	"testing"
	"time"

	"amartha-test/entity"

	"github.com/stretchr/testify/assert"
)

func TestBillingEngineUsecase_PublishStatement(t *testing.T) {
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)

	t.Run("no candidates returns zero published, no error", func(t *testing.T) {
		repo := new(mockRepo)
		repo.On("ListPublishableCandidates", now).Return([]entity.PublishCandidate{}, nil)

		uc := NewBillingEngineUsecase(repo)
		published, err := uc.PublishStatement(now)

		assert.NoError(t, err)
		assert.Equal(t, 0, published)
	})

	t.Run("candidate listing error is wrapped", func(t *testing.T) {
		repo := new(mockRepo)
		repoErr := errors.New("connection lost")
		repo.On("ListPublishableCandidates", now).Return(nil, repoErr)

		uc := NewBillingEngineUsecase(repo)
		_, err := uc.PublishStatement(now)

		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("publishes statement due today", func(t *testing.T) {
		c := entity.PublishCandidate{StatementID: 10, LoanID: 1}

		repo := new(mockRepo)
		tx := new(mockTx)

		repo.On("ListPublishableCandidates", now).Return([]entity.PublishCandidate{c}, nil)
		repo.On("BeginTx").Return(tx, nil)
		repo.On("GetStatementForUpdate", tx, c.StatementID).Return(entity.StatementDB{
			StatementID: 10, LoanID: 1, Status: entity.StatementStatusCreated,
		}, nil)
		repo.On("MarkStatementPublished", tx, c.StatementID, now).Return(nil)
		tx.On("Commit").Return(nil)

		uc := NewBillingEngineUsecase(repo)
		published, err := uc.PublishStatement(now)

		assert.NoError(t, err)
		assert.Equal(t, 1, published)
		repo.AssertExpectations(t)
		tx.AssertExpectations(t)
	})

	t.Run("already resolved candidate is skipped without error", func(t *testing.T) {
		c := entity.PublishCandidate{StatementID: 10, LoanID: 1}

		repo := new(mockRepo)
		tx := new(mockTx)

		repo.On("ListPublishableCandidates", now).Return([]entity.PublishCandidate{c}, nil)
		repo.On("BeginTx").Return(tx, nil)
		repo.On("GetStatementForUpdate", tx, c.StatementID).Return(entity.StatementDB{
			StatementID: 10, LoanID: 1, Status: entity.StatementStatusPublished,
		}, nil)
		tx.On("Rollback").Return(nil)

		uc := NewBillingEngineUsecase(repo)
		published, err := uc.PublishStatement(now)

		assert.NoError(t, err)
		assert.Equal(t, 0, published)
		tx.AssertExpectations(t)
		tx.AssertNotCalled(t, "Commit")
	})

	t.Run("a failing candidate is aggregated but doesn't block the rest of the batch", func(t *testing.T) {
		c1 := entity.PublishCandidate{StatementID: 10, LoanID: 1}
		c2 := entity.PublishCandidate{StatementID: 20, LoanID: 3}

		repo := new(mockRepo)
		tx1 := new(mockTx)
		tx2 := new(mockTx)

		repo.On("ListPublishableCandidates", now).Return([]entity.PublishCandidate{c1, c2}, nil)

		repo.On("BeginTx").Return(tx1, nil).Once()
		repoErr := errors.New("connection refused")
		repo.On("GetStatementForUpdate", tx1, c1.StatementID).Return(entity.StatementDB{}, repoErr)
		tx1.On("Rollback").Return(nil)

		repo.On("BeginTx").Return(tx2, nil).Once()
		repo.On("GetStatementForUpdate", tx2, c2.StatementID).Return(entity.StatementDB{
			StatementID: 20, LoanID: 3, Status: entity.StatementStatusCreated,
		}, nil)
		repo.On("MarkStatementPublished", tx2, c2.StatementID, now).Return(nil)
		tx2.On("Commit").Return(nil)

		uc := NewBillingEngineUsecase(repo)
		published, err := uc.PublishStatement(now)

		assert.ErrorContains(t, err, "connection refused")
		assert.Equal(t, 1, published)
		repo.AssertExpectations(t)
		tx1.AssertExpectations(t)
		tx2.AssertExpectations(t)
	})
}
