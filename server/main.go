package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ksankeerth/open-image-registry/db"
	"github.com/ksankeerth/open-image-registry/listeners"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/routes"
	"github.com/ksankeerth/open-image-registry/storage"
)

const HostedRegistryId = "1"

func main() {
	log.Logger().Info().Msg("Starting Docker Registry .....")

	webappBuildPath := flag.String("webapp-build-path", "./../webapp/build", "Path to the built web app")
	flag.Parse()

	log.Logger().Info().Msgf("Webapp build path: %s", *webappBuildPath)

	_, err := db.InitDB()
	if err != nil {
		log.Logger().Fatal().Err(err).Msgf("Server startup failed due to database errors.")
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Logger().Fatal().Err(err).Msg("Unable to get current working directory")
	}

	fsStoragePath := cwd + "/temp/artispace"
	err = storage.Init(fsStoragePath)
	if err != nil {
		log.Logger().Fatal().Err(err).Msgf("Unable to intialize storage at location: %s", fsStoragePath)
	}

	router := routes.InitRouter(*webappBuildPath)

	address := fmt.Sprintf(":%d", 8000)

	server := &http.Server{
		Addr:    address,
		Handler: router,
	}

	go func(server *http.Server) {
		log.Logger().Info().Msgf("Server is listening on port: %d\n", 8000)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Logger().Error().Err(err).Msgf("Server stopped due to errors: %v \n", err)
		}
	}(server)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go startRegistryListeners(5000, true)

	<-shutdown

	log.Logger().Info().Msg("Server is about to shoutdown.")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = server.Shutdown(ctx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occured when shutting down server")
	}
}

func startRegistryListeners(hostedRegPort int, enableHostedReg bool) bool {
	lm := listeners.GetListenerManager()

	if enableHostedReg {

		err := lm.RegisterListener(HostedRegistryId, "HostedRegistry", uint(hostedRegPort), routes.GetDockerV2Router(HostedRegistryId, "HostedRegistry"), time.Second*10)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Unable to start listener for HostedRegistry")
			return false
		}
	}

	upstreamAddrs, err := db.LoadUpstreamRegistryAddresses()
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to start listeners for upstream registeries")
		return false
	}

	for _, upstreamAddr := range upstreamAddrs {
		err := lm.RegisterListener(upstreamAddr.Id, upstreamAddr.Name, uint(upstreamAddr.Port), routes.GetDockerV2Router(HostedRegistryId, upstreamAddr.Name), time.Second*10)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Unable to start listener for %s", upstreamAddr.Name)
			return false
		}
	}
	return true
}
