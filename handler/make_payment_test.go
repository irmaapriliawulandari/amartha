package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"amartha-test/entity"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func doMakePayment(h *httpHandler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/billing-engine/make-payment", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.MakePayment(rec, req)
	return rec
}

func TestBillingEngine_MakePayment(t *testing.T) {
	t.Run("success returns payment confirmation from usecase", func(t *testing.T) {
		uc := new(mockUsecases)
		want := entity.MakePaymentResponse{
			LoanID: 1, StatementID: 5, PaidAmount: decimal.NewFromInt(120000),
			PaidAt: "2026-08-25 00:00:00", OutstandingAmount: decimal.NewFromInt(4380000),
		}
		uc.On("MakePayment", int64(1), mock.Anything).Return(want, nil)

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doMakePayment(h, `{"loan_id":1}`)

		assert.Equal(t, http.StatusOK, rec.Code)

		var got entity.MakePaymentResponse
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Equal(t, want, got)
		uc.AssertExpectations(t)
	})

	t.Run("invalid json body returns 400", func(t *testing.T) {
		uc := new(mockUsecases)
		h := NewHTTPHandler(uc, new(mockPublisher))

		rec := doMakePayment(h, `not-json`)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		uc.AssertNotCalled(t, "MakePayment")
	})

	t.Run("loan id zero returns 400", func(t *testing.T) {
		uc := new(mockUsecases)
		h := NewHTTPHandler(uc, new(mockPublisher))

		rec := doMakePayment(h, `{"loan_id":0}`)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		uc.AssertNotCalled(t, "MakePayment")
	})

	t.Run("no statement to pay returns 404", func(t *testing.T) {
		uc := new(mockUsecases)
		uc.On("MakePayment", int64(1), mock.Anything).Return(entity.MakePaymentResponse{}, entity.ErrStatementNotFound)

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doMakePayment(h, `{"loan_id":1}`)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		uc.AssertExpectations(t)
	})

	t.Run("loan not found returns 404", func(t *testing.T) {
		uc := new(mockUsecases)
		uc.On("MakePayment", int64(1), mock.Anything).Return(entity.MakePaymentResponse{}, entity.ErrLoanNotFound)

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doMakePayment(h, `{"loan_id":1}`)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		uc.AssertExpectations(t)
	})

	t.Run("already paid returns 409", func(t *testing.T) {
		uc := new(mockUsecases)
		uc.On("MakePayment", int64(1), mock.Anything).Return(entity.MakePaymentResponse{}, entity.ErrStatementAlreadyPaid)

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doMakePayment(h, `{"loan_id":1}`)

		assert.Equal(t, http.StatusConflict, rec.Code)
		uc.AssertExpectations(t)
	})

	t.Run("usecase error returns 500", func(t *testing.T) {
		uc := new(mockUsecases)
		uc.On("MakePayment", int64(1), mock.Anything).Return(entity.MakePaymentResponse{}, errors.New("db down"))

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doMakePayment(h, `{"loan_id":1}`)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		uc.AssertExpectations(t)
	})
}
