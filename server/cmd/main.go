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
	"github.com/ksankeerth/open-image-registry/auth"
	"github.com/ksankeerth/open-image-registry/client/email"
	"github.com/ksankeerth/open-image-registry/config"
	"github.com/ksankeerth/open-image-registry/db"
	"github.com/ksankeerth/open-image-registry/listeners"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/registry"
	"github.com/ksankeerth/open-image-registry/security"
	"github.com/ksankeerth/open-image-registry/storage"
	"github.com/ksankeerth/open-image-registry/upstream"
	"github.com/ksankeerth/open-image-registry/user"
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
	database, tm, err := db.InitDB(&appConfig.Database)
	if err != nil {
		log.Logger().Fatal().Err(err).Msgf("Server startup failed due to database initialization errors.")
		return
	}

	upstreamDao := db.NewUpstreamDAO(database, tm)
	imageRegistryDao := db.NewImageRegistryDAO(database, tm)
	userDao := db.NewUserDAO(database, tm)
	resourceAccessDao := db.NewResourceAccessDAO(database, tm)
	oauthDao := db.NewOauthDAO(database, tm)

	// -------------------- Initialize storage ------------------

	err = storage.Init(&appConfig.Storage)
	if err != nil {
		log.Logger().Fatal().Err(err).Msg("Server startup failed due to storage errors")
		return
	}

	// --------------------- Initialize admin user account ---------------
	err = initializeAdminUserAccount(userDao, &appConfig.Admin)
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

	// ------------ start serving ManagementAPIs and UI -----------------------
	appRouter := server.AppRouter(&appConfig.WebApp,
		upstream.NewUpstreamRegistryHandler(upstreamDao),
		auth.NewAuthAPIHandler(userDao, resourceAccessDao, oauthDao),
		user.NewUserAPIHandler(userDao, resourceAccessDao, emailClient),
	)

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

	go startRegistryListeners(appConfig.ImageRegistry.Enabled, appConfig.ImageRegistry.Port, imageRegistryDao, upstreamDao)

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

func initializeAdminUserAccount(userDao db.UserDAO, adminConfig *config.AdminUserAccountConfig) error {
	txKey := fmt.Sprintf("initialize-admin-account-%s", adminConfig.Username)

	err := userDao.Begin(txKey)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			userDao.Rollback(txKey)
		} else {
			userDao.Commit(txKey)
		}
	}()

	userAcc, err := userDao.GetUserAccount(adminConfig.Username, txKey)
	if err != nil {
		return err
	}
	if userAcc != nil && userAcc.Locked {
		_, err = userDao.UnlockUserAccount(adminConfig.Username, txKey)
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
		userId, err := userDao.CreateUserAccount(adminConfig.Username, adminConfig.Email, adminConfig.Username, passwordHash, salt, txKey)
		if err != nil {
			return err
		}
		_, err = userDao.UnlockUserAccount(adminConfig.Username, txKey)
		if err != nil {
			return err
		}

		err = userDao.AssignUserRole(user.RoleAdmin, userId, txKey)
		if err != nil {
			return err
		}
	}

	return nil
}