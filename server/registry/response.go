package registry

import (
	"fmt"
	"net/http"
)

func writeBlobExistsResponse(w http.ResponseWriter, digest string) {
	w.Header().Set("Content-Length", "0")
	w.Header().Set("Docker-Content-Digest", digest)
	w.WriteHeader(http.StatusOK)
}

func writeBlobUploadSuccess(w http.ResponseWriter, url, sessionId string, start, end int) {
	w.Header().Add("Location", url)
	w.Header().Add("Range", fmt.Sprintf("%d-%d", start, end))
	w.Header().Add("Docker-Upload-UUID", sessionId)
	w.Header().Set("Content-Length", "0")
	w.WriteHeader(http.StatusCreated)
}