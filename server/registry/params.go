package registry

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func extractNamespaceRepositoryAndDigest(r *http.Request) (namespace, repository, digest string) {
	namespace = chi.URLParam(r, "namespace")
	if namespace == "" {
		namespace = DefaultNamespace
	}
	repository = chi.URLParam(r, "repository")
	digest = chi.URLParam(r, "digest")
	return
}

func extractNamespaceRepositoryAndSessionId(r *http.Request) (namespace, repository, sessionId string) {
	namespace = chi.URLParam(r, "namespace")
	if namespace == "" {
		namespace = DefaultNamespace
	}
	repository = chi.URLParam(r, "repository")
	sessionId = chi.URLParam(r, "session_id")
	return
}

func extractNamespaceRepositoryAndTagOrDigest(r *http.Request) (namespace, repository, tagOrDigest string) {
	namespace = chi.URLParam(r, "namespace")
	if namespace == "" {
		namespace = DefaultNamespace
	}
	repository = chi.URLParam(r, "repository")
	tagOrDigest = chi.URLParam(r, "tag_or_digest")
	return
}

func extractNamespaceRepositoryAndTag(r *http.Request) (namespace, repository, tag string) {
	tag = chi.URLParam(r, "tag")
	namespace = chi.URLParam(r, "namespace")
	if namespace == "" {
		namespace = DefaultNamespace
	}
	repository = chi.URLParam(r, "repository")
	return
}

func extractNamespaceAndRepository(r *http.Request) (namespace, repository string) {
	namespace = chi.URLParam(r, "namespace")
	if namespace == "" {
		namespace = DefaultNamespace
	}
	repository = chi.URLParam(r, "repository")
	return
}