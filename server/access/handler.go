package access

import (
	"github.com/go-chi/chi/v5"
	"github.com/ksankeerth/open-image-registry/access/namespace"
	"github.com/ksankeerth/open-image-registry/access/repository"
	"github.com/ksankeerth/open-image-registry/access/upstream"
	"github.com/ksankeerth/open-image-registry/store"
)

type RegistryAccessHandler struct {
	namespaceHandler  *namespace.NamespaceHandler
	repositoryHandler *repository.RepositoryHandler
	upstreamHandler   *upstream.UpstreamAccessHandler
}

func NewRegistryAccessHandler(s store.Store) *RegistryAccessHandler {
	return &RegistryAccessHandler{
		namespaceHandler:  namespace.NewHandler(s),
		repositoryHandler: repository.NewHandler(s),
		upstreamHandler:   upstream.NewHandler(s),
	}
}

func (h *RegistryAccessHandler) Routes() chi.Router {
	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Mount("/upstreams", h.upstreamHandler.Routes())
		r.Mount("/namespaces", h.namespaceHandler.Routes())
		r.Mount("/repositories", h.repositoryHandler.Routes())
	})

	return router
}
