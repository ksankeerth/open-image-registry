package helpers

import (
	"net/http"
	"testing"
)

func AssertStatusCode(t *testing.T, response *http.Response, statusCode int) {
	t.Helper()
	if response != nil {
		if response.StatusCode != statusCode {
			t.Errorf("expected status code %d but got %d", statusCode, response.StatusCode)
		}
	} else {
		t.Errorf("expected status code %d but reponse was nil", statusCode)
	}
}