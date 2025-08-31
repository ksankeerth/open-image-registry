package routes

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	"github.com/ksankeerth/open-image-registry/db"
	"github.com/ksankeerth/open-image-registry/handlers/api"
)

func InitRouter(webappBuildPath string) *chi.Mux {

	router := chi.NewRouter()

	// Cors configuration
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	router.Use(corsMiddleware.Handler)
	router.Use(httplog.RequestLogger(httplog.NewLogger("ManagementAPI", httplog.Options{
		LogLevel:         slog.LevelDebug,
		Concise:          true,
		RequestHeaders:   true,
		MessageFieldName: "message",
	})))

	upstreamRegistryHandler := api.NewUpstreamRegistryHandler(db.GetUpstreamDAO())

	router.Route("/api/v1/upstreams", func(r chi.Router) {
		r.Put("/{registry_id}", upstreamRegistryHandler.UpdateUpstreamRegistry)
		r.Get("/{registry_id}", upstreamRegistryHandler.GetUpstreamRegistry)
		r.Delete("/{registry_id}", upstreamRegistryHandler.DeleteUpstreamRegistry)
		r.Post("/", upstreamRegistryHandler.CreateUpstreamRegistry)
		r.Get("/", upstreamRegistryHandler.ListUpstreamRegistries)
	})

	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		requestedPath := filepath.Join(webappBuildPath, r.URL.Path)
		fmt.Printf("Requested path: %s\n", requestedPath)

		if _, err := os.Stat(requestedPath); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(webappBuildPath, "index.html"))
			return
		}

		http.FileServer(http.Dir(webappBuildPath)).ServeHTTP(w, r)
	})

	return router
}

func GetDockerV2Router(regId, regName string) *chi.Mux {
	router := chi.NewRouter()

	router.Use(httplog.RequestLogger(httplog.NewLogger(fmt.Sprintf("DockerV2API-%s", regName), httplog.Options{
		LogLevel:         slog.LevelDebug,
		Concise:          true,
		RequestHeaders:   true,
		MessageFieldName: "message",
	})))

	dockerV2Handler := api.NewDockerV2Handler(regId, regName, db.GetImageRegDAO())

	router.Route("/v2", func(r chi.Router) {
		r.Get("/", dockerV2Handler.GetDockerV2APISupport)

		//blob
		r.Post("/{namespace}/{repository}/blobs/uploads/", dockerV2Handler.InitiateImageBlobUpload)
		r.Post("/{repository}/blobs/uploads/", dockerV2Handler.InitiateImageBlobUpload)

		r.Head("/{namespace}/{repository}/blobs/{digest}", dockerV2Handler.CheckImageBlobExistence)
		r.Head("/{repository}/blobs/{digest}", dockerV2Handler.CheckImageBlobExistence)

		r.Put("/{namespace}/{repository}/blobs/uploads/{session_id}", dockerV2Handler.HandleImageBlobUpload)
		r.Put("/{repository}/blobs/uploads/{session_id}", dockerV2Handler.HandleImageBlobUpload)

		r.Patch("/{namespace}/{repository}/blobs/uploads/{session_id}", dockerV2Handler.HandleImageBlobUpload)
		r.Patch("/{repository}/blobs/uploads/{session_id}", dockerV2Handler.HandleImageBlobUpload)

		r.Get("/{namespace}/{repository}/blobs/{digest}", dockerV2Handler.GetImageBlob)
		r.Get("/{repository}/blobs/{digest}", dockerV2Handler.GetImageBlob)

		r.Head("/{namespace}/{repository}/manifests/{tag_or_digest}", dockerV2Handler.CheckImageManifestExistence)
		r.Head("/{repository}/manifests/{tag_or_digest}", dockerV2Handler.CheckImageManifestExistence)

		r.Put("/{namespace}/{repository}/manifests/{tag}", dockerV2Handler.UpdateManifest)
		r.Put("/{repository}/manifests/{tag}", dockerV2Handler.UpdateManifest)

		r.Get("/{namespace}/{repository}/manifests/{tag_or_digest}", dockerV2Handler.GetImageManifest)
		r.Get("/{repository}/manifests/{tag_or_digest}", dockerV2Handler.GetImageManifest)
	})

	return router
}