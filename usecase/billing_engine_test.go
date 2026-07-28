package usecase

import (
	"errors"
	"testing"
	"time"

	"amartha-test/entity"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) GetStatements(loanID int64, until time.Time, limit, offset int) ([]entity.StatementDB, error) {
	args := m.Called(loanID, until, limit, offset)

	var res []entity.StatementDB
	if args.Get(0) != nil {
		res = args.Get(0).([]entity.StatementDB)
	}
	return res, args.Error(1)
}

func (m *mockRepo) InsertLoan(loan entity.LoanDB) error {
	args := m.Called(loan)
	return args.Error(0)
}

func (m *mockRepo) InsertStatement(statement entity.StatementDB) (int64, error) {
	args := m.Called(statement)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRepo) InsertDelinquencyHist(dh entity.DelinquencyHistDB) (int64, error) {
	args := m.Called(dh)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRepo) InsertLoanWithStatements(loan entity.LoanDB, statements []entity.StatementDB) error {
	args := m.Called(loan, statements)
	return args.Error(0)
}

func (m *mockRepo) GetOutstandingAmount(loanID int64) (decimal.Decimal, error) {
	args := m.Called(loanID)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *mockRepo) GetLatestStatement(loanID int64, before time.Time) (entity.StatementDB, error) {
	args := m.Called(loanID, before)
	return args.Get(0).(entity.StatementDB), args.Error(1)
}

func TestBillingEngineUsecase_GetStatements(t *testing.T) {
	loanID := int64(1)
	until := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	limit, offset := 10, 0

	t.Run("success maps db rows to statements", func(t *testing.T) {
		repo := new(mockRepo)
		repo.On("GetStatements", loanID, until, limit, offset).Return([]entity.StatementDB{
			{
				LoanID: loanID, StatementDate: time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
				InstallmentAmount: decimal.NewFromInt(100000), CarryOverAmount: decimal.Zero, PaidAmount: decimal.Zero,
				Status: entity.StatementStatusUnpaid,
			},
			{
				LoanID: loanID, StatementDate: time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
				InstallmentAmount: decimal.NewFromInt(100000), CarryOverAmount: decimal.Zero, PaidAmount: decimal.NewFromInt(100000),
				Status: entity.StatementStatusPaid,
			},
			{
				LoanID: loanID, StatementDate: time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC),
				InstallmentAmount: decimal.NewFromInt(100000), CarryOverAmount: decimal.Zero, PaidAmount: decimal.Zero,
				Status: entity.StatementStatusOverdue,
			},
		}, nil)

		uc := NewBillingEngineUsecase(repo)
		got, err := uc.GetStatements(loanID, until, limit, offset)

		assert.NoError(t, err)
		assert.Equal(t, []entity.Statement{
			{LoanID: loanID, StatementDate: "2026-07-25", ToPayAmount: decimal.NewFromInt(100000), Status: "Unpaid"},
			{LoanID: loanID, StatementDate: "2026-07-18", ToPayAmount: decimal.NewFromInt(100000), Status: "Paid"},
			{LoanID: loanID, StatementDate: "2026-07-11", ToPayAmount: decimal.NewFromInt(100000), Status: "Overdue"},
		}, got)
		repo.AssertExpectations(t)
	})

	t.Run("empty result returns empty slice, not nil", func(t *testing.T) {
		repo := new(mockRepo)
		repo.On("GetStatements", loanID, until, limit, offset).Return([]entity.StatementDB{}, nil)

		uc := NewBillingEngineUsecase(repo)
		got, err := uc.GetStatements(loanID, until, limit, offset)

		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("repo error is wrapped", func(t *testing.T) {
		repo := new(mockRepo)
		repoErr := errors.New("connection lost")
		repo.On("GetStatements", loanID, until, limit, offset).Return(nil, repoErr)

		uc := NewBillingEngineUsecase(repo)
		got, err := uc.GetStatements(loanID, until, limit, offset)

		assert.Nil(t, got)
		assert.ErrorIs(t, err, repoErr)
	})
}

func TestBillingEngineUsecase_GetOutstandingAmount(t *testing.T) {
	loanID := int64(1)

	t.Run("success returns outstanding amount", func(t *testing.T) {
		repo := new(mockRepo)
		repo.On("GetOutstandingAmount", loanID).Return(decimal.NewFromInt(4500000), nil)

		uc := NewBillingEngineUsecase(repo)
		got, err := uc.GetOutstandingAmount(loanID)

		assert.NoError(t, err)
		assert.Equal(t, entity.OutstandingAmount{LoanID: loanID, OutstandingAmount: decimal.NewFromInt(4500000)}, got)
	})

	t.Run("loan not found is passed through", func(t *testing.T) {
		repo := new(mockRepo)
		repo.On("GetOutstandingAmount", loanID).Return(decimal.Decimal{}, entity.ErrLoanNotFound)

		uc := NewBillingEngineUsecase(repo)
		_, err := uc.GetOutstandingAmount(loanID)

		assert.ErrorIs(t, err, entity.ErrLoanNotFound)
	})

	t.Run("repo error is wrapped", func(t *testing.T) {
		repo := new(mockRepo)
		repoErr := errors.New("connection lost")
		repo.On("GetOutstandingAmount", loanID).Return(decimal.Decimal{}, repoErr)

		uc := NewBillingEngineUsecase(repo)
		_, err := uc.GetOutstandingAmount(loanID)

		assert.ErrorIs(t, err, repoErr)
	})
}

func TestBillingEngineUsecase_GetLatestStatement(t *testing.T) {
	loanID := int64(1)
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)

	t.Run("success combines latest statement and outstanding amount, paid or not", func(t *testing.T) {
		repo := new(mockRepo)
		repo.On("GetLatestStatement", loanID, now).Return(entity.StatementDB{
			LoanID:            loanID,
			InstallmentAmount: decimal.NewFromInt(110000),
			CarryOverAmount:   decimal.NewFromInt(10000),
			StatementDate:     time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC),
			Deadline:          time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC),
			Status:            entity.StatementStatusPaid,
		}, nil)
		repo.On("GetOutstandingAmount", loanID).Return(decimal.NewFromInt(4500000), nil)

		uc := NewBillingEngineUsecase(repo)
		got, err := uc.GetLatestStatement(loanID, now)

		assert.NoError(t, err)
		assert.Equal(t, entity.LatestStatement{
			LoanID:            loanID,
			StatementDate:     "2026-08-18",
			CarryOverAmount:   decimal.NewFromInt(10000),
			InstallmentAmount: decimal.NewFromInt(110000),
			TotalToPay:        decimal.NewFromInt(120000),
			Status:            "Paid",
			Deadline:          "2026-08-24",
			OutstandingAmount: decimal.NewFromInt(4500000),
		}, got)
	})

	t.Run("no statement before the reference date is passed through", func(t *testing.T) {
		repo := new(mockRepo)
		repo.On("GetLatestStatement", loanID, now).Return(entity.StatementDB{}, entity.ErrStatementNotFound)

		uc := NewBillingEngineUsecase(repo)
		_, err := uc.GetLatestStatement(loanID, now)

		assert.ErrorIs(t, err, entity.ErrStatementNotFound)
		repo.AssertNotCalled(t, "GetOutstandingAmount", mock.Anything)
	})

	t.Run("outstanding amount repo error is wrapped", func(t *testing.T) {
		repo := new(mockRepo)
		repo.On("GetLatestStatement", loanID, now).Return(entity.StatementDB{LoanID: loanID}, nil)
		repoErr := errors.New("connection lost")
		repo.On("GetOutstandingAmount", loanID).Return(decimal.Decimal{}, repoErr)

		uc := NewBillingEngineUsecase(repo)
		_, err := uc.GetLatestStatement(loanID, now)

		assert.ErrorIs(t, err, repoErr)
	})
}
