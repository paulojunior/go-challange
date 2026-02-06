package api

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/services"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
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

	t.Run("returns nil when handler succeeds", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		handler := ErrorHandler(func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusOK)
			return nil
		})
		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestCreatedResponse(t *testing.T) {
	type sampleResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	sample := sampleResponse{ID: 1, Name: "Created"}

	t.Run("successful http201 json response", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		CreatedResponse(recorder, req, sample)

		assert.Equal(t, http.StatusCreated, recorder.Code, "Expected status code 201 Created")
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"), "Expected Content-Type to be application/json")

		expected := `{"id":1,"name":"Created"}`
		assert.JSONEq(t, expected, recorder.Body.String(), "Response body does not match expected")
	})

	t.Run("handles encode error gracefully", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/test", nil)

		// Use a channel which cannot be JSON encoded to trigger encoding error
		CreatedResponse(recorder, req, make(chan int))

		assert.Equal(t, http.StatusCreated, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	})
}

func TestHandleError_SpecificValidationErrors(t *testing.T) {
	t.Run("handles ErrInvalidOffset", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		HandleError(recorder, req, services.ErrInvalidOffset)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		expected := `{"code":"invalid_input","message":"offset must be a non-negative integer"}`
		assert.JSONEq(t, expected, recorder.Body.String())
	})

	t.Run("handles ErrInvalidLimit", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		HandleError(recorder, req, services.ErrInvalidLimit)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		expected := `{"code":"invalid_input","message":"limit must be a positive integer"}`
		assert.JSONEq(t, expected, recorder.Body.String())
	})

	t.Run("handles ErrInvalidPrice", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		HandleError(recorder, req, services.ErrInvalidPrice)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		expected := `{"code":"invalid_input","message":"priceLessThan must be a valid decimal number"}`
		assert.JSONEq(t, expected, recorder.Body.String())
	})

	t.Run("handles ErrNegativePrice", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		HandleError(recorder, req, services.ErrNegativePrice)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		expected := `{"code":"invalid_input","message":"priceLessThan must be a non-negative value"}`
		assert.JSONEq(t, expected, recorder.Body.String())
	})

	t.Run("handles ErrInvalidCategoryInput", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		HandleError(recorder, req, services.ErrInvalidCategoryInput)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		expected := `{"code":"invalid_input","message":"category code and name are required"}`
		assert.JSONEq(t, expected, recorder.Body.String())
	})

	t.Run("handles gorm.ErrRecordNotFound", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		HandleError(recorder, req, gorm.ErrRecordNotFound)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		expected := `{"code":"not_found","message":"Resource not found"}`
		assert.JSONEq(t, expected, recorder.Body.String())
	})
}

func TestOKResponse_EncodeError(t *testing.T) {
	t.Run("handles encode error gracefully", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		// Use a channel which cannot be JSON encoded to trigger encoding error
		OKResponse(recorder, req, make(chan int))

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	})
}

func TestHandleError_EncodeError(t *testing.T) {
	t.Run("handles json encode error gracefully", func(t *testing.T) {
		// Create a response writer that will fail when trying to write
		recorder := &brokenWriter{header: make(http.Header)}
		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		// This should trigger the json encode error path
		HandleError(recorder, req, services.ErrInvalidInput)

		assert.Equal(t, http.StatusBadRequest, recorder.statusCode)
	})
}

// brokenWriter is a ResponseWriter that fails on Write calls
type brokenWriter struct {
	header     http.Header
	statusCode int
}

func (b *brokenWriter) Header() http.Header {
	return b.header
}

func (b *brokenWriter) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func (b *brokenWriter) WriteHeader(statusCode int) {
	b.statusCode = statusCode
}
