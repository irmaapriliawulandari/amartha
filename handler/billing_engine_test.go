package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"amartha-test/entity"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUsecases struct {
	mock.Mock
}

func (m *mockUsecases) GetStatements(loanID int64, until time.Time, limit, offset int) ([]entity.Statement, error) {
	args := m.Called(loanID, until, limit, offset)

	var res []entity.Statement
	if args.Get(0) != nil {
		res = args.Get(0).([]entity.Statement)
	}
	return res, args.Error(1)
}

func (m *mockUsecases) GetOutstandingAmount(loanID int64) (entity.OutstandingAmount, error) {
	args := m.Called(loanID)
	return args.Get(0).(entity.OutstandingAmount), args.Error(1)
}

func (m *mockUsecases) GetLatestStatement(loanID int64, now time.Time) (entity.LatestStatement, error) {
	args := m.Called(loanID, now)
	return args.Get(0).(entity.LatestStatement), args.Error(1)
}

type mockPublisher struct {
	mock.Mock
}

func (m *mockPublisher) Publish(topic string, body []byte) error {
	args := m.Called(topic, body)
	return args.Error(0)
}

func doGetStatements(h *httpHandler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/billing-engine/get-statements", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.GetStatements(rec, req)
	return rec
}

func doGetOutstandingAmount(h *httpHandler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/billing-engine/get-outstanding", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.GetOutstandingAmount(rec, req)
	return rec
}

func doGetLatestStatement(h *httpHandler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/billing-engine/get-latest-statement", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.GetLatestStatement(rec, req)
	return rec
}

func TestBillingEngine_GetStatements(t *testing.T) {
	t.Run("success returns statements from usecase", func(t *testing.T) {
		uc := new(mockUsecases)
		want := []entity.Statement{
			{LoanID: 1, StatementDate: "2026-07-25", ToPayAmount: decimal.NewFromInt(100000), Status: "Unpaid"},
		}
		uc.On("GetStatements", int64(1), mock.Anything, 5, 2).Return(want, nil)

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doGetStatements(h, `{"loan_id":1,"limit":5,"offset":2}`)

		assert.Equal(t, http.StatusOK, rec.Code)

		var got []entity.Statement
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Equal(t, want, got)
		uc.AssertExpectations(t)
	})

	t.Run("invalid json body returns 400", func(t *testing.T) {
		uc := new(mockUsecases)
		h := NewHTTPHandler(uc, new(mockPublisher))

		rec := doGetStatements(h, `not-json`)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		uc.AssertNotCalled(t, "GetStatements")
	})

	t.Run("loan id zero returns 400", func(t *testing.T) {
		uc := new(mockUsecases)
		h := NewHTTPHandler(uc, new(mockPublisher))

		rec := doGetStatements(h, `{"loan_id":0,"limit":5,"offset":0}`)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		uc.AssertNotCalled(t, "GetStatements")
	})

	t.Run("usecase error returns 500", func(t *testing.T) {
		uc := new(mockUsecases)
		uc.On("GetStatements", int64(1), mock.Anything, 5, 0).Return(nil, errors.New("db down"))

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doGetStatements(h, `{"loan_id":1,"limit":5,"offset":0}`)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		uc.AssertExpectations(t)
	})
}

func TestBillingEngine_GetOutstandingAmount(t *testing.T) {
	t.Run("success returns outstanding amount from usecase", func(t *testing.T) {
		uc := new(mockUsecases)
		want := entity.OutstandingAmount{LoanID: 1, OutstandingAmount: decimal.NewFromInt(4500000)}
		uc.On("GetOutstandingAmount", int64(1)).Return(want, nil)

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doGetOutstandingAmount(h, `{"loan_id":1}`)

		assert.Equal(t, http.StatusOK, rec.Code)

		var got entity.OutstandingAmount
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Equal(t, want, got)
		uc.AssertExpectations(t)
	})

	t.Run("invalid json body returns 400", func(t *testing.T) {
		uc := new(mockUsecases)
		h := NewHTTPHandler(uc, new(mockPublisher))

		rec := doGetOutstandingAmount(h, `not-json`)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		uc.AssertNotCalled(t, "GetOutstandingAmount")
	})

	t.Run("loan id zero returns 400", func(t *testing.T) {
		uc := new(mockUsecases)
		h := NewHTTPHandler(uc, new(mockPublisher))

		rec := doGetOutstandingAmount(h, `{"loan_id":0}`)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		uc.AssertNotCalled(t, "GetOutstandingAmount")
	})

	t.Run("loan not found returns 404", func(t *testing.T) {
		uc := new(mockUsecases)
		uc.On("GetOutstandingAmount", int64(1)).Return(entity.OutstandingAmount{}, entity.ErrLoanNotFound)

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doGetOutstandingAmount(h, `{"loan_id":1}`)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		uc.AssertExpectations(t)
	})

	t.Run("usecase error returns 500", func(t *testing.T) {
		uc := new(mockUsecases)
		uc.On("GetOutstandingAmount", int64(1)).Return(entity.OutstandingAmount{}, errors.New("db down"))

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doGetOutstandingAmount(h, `{"loan_id":1}`)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		uc.AssertExpectations(t)
	})
}

func TestBillingEngine_GetLatestStatement(t *testing.T) {
	t.Run("success returns latest statement from usecase", func(t *testing.T) {
		uc := new(mockUsecases)
		want := entity.LatestStatement{
			LoanID: 1, StatementDate: "2026-08-18", CarryOverAmount: decimal.NewFromInt(10000), InstallmentAmount: decimal.NewFromInt(110000),
			TotalToPay: decimal.NewFromInt(120000), Status: "Paid", Deadline: "2026-08-24",
			OutstandingAmount: decimal.NewFromInt(4500000),
		}
		uc.On("GetLatestStatement", int64(1), mock.Anything).Return(want, nil)

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doGetLatestStatement(h, `{"loan_id":1}`)

		assert.Equal(t, http.StatusOK, rec.Code)

		var got entity.LatestStatement
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Equal(t, want, got)
		uc.AssertExpectations(t)
	})

	t.Run("invalid json body returns 400", func(t *testing.T) {
		uc := new(mockUsecases)
		h := NewHTTPHandler(uc, new(mockPublisher))

		rec := doGetLatestStatement(h, `not-json`)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		uc.AssertNotCalled(t, "GetLatestStatement")
	})

	t.Run("loan id zero returns 400", func(t *testing.T) {
		uc := new(mockUsecases)
		h := NewHTTPHandler(uc, new(mockPublisher))

		rec := doGetLatestStatement(h, `{"loan_id":0}`)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		uc.AssertNotCalled(t, "GetLatestStatement")
	})

	t.Run("no statement found returns 404", func(t *testing.T) {
		uc := new(mockUsecases)
		uc.On("GetLatestStatement", int64(1), mock.Anything).Return(entity.LatestStatement{}, entity.ErrStatementNotFound)

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doGetLatestStatement(h, `{"loan_id":1}`)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		uc.AssertExpectations(t)
	})

	t.Run("loan not found returns 404", func(t *testing.T) {
		uc := new(mockUsecases)
		uc.On("GetLatestStatement", int64(1), mock.Anything).Return(entity.LatestStatement{}, entity.ErrLoanNotFound)

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doGetLatestStatement(h, `{"loan_id":1}`)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		uc.AssertExpectations(t)
	})

	t.Run("usecase error returns 500", func(t *testing.T) {
		uc := new(mockUsecases)
		uc.On("GetLatestStatement", int64(1), mock.Anything).Return(entity.LatestStatement{}, errors.New("db down"))

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doGetLatestStatement(h, `{"loan_id":1}`)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		uc.AssertExpectations(t)
	})
}
