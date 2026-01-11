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

	"github.com/ksankeerth/open-image-registry/access/resource"
	"github.com/ksankeerth/open-image-registry/client/email"
	"github.com/ksankeerth/open-image-registry/config"
	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/lib"
	"github.com/ksankeerth/open-image-registry/listeners"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/registry"
	"github.com/ksankeerth/open-image-registry/rest"
	"github.com/ksankeerth/open-image-registry/security"
	"github.com/ksankeerth/open-image-registry/storage"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/store/sqlite"
)

func main() {

	log.Logger().Info().Msg("Starting OpenImageRegistry ...........")

	// ------------- parse flags and options ---------------
	appHomeDir := flag.String("app_home", "", "Path to app home directory")
	flag.Parse()

	if *appHomeDir == "" {
		log.Logger().Warn().Msg("app_home is not provided with binary.")
	}

	binaryPath, err := os.Executable()
	if err != nil {
		log.Logger().Fatal().Err(err).Msg("Server startup failed due to filesystem errors.")
		return
	}
	actualBinaryPath, err := filepath.EvalSymlinks(binaryPath)
	if err != nil {
		log.Logger().Fatal().Err(err).Msg("Server startup failed due to filesystem errors.")
		return
	}

	*appHomeDir = filepath.Dir(filepath.Dir(filepath.Dir(actualBinaryPath)))

	log.Logger().Info().Msgf("OpenImageRegistry Home Directory: %s", *appHomeDir)

	serverHomeDir := filepath.Join(*appHomeDir, "server")

	// ------------- initialize configuration -----------------
	configFileLocation := filepath.Join(serverHomeDir, "config.yaml")
	log.Logger().Info().Msgf("Loading configuration from: %s", configFileLocation)
	appConfig, err := config.LoadConfig(configFileLocation, *appHomeDir)
	if err != nil {
		log.Logger().Fatal().Err(err).Msg("Initializing configuration failed.")
		return
	}

	if appConfig.Development.Enable {
		log.Logger().Info().Msg("Development mode is enabled.")
	}

	// ------------ initialize database and daos ---------------

	store, err := sqlite.New(appConfig.Database)
	if err != nil {
		log.Logger().Fatal().Err(err).Msgf("Server startup failed due to db store initialization errors.")
		return
	}

	// -------------------- Initialize storage ------------------

	err = storage.Init(&appConfig.Storage)
	if err != nil {
		log.Logger().Fatal().Err(err).Msg("Server startup failed due to storage errors")
		return
	}

	// --------------------- Initialize admin user account ---------------
	err = initializeAdminUserAccount(store, &appConfig.Admin)
	if err != nil {
		log.Logger().Fatal().Err(err).Msg("Server startup failed due to admin user account initialization errors")
	}

	// --------------------- Initialize email client ---------------------
	var emailClient *email.EmailClient
	emailTemplatesDir := filepath.Join(serverHomeDir, "templates")

	if appConfig.Development.Enable && appConfig.Development.MockEmail {
		if !appConfig.Notification.Email.Enabled {
			// If development config  and mock email is enabled, We'll set a dummay email config to avoid errors.
			emailClient, err = email.NewEmailClient(config.GetDefaultEmailSenderConfig(), emailTemplatesDir, "todo-logo-url")
			if err != nil {
				log.Logger().Fatal().Err(err).Msg("Server startup failed due to email client intialization errors")
				return
			}
		}
	}

	if appConfig.Notification.Email.Enabled {
		emailClient, err = email.NewEmailClient(&appConfig.Notification.Email, emailTemplatesDir, "todo-logo-url")
		if err != nil {
			log.Logger().Fatal().Err(err).Msg("Server startup failed due to email client intialization errors")
			return
		}
	}

	// ------------- create instance of resource access manager ----------------
	accessManager := resource.NewManager(store)

	// -------------- create jwt provider --------------------------------------
	authConfig := appConfig.Security.AuthToken
	jwtAuth := lib.NewOAuthEC256JWTAuthenticator(authConfig.GetPrivateKey(), authConfig.GetPublicKey(), authConfig.Issuer,
		time.Duration(authConfig.Expiry)*time.Second)

	// ------------ start serving ManagementAPIs and UI -----------------------
	appRouter := rest.AppRouter(&appConfig.WebApp, store, jwtAuth, accessManager, emailClient)

	address := fmt.Sprintf("%s:%d", appConfig.Server.Hostname, appConfig.Server.Port)

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

	if appConfig.WebApp.EnableUI {
		log.Logger().Info().Msgf("Serving UI from: %s", appConfig.WebApp.DistPath)
	}

	go startRegistryListeners(appConfig.ImageRegistry.Enabled, appConfig.ImageRegistry.Port, store)

	<-shutdown

	log.Logger().Info().Msg("Server is about to shutdown.")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = server.Shutdown(ctx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occured when shutting down server")
	}
}

func startRegistryListeners(localRegistryEnabled bool, localRegistryPort uint, store store.Store) {
	lm := listeners.GetListenerManager()

	if localRegistryEnabled {
		err := lm.RegisterListener(constants.HostedRegistryID, constants.HostedRegistryName, localRegistryPort,
			registry.NewRegistryHandler(constants.HostedRegistryID, constants.HostedRegistryName, store).Routes(),
			time.Second*10)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Unable to start listener for LocalRegistry")
			return
		}
	}

	upstreamAddrs, err := store.Upstreams().GetAllUpstreamRegistryAddresses(context.Background())
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error ocurred when loading active upstream registeries' address")
		return
	}

	for _, upstreamAddr := range upstreamAddrs {
		err = lm.RegisterListener(upstreamAddr.ID, upstreamAddr.Name, uint(upstreamAddr.Port),
			registry.NewRegistryHandler(upstreamAddr.ID, upstreamAddr.Name, store).Routes(), time.Second*10)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Unable to start listener for %s", upstreamAddr.Name)
			continue
		}
	}
}

func initializeAdminUserAccount(s store.Store, adminConfig *config.AdminUserAccountConfig) error {
	ctx := context.Background()
	tx, err := s.Begin(ctx)
	if err != nil {
		return err
	}
	ctx = store.WithTxContext(ctx, tx)

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	userAcc, err := s.Users().GetByUsername(ctx, adminConfig.Username)
	if err != nil {
		return err
	}
	if userAcc != nil && userAcc.Locked {
		err = s.Users().UnlockAccount(ctx, adminConfig.Username)
		if err != nil {
			return err
		}
	}
	if userAcc == nil && adminConfig.CreateAccount {
		salt, err := security.GenerateSalt(16)
		if err != nil {
			return err
		}
		passwordHash := security.GeneratePasswordHash(adminConfig.Password, salt)
		userId, err := s.Users().Create(ctx, adminConfig.Username, adminConfig.Email, adminConfig.Username, passwordHash, salt)
		if err != nil {
			return err
		}
		err = s.Users().UnlockAccount(ctx, adminConfig.Username)
		if err != nil {
			return err
		}

		err = s.Users().AssignRole(ctx, userId, constants.RoleAdmin)
		if err != nil {
			return err
		}
	}

	return nil
}