package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	Ping(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status":"ok","msg":"pong"}`, rec.Body.String())
}

func TestRegisterRoutes(t *testing.T) {
	mux := http.NewServeMux()
	h := &httpHandler{}

	RegisterRoutes(mux, h)

	t.Run("ping is public", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"status":"ok","msg":"pong"}`, rec.Body.String())
	})

	t.Run("get-statements without api key is unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/billing-engine/get-statements", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("get-statements with valid api key reaches the handler", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/billing-engine/get-statements", nil)
		req.Header.Set("X-API-Key", "fortestingonly")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		// h.uc is nil here, so decoding an empty/invalid body must short-circuit
		// with 400 before ever touching the nil usecase.
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
