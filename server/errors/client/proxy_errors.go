package client

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

const (
	CodeProxyConnectionFailed     = 7001
	CodeProxyUnexpectedStatusCode = 7002
	CodeProxyResponseBodyMismatch = 7003
	CodeProxyUnsupportedMediaType = 7004
	CodeProxyArtifactNotFound     = 7005
	CodeUnclassifiedClientError   = 7049
)

var (
	ErrProxyConnectionFailed = &ProxyClientError{
		errCode: CodeProxyConnectionFailed,
	}
	ErrProxyUnexpectedStatusCode = &ProxyClientError{
		errCode: CodeProxyUnexpectedStatusCode,
	}
	ErrProxyResponseBodyMismatch = &ProxyClientError{
		errCode: CodeProxyResponseBodyMismatch,
	}
	ErrProxyUnsupportedMediaType = &ProxyClientError{
		errCode: CodeProxyUnsupportedMediaType,
	}
	ErrProxyArtifactNotFound = &ProxyClientError{
		errCode: CodeProxyArtifactNotFound,
	}
	ErrUnclassifiedClientError = &ProxyClientError{
		errCode: CodeUnclassifiedClientError,
	}
)

// ClientError represents an error returned by a proxy/registry client.
type ProxyClientError struct {
	request string
	err     error
	errCode int
}

func NewProxyClientError(req string, err error, errCode int) *ProxyClientError {
	return &ProxyClientError{
		request: req,
		err:     err,
		errCode: errCode,
	}
}

func (ce *ProxyClientError) Error() string {
	msg := fmt.Sprintf("client error: [%d]", ce.errCode)
	if ce.request != "" {
		msg += fmt.Sprintf(" (request: %q)", ce.request)
	}
	return msg
}

func (ce *ProxyClientError) Unwrap() error {
	return ce.err
}

func (ce *ProxyClientError) Is(target error) bool {
	if cErr, ok := target.(*ProxyClientError); ok {
		return cErr.errCode == ce.errCode
	}
	return false
}

func ClassifyError(err error, request string, resp *http.Response) *ProxyClientError {
	if err == nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return &ProxyClientError{
				errCode: CodeProxyArtifactNotFound,
				request: request,
			}
		}

		return &ProxyClientError{
			errCode: CodeUnclassifiedClientError,
			request: request,
		}
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return &ProxyClientError{
			err:     err,
			errCode: CodeProxyConnectionFailed,
			request: request,
		}
	}

	if errors.Is(err, io.ErrUnexpectedEOF) {
		return &ProxyClientError{
			err:     err,
			errCode: CodeProxyResponseBodyMismatch,
			request: request,
		}
	}

	if resp != nil && resp.StatusCode >= 400 {
		return &ProxyClientError{
			err:     fmt.Errorf("unexpected status code: %d", resp.StatusCode),
			errCode: CodeProxyUnexpectedStatusCode,
			request: request,
		}
	}

	return &ProxyClientError{
		err:     err,
		errCode: CodeUnclassifiedClientError,
		request: request,
	}
}

func UnwrapClientError(err error) (ok bool, errorCode int) {
	if ce, ok := err.(*ProxyClientError); ok {
		return ok, ce.errCode
	}
	return false, CodeUnclassifiedClientError
}

func IsProxyConnectionFailed(err error) bool {
	cErr, ok := err.(*ProxyClientError)
	return ok && cErr.errCode == CodeProxyConnectionFailed
}

func IsProxyUnexpectedStatus(err error) bool {
	cErr, ok := err.(*ProxyClientError)
	return ok && cErr.errCode == CodeProxyUnexpectedStatusCode
}

func IsProxyResponseBodyMismatch(err error) bool {
	cErr, ok := err.(*ProxyClientError)
	return ok && cErr.errCode == CodeProxyResponseBodyMismatch
}

func IsNotFound(err error) bool {
	cErr, ok := err.(*ProxyClientError)
	return ok && cErr.errCode == CodeProxyArtifactNotFound
}

func ProxyUnexpectedStatusError(expected, actual int, req string) *ProxyClientError {
	return &ProxyClientError{
		errCode: CodeProxyUnexpectedStatusCode,
		err:     fmt.Errorf("Expected Status code: %d, got: %d", expected, actual),
		request: req,
	}
}

func ProxyUnsupportedMediaTypeError(mediaType, req string) *ProxyClientError {
	return &ProxyClientError{
		errCode: CodeProxyUnexpectedStatusCode,
		err:     fmt.Errorf("Unsuppoted Media type: %s", mediaType),
		request: req,
	}
}