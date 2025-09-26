package dockererrors

// DockerError represents a single Docker Registry v2 API error
type DockerError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Detail  interface{} `json:"detail,omitempty"`
}

// DockerErrorResponse represents the Docker Registry v2 API error response format
type DockerErrorResponse struct {
	Errors []DockerError `json:"errors"`
}

func NewDockerError(code string, detail interface{}) DockerError {
	message, exists := ErrorMessages[code]
	if !exists {
		message = "unknown error"
	}

	return DockerError{
		Code:    code,
		Message: message,
		Detail:  detail,
	}
}

func NewDockerErrorResponse(errors ...DockerError) DockerErrorResponse {
	return DockerErrorResponse{
		Errors: errors,
	}
}