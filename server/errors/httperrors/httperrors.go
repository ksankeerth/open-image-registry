package httperrors

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the standard JSON response for HTTP errors
type ErrorResponse struct {
	Message string `json:"error_message"`
	Code    int    `json:"error_code"`
}

// HTTPError defines an error type with a status code and app-specific code
type HTTPError struct {
	HTTPStatus int
	Code       int
	Message    string
}

func SendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if encodeErr := json.NewEncoder(w).Encode(ErrorResponse{
		Code:    statusCode,
		Message: message,
	}); encodeErr != nil {
		// Optionally log the error if JSON encoding fails
		// log.Println("Failed to encode error response:", encodeErr)
	}
}

// writeError writes the error response as JSON
func writeError(w http.ResponseWriter, err HTTPError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatus)
	if encodeErr := json.NewEncoder(w).Encode(ErrorResponse{
		Code:    err.Code,
		Message: err.Message,
	}); encodeErr != nil {
		// Optionally log the error if JSON encoding fails
		// log.Println("Failed to encode error response:", encodeErr)
	}
}

// Predefined helper functions
func AlreadyExist(w http.ResponseWriter, code int, msg string) {
	writeError(w, HTTPError{HTTPStatus: http.StatusConflict, Code: code, Message: msg})
}

func InternalError(w http.ResponseWriter, code int, msg string) {
	writeError(w, HTTPError{HTTPStatus: http.StatusInternalServerError, Code: code, Message: msg})
}

func BadRequest(w http.ResponseWriter, code int, msg string) {
	writeError(w, HTTPError{HTTPStatus: http.StatusBadRequest, Code: code, Message: msg})
}

func NotFound(w http.ResponseWriter, code int, msg string) {
	writeError(w, HTTPError{HTTPStatus: http.StatusNotFound, Code: code, Message: msg})
}

func NotAllowed(w http.ResponseWriter, code int, msg string) {
	writeError(w, HTTPError{HTTPStatus: http.StatusForbidden, Code: code, Message: msg})
}

func Unauthorized(w http.ResponseWriter, code int, msg string) {
	writeError(w, HTTPError{HTTPStatus: http.StatusUnauthorized, Code: code, Message: msg})
}