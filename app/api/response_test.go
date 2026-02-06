package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/services"
	"github.com/stretchr/testify/assert"
)

func TestOKResponse(t *testing.T) {

	type sampleResponse struct {
		Message string `json:"message"`
	}

	sample := sampleResponse{Message: "Success"}

	t.Run("succesful http200 json response", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		OKResponse(recorder, req, sample)

		assert.Equal(t, http.StatusOK, recorder.Code, "Expected status code 200 OK")
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"), "Expected Content-Type to be application/json")

		expected := `{"message":"Success"}`
		assert.JSONEq(t, expected, recorder.Body.String(), "Response body does not match expected")
	})
}

func TestHandleError(t *testing.T) {
	t.Run("handles invalid input error", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		HandleError(recorder, req, services.ErrInvalidInput)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		expected := `{"code":"invalid_input","message":"Invalid input provided"}`
		assert.JSONEq(t, expected, recorder.Body.String())
	})

	t.Run("handles not found error", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		HandleError(recorder, req, services.ErrNotFound)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		expected := `{"code":"not_found","message":"Resource not found"}`
		assert.JSONEq(t, expected, recorder.Body.String())
	})

	t.Run("handles internal error", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		HandleError(recorder, req, errors.New("some internal error"))

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		expected := `{"code":"internal_error","message":"An internal error occurred"}`
		assert.JSONEq(t, expected, recorder.Body.String())
	})
}

func TestErrorHandler(t *testing.T) {
	t.Run("writes error response when handler returns error", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		handler := ErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
			return services.ErrInvalidInput
		})
		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		expected := `{"code":"invalid_input","message":"Invalid input provided"}`
		assert.JSONEq(t, expected, recorder.Body.String())
	})
}
