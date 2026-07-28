package usecase

import (
	"errors"
	"testing"

	"amartha-test/entity"

	"github.com/stretchr/testify/assert"
)

func TestBillingEngineUsecase_IsDelinquent(t *testing.T) {
	borrowerID := int64(1)

	t.Run("success returns delinquent status", func(t *testing.T) {
		repo := new(mockRepo)
		repo.On("IsDelinquent", borrowerID).Return(true, nil)

		uc := NewBillingEngineUsecase(repo)
		got, err := uc.IsDelinquent(borrowerID)

		assert.NoError(t, err)
		assert.Equal(t, entity.IsDelinquentResponse{BorrowerID: borrowerID, IsDelinquent: true}, got)
	})

	t.Run("repo error is wrapped", func(t *testing.T) {
		repo := new(mockRepo)
		repoErr := errors.New("connection lost")
		repo.On("IsDelinquent", borrowerID).Return(false, repoErr)

		uc := NewBillingEngineUsecase(repo)
		_, err := uc.IsDelinquent(borrowerID)

		assert.ErrorIs(t, err, repoErr)
	})
}

func TestBillingEngineUsecase_IsEverDelinquent(t *testing.T) {
	borrowerID := int64(1)

	t.Run("success returns ever-delinquent status", func(t *testing.T) {
		repo := new(mockRepo)
		repo.On("IsEverDelinquent", borrowerID).Return(true, nil)

		uc := NewBillingEngineUsecase(repo)
		got, err := uc.IsEverDelinquent(borrowerID)

		assert.NoError(t, err)
		assert.Equal(t, entity.IsEverDelinquentResponse{BorrowerID: borrowerID, IsEverDelinquent: true}, got)
	})

	t.Run("repo error is wrapped", func(t *testing.T) {
		repo := new(mockRepo)
		repoErr := errors.New("connection lost")
		repo.On("IsEverDelinquent", borrowerID).Return(false, repoErr)

		uc := NewBillingEngineUsecase(repo)
		_, err := uc.IsEverDelinquent(borrowerID)

		assert.ErrorIs(t, err, repoErr)
	})
}
