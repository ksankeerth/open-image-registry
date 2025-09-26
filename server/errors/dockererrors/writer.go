package dockererrors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// WriteError writes a single Docker Registry error response
func WriteError(w http.ResponseWriter, errorCode string, detail interface{}) {
	dockerError := NewDockerError(errorCode, detail)
	response := NewDockerErrorResponse(dockerError)

	statusCode, exists := StatusCodes[errorCode]
	if !exists {
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// WriteErrors writes multiple Docker Registry errors in a single response
func WriteErrors(w http.ResponseWriter, errors []DockerError) {
	if len(errors) == 0 {
		WriteError(w, ErrCodeUnsupported, nil)
		return
	}

	response := NewDockerErrorResponse(errors...)

	// Use status code from first error
	statusCode, exists := StatusCodes[errors[0].Code]
	if !exists {
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// WriteErrorWithStatus writes an error with a custom HTTP status code
func WriteErrorWithStatus(w http.ResponseWriter, statusCode int, errorCode string, detail interface{}) {
	dockerError := NewDockerError(errorCode, detail)
	response := NewDockerErrorResponse(dockerError)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}


func WriteManifestNotFound(w http.ResponseWriter) {
	WriteError(w, ErrCodeManifestUnknown, nil)
}

func WriteBlobNotFound(w http.ResponseWriter) {
	WriteError(w, ErrCodeBlobUnknown, nil)
}

func WriteRepositoryNotFound(w http.ResponseWriter) {
	WriteError(w, ErrCodeNameUnknown, nil)
}

func WriteUnauthorized(w http.ResponseWriter, realm, service string) {
	if realm == "" {
		realm = "registry"
	}
	if service == "" {
		service = "registry"
	}

	w.Header().Set("WWW-Authenticate",
		fmt.Sprintf(`Bearer realm="%s",service="%s"`, realm, service))
	WriteError(w, ErrCodeUnauthorized, nil)
}

func WriteInvalidDigest(w http.ResponseWriter, digest string) {
	detail := map[string]string{"digest": digest}
	WriteError(w, ErrCodeDigestInvalid, detail)
}

func WriteBlobUnknownErrors(w http.ResponseWriter, unknownBlobs []string) {
	var dockerErrors []DockerError

	for _, digest := range unknownBlobs {
		detail := map[string]string{"digest": digest}
		dockerErrors = append(dockerErrors, NewDockerError(ErrCodeBlobUnknown, detail))
	}

	WriteErrors(w, dockerErrors)
}

func WriteUnsupported(w http.ResponseWriter) {
	WriteError(w, ErrCodeUnsupported, nil)
}

func WriteTooManyRequests(w http.ResponseWriter) {
	WriteError(w, ErrCodeTooManyRequests, nil)
}

func WriteInvalidRepository(w http.ResponseWriter) {
	WriteError(w, ErrCodeNameInvalid, nil)
}

func WriteInvalidTag(w http.ResponseWriter) {
	WriteError(w, ErrCodeTagInvalid, nil)
}

func WriteManifestInvalid(w http.ResponseWriter, detail interface{}) {
	WriteError(w, ErrCodeManifestInvalid, detail)
}

func WriteBlobUploadInvalid(w http.ResponseWriter) {
	WriteError(w, ErrCodeBlobUploadInvalid, nil)
}

func WriteBlobUploadNotFound(w http.ResponseWriter) {
	WriteError(w, ErrCodeBlobUploadUnknown, nil)
}

func WriteAccessDenied(w http.ResponseWriter) {
	WriteError(w, ErrCodeDenied, nil)
}

func WriteInvalidRange(w http.ResponseWriter) {
	WriteError(w, ErrCodeRangeInvalid, nil)
}

func WriteInvalidSize(w http.ResponseWriter, expected, actual int64) {
	detail := map[string]int64{
		"expected": expected,
		"actual":   actual,
	}
	WriteError(w, ErrCodeSizeInvalid, detail)
}

func WriteInvalidPagination(w http.ResponseWriter, value string) {
	detail := map[string]string{"value": value}
	WriteError(w, ErrCodePaginationNumberInvalid, detail)
}
