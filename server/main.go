package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ksankeerth/docker-registry/routes"
)

func main() {
	fmt.Println("Starting Docker Registry .....")

	webappBuildPath := flag.String("webapp-build-path", "./../webapp/build", "Path to the built web app")
	flag.Parse()

	fmt.Println("Webapp build path:", *webappBuildPath)

	router := routes.InitRouter(*webappBuildPath)

	address := fmt.Sprintf(":%d", 8000)

	server := &http.Server{
		Addr:    address,
		Handler: router,
	}

	go func(server *http.Server) {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Server stopped due to errors: %v \n", err)
		}
	}(server)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown

	fmt.Println("Server is about to shoutdown.")
}
