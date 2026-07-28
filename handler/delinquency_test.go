package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"amartha-test/entity"

	"github.com/stretchr/testify/assert"
)

func doIsDelinquent(h *httpHandler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/billing-engine/is-delinquent", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.IsDelinquent(rec, req)
	return rec
}

func doIsEverDelinquent(h *httpHandler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/billing-engine/is-ever-delinquent", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.IsEverDelinquent(rec, req)
	return rec
}

func TestBillingEngine_IsDelinquent(t *testing.T) {
	t.Run("success returns status from usecase", func(t *testing.T) {
		uc := new(mockUsecases)
		want := entity.IsDelinquentResponse{BorrowerID: 1, IsDelinquent: true}
		uc.On("IsDelinquent", int64(1)).Return(want, nil)

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doIsDelinquent(h, `{"borrower_id":1}`)

		assert.Equal(t, http.StatusOK, rec.Code)

		var got entity.IsDelinquentResponse
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Equal(t, want, got)
		uc.AssertExpectations(t)
	})

	t.Run("invalid json body returns 400", func(t *testing.T) {
		uc := new(mockUsecases)
		h := NewHTTPHandler(uc, new(mockPublisher))

		rec := doIsDelinquent(h, `not-json`)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		uc.AssertNotCalled(t, "IsDelinquent")
	})

	t.Run("borrower id zero returns 400", func(t *testing.T) {
		uc := new(mockUsecases)
		h := NewHTTPHandler(uc, new(mockPublisher))

		rec := doIsDelinquent(h, `{"borrower_id":0}`)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		uc.AssertNotCalled(t, "IsDelinquent")
	})

	t.Run("usecase error returns 500", func(t *testing.T) {
		uc := new(mockUsecases)
		uc.On("IsDelinquent", int64(1)).Return(entity.IsDelinquentResponse{}, errors.New("db down"))

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doIsDelinquent(h, `{"borrower_id":1}`)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		uc.AssertExpectations(t)
	})
}

func TestBillingEngine_IsEverDelinquent(t *testing.T) {
	t.Run("success returns status from usecase", func(t *testing.T) {
		uc := new(mockUsecases)
		want := entity.IsEverDelinquentResponse{BorrowerID: 1, IsEverDelinquent: true}
		uc.On("IsEverDelinquent", int64(1)).Return(want, nil)

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doIsEverDelinquent(h, `{"borrower_id":1}`)

		assert.Equal(t, http.StatusOK, rec.Code)

		var got entity.IsEverDelinquentResponse
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		assert.Equal(t, want, got)
		uc.AssertExpectations(t)
	})

	t.Run("invalid json body returns 400", func(t *testing.T) {
		uc := new(mockUsecases)
		h := NewHTTPHandler(uc, new(mockPublisher))

		rec := doIsEverDelinquent(h, `not-json`)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		uc.AssertNotCalled(t, "IsEverDelinquent")
	})

	t.Run("borrower id zero returns 400", func(t *testing.T) {
		uc := new(mockUsecases)
		h := NewHTTPHandler(uc, new(mockPublisher))

		rec := doIsEverDelinquent(h, `{"borrower_id":0}`)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		uc.AssertNotCalled(t, "IsEverDelinquent")
	})

	t.Run("usecase error returns 500", func(t *testing.T) {
		uc := new(mockUsecases)
		uc.On("IsEverDelinquent", int64(1)).Return(entity.IsEverDelinquentResponse{}, errors.New("db down"))

		h := NewHTTPHandler(uc, new(mockPublisher))
		rec := doIsEverDelinquent(h, `{"borrower_id":1}`)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		uc.AssertExpectations(t)
	})
}
