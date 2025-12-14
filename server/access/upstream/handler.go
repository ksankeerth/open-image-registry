package upstream

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ksankeerth/open-image-registry/store"
)

type UpstreamAccessHandler struct {
	svc *upstreamService
}

func NewHandler(s store.Store) *UpstreamAccessHandler {
	svc := &upstreamService{
		s,
	}
	return &UpstreamAccessHandler{
		svc,
	}
}

func (u *UpstreamAccessHandler) Routes() chi.Router {
	return chi.NewRouter()
}

func (u *UpstreamAccessHandler) CreateUpstreamRegistry(w http.ResponseWriter, r *http.Request) {

}

func (u *UpstreamAccessHandler) DeleteUpstreamRegistry(w http.ResponseWriter, r *http.Request) {

}

func (u *UpstreamAccessHandler) UpdateUpstreamRegistry(w http.ResponseWriter, r *http.Request) {

}

func (u *UpstreamAccessHandler) UpdateUpstreamRegistryAuthConfig(w http.ResponseWriter, r *http.Request) {

}

func (u *UpstreamAccessHandler) UpdateUpstreamRegistryCacheConfig(w http.ResponseWriter, r *http.Request) {

}

func (u *UpstreamAccessHandler) UpdateUpstreamRegistryNetworkConfig(w http.ResponseWriter, r *http.Request) {

}

func (u *UpstreamAccessHandler) ChangeUpstreamRegistryState(w http.ResponseWriter, r *http.Request) {

}

func (u *UpstreamAccessHandler) GetUserAccessList(w http.ResponseWriter, r *http.Request) {

}
