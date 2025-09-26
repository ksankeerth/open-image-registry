package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	server "github.com/ksankeerth/open-image-registry"
	"github.com/ksankeerth/open-image-registry/db"
	"github.com/ksankeerth/open-image-registry/listeners"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/registry"
	"github.com/ksankeerth/open-image-registry/storage"
	"github.com/ksankeerth/open-image-registry/upstream"
)

func main() {

	log.Logger().Info().Msg("Starting OpenImageRegistry ...........")
	// ------------- parse flags and options ---------------
	webappBuildPath := flag.String("webapp-build-path", "./../webapp/build", "Path to the built web app")
	flag.Parse()

	// ------------- initialize configuration -----------------
	// TODO

	// ------------ initialize database and daos ---------------
	database, tm, err := db.InitDB()
	if err != nil {
		log.Logger().Fatal().Err(err).Msgf("Server startup failed due to database initialization errors.")
		return
	}

	upstreamDao := db.NewUpstreamDAO(database, tm)
	imageRegistryDao := db.NewImageRegistryDAO(database, tm)

	// -------------------- Initialize storage ------------------
	cwd, err := os.Getwd()
	if err != nil {
		log.Logger().Fatal().Err(err).Msg("Server startup failed due to file-system errors")
	}

	fsStoragePath := filepath.Join(filepath.Dir(cwd), "/temp/open-image-registry")
	err = storage.Init(fsStoragePath)
	if err != nil {
		log.Logger().Fatal().Err(err).Msg("Server startup failed due to storage errors")
	}

	// ------------ start serving ManagementAPIs and UI -----------------------
	appRouter := server.AppRouter(*webappBuildPath, upstream.NewUpstreamRegistryHandler(upstreamDao))

	address := fmt.Sprintf(":%d", 8000)

	server := &http.Server{
		Addr:    address,
		Handler: appRouter,
	}

	go func(server *http.Server) {
		log.Logger().Info().Msgf("Server started on: %s\n", address)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Logger().Error().Err(err).Msgf("Server stopped due to errors")
		}
	}(server)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	log.Logger().Info().Msgf("Serving UI from: %s", *webappBuildPath)

	// TODO: Avoid hard-coded values and read from configuration
	go startRegistryListeners(true, 5000, imageRegistryDao, upstreamDao)

	<-shutdown

	log.Logger().Info().Msg("Server is about to shutdown.")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = server.Shutdown(ctx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occured when shutting down server")
	}
}

func startRegistryListeners(localRegistryEnabled bool, localRegistryPort uint,
	registryDao db.ImageRegistryDAO, upstreamDao db.UpstreamDAO) {
	lm := listeners.GetListenerManager()

	if localRegistryEnabled {
		err := lm.RegisterListener(registry.HostedRegistryID, registry.HostedRegistryName, localRegistryPort,
			registry.NewRegistryHandler(registry.HostedRegistryID, registry.HostedRegistryName, registryDao, upstreamDao).Routes(),
			time.Second*10)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Unable to start listener for LocalRegistry")
			return
		}
	}

	upstreamAddrs, err := upstreamDao.GetActiveUpstreamAddresses()
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error ocurred when loading active upstream registeries' address")
		return
	}

	for _, upstreamAddr := range upstreamAddrs {
		err = lm.RegisterListener(upstreamAddr.Id, upstreamAddr.Name, uint(upstreamAddr.Port),
			registry.NewRegistryHandler(upstreamAddr.Id, upstreamAddr.Name, registryDao, upstreamDao).Routes(), time.Second*10)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Unable to start listener for %s", upstreamAddr.Name)
			continue
		}
	}
}