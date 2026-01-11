package middleware

import (
	"net/http"
	"strings"
)

func EnforceJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			ct := r.Header.Get("Content-Type")

			if ct != "" && !strings.HasPrefix(strings.ToLower(ct), "application/json") {
				w.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}