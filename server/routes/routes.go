package routes

import (
	"database/sql"
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

func InitRouter(webappBuildPath string, database *sql.DB) *chi.Mux {

	router := chi.NewRouter()

	// Cors configuration
	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"*"},
		AllowedHeaders: []string{"*"},
	})

	router.Use(cors.Handler)
	router.Use(httplog.RequestLogger(httplog.NewLogger("open-image-registry", httplog.Options{
		LogLevel:         slog.LevelDebug,
		Concise:          true,
		RequestHeaders:   true,
		MessageFieldName: "message",
	})))

	upstreamRegDAO := db.NewUpstreamRegistryDAO(database)
	upstreamRegistryHandler := api.NewUpstreamRegistryHandler(&upstreamRegDAO)

	router.Route("/v2", func(r chi.Router) {})

	router.Route("/api/v1/upstreams", func(r chi.Router) {
		r.Put("/{registry_id}", upstreamRegistryHandler.UpdateUpstreamRegistry)
		r.Get("/{registry_id}", upstreamRegistryHandler.GetUpstreamRegistry)
		r.Delete("/{registry_id}", upstreamRegistryHandler.DeleteUpstreamRegistry)
		r.Post("/", upstreamRegistryHandler.CreateUpstreamRegistry)
		r.Get("/", upstreamRegistryHandler.ListUpstreamRegisteries)
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

func handleManagementAPIs(r chi.Router) {

}
