package server

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/upstream"
)

func AppRouter(uiBuildPath string, upstreamHandler *upstream.UpstreamRegistryHandler) *chi.Mux {
	router := chi.NewRouter()

	// Middleware setup
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	router.Use(corsMiddleware.Handler)
	router.Use(httplog.RequestLogger(httplog.NewLogger("OIR_API", httplog.Options{
		LogLevel:         slog.LevelDebug,
		Concise:          true,
		RequestHeaders:   true,
		MessageFieldName: "message",
	})))

	// API routes
	router.Route("/api/v1", func(r chi.Router) {
		r.Mount("/upstreams", upstreamHandler.Routes())
		// Add other API routes here
	})

	// WebUI
	absUiBuildPath, err := filepath.Abs(uiBuildPath)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to get the abousolute path for : %s", uiBuildPath)
		panic(err.Error())
	}
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		// Only serve WebUI for non-API requests
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		cleanPath := filepath.Clean(r.URL.Path)
		requestedPath := filepath.Join(uiBuildPath, cleanPath)

		absRequestedPath, err := filepath.Abs(requestedPath)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Unable to get absolute path for: %s", absRequestedPath)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if absRequestedPath == absUiBuildPath {
			http.ServeFile(w, r, filepath.Join(uiBuildPath, "index.html"))
			return
		}

		if !strings.HasPrefix(absRequestedPath, absUiBuildPath+string(filepath.Separator)) {
			log.Logger().Warn().Msgf("Directory traversal attempt blocked; Path: %s", requestedPath)
			http.NotFound(w, r)
			return
		}

		log.Logger().Debug().Msgf("WebUI loads %s", requestedPath)

		_, err = os.Stat(absRequestedPath)
		if os.IsNotExist(err) {
			// fallback to index.html for SPA routes
			http.ServeFile(w, r, filepath.Join(uiBuildPath, "index.html"))
			return
		}
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when checking file: %s", absRequestedPath)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.ServeFile(w, r, absRequestedPath)
	})

	return router
}