package usecase

import (
	"amartha-test/entity"
	"fmt"
)

func (uc *billingEngineUsecase) IsDelinquent(borrowerID int64) (entity.IsDelinquentResponse, error) {
	isDelinquent, err := uc.repo.IsDelinquent(borrowerID)
	if err != nil {
		return entity.IsDelinquentResponse{}, fmt.Errorf("check is delinquent: %w", err)
	}

	return entity.IsDelinquentResponse{BorrowerID: borrowerID, IsDelinquent: isDelinquent}, nil
}

func (uc *billingEngineUsecase) IsEverDelinquent(borrowerID int64) (entity.IsEverDelinquentResponse, error) {
	isEverDelinquent, err := uc.repo.IsEverDelinquent(borrowerID)
	if err != nil {
		return entity.IsEverDelinquentResponse{}, fmt.Errorf("check is ever delinquent: %w", err)
	}

	return entity.IsEverDelinquentResponse{BorrowerID: borrowerID, IsEverDelinquent: isEverDelinquent}, nil
}
