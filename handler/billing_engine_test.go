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
