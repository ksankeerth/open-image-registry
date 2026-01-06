package rest

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	"github.com/ksankeerth/open-image-registry/access"
	"github.com/ksankeerth/open-image-registry/access/resource"
	"github.com/ksankeerth/open-image-registry/auth"
	"github.com/ksankeerth/open-image-registry/client/email"
	"github.com/ksankeerth/open-image-registry/config"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/user"
)

func AppRouter(webappConfig *config.WebAppConfig, store store.Store, accessManager *resource.Manager, ec *email.EmailClient) *chi.Mux {
	router := chi.NewRouter()

	// Middleware setup
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
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

	router.Use(EnforceJSON)

	authHandler := auth.NewAuthAPIHandler(store)
	userHandler := user.NewUserAPIHandler(store, ec)
	registryAccessHandler := access.NewRegistryAccessHandler(store, accessManager)

	// API routes
	router.Route("/api/v1", func(r chi.Router) {
		r.Mount("/users", userHandler.Routes())
		r.Mount("/auth", authHandler.Routes())
		r.Mount("/access", registryAccessHandler.Routes())
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			//TODO: develop health check endpoint later
			w.WriteHeader(http.StatusOK)
		})
		// Add other API routes here
	})

	if webappConfig == nil && !webappConfig.EnableUI {
		return router
	}

	// WebUI
	absUiBuildPath, err := filepath.Abs(webappConfig.DistPath)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to get the abousolute path for : %s", webappConfig.DistPath)
		panic(err.Error())
	}
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		// Only serve WebUI for non-API requests
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		cleanPath := filepath.Clean(r.URL.Path)
		requestedPath := filepath.Join(webappConfig.DistPath, cleanPath)

		absRequestedPath, err := filepath.Abs(requestedPath)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Unable to get absolute path for: %s", absRequestedPath)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if absRequestedPath == absUiBuildPath {
			http.ServeFile(w, r, filepath.Join(webappConfig.DistPath, "index.html"))
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
			http.ServeFile(w, r, filepath.Join(webappConfig.DistPath, "index.html"))
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