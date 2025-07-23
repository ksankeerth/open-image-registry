package common

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"error_message"`
	Code    int    `json:"error_code"`
}

func writeError(w http.ResponseWriter, statusCode int, errorCode int, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Code: errorCode, Message: errorMsg})
}

func HandleAlreadyExist(w http.ResponseWriter, errorCode int, errorMessage string) {
	writeError(w, http.StatusConflict, errorCode, errorMessage)
}

func HandleInternalError(w http.ResponseWriter, errorCode int, errorMessage string) {
	writeError(w, http.StatusInternalServerError, errorCode, errorMessage)
}

func HandleBadRequest(w http.ResponseWriter, errorCode int, errorMessage string) {
	writeError(w, http.StatusBadRequest, errorCode, errorMessage)
}

func HandleNotFound(w http.ResponseWriter, errorCode int, errorMesssge string) {
	writeError(w, http.StatusNotFound, errorCode, errorMesssge)
}
