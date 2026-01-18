package resource

import (
	"github.com/go-chi/chi/v5"

	acesss "github.com/ksankeerth/open-image-registry/resource/access"
	"github.com/ksankeerth/open-image-registry/resource/namespace"
	"github.com/ksankeerth/open-image-registry/resource/repository"
	"github.com/ksankeerth/open-image-registry/resource/upstream"
	"github.com/ksankeerth/open-image-registry/store"
)

type RegistryResourceHandler struct {
	namespaceHandler  *namespace.NamespaceHandler
	repositoryHandler *repository.RepositoryHandler
	upstreamHandler   *upstream.UpstreamAccessHandler
}

func NewRegistryResourceHandler(s store.Store, accessManager *acesss.Manager) *RegistryResourceHandler {
	return &RegistryResourceHandler{
		namespaceHandler:  namespace.NewHandler(s, accessManager),
		repositoryHandler: repository.NewHandler(s, accessManager),
		upstreamHandler:   upstream.NewHandler(s),
	}
}

func (h *RegistryResourceHandler) Routes() chi.Router {
	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Mount("/upstreams", h.upstreamHandler.Routes())
		r.Mount("/namespaces", h.namespaceHandler.Routes())
		r.Mount("/repositories", h.repositoryHandler.Routes())
	})

	return router
}