package integration

import (
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/ksankeerth/open-image-registry/access/resource"
	"github.com/ksankeerth/open-image-registry/client/email"
	"github.com/ksankeerth/open-image-registry/config"
	"github.com/ksankeerth/open-image-registry/lib"
	"github.com/ksankeerth/open-image-registry/rest"
	"github.com/ksankeerth/open-image-registry/storage"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/store/sqlite"
	"github.com/ksankeerth/open-image-registry/tests/integration/helpers"
	"github.com/ksankeerth/open-image-registry/tests/integration/seeder"
	v1 "github.com/ksankeerth/open-image-registry/tests/integration/v1"
)

var (
	testServer      *httptest.Server
	testStore       store.Store
	testConfig      *config.AppConfig
	testBaseURL     string
	testEmailClient *email.EmailClient
	tempDir         string
	jwtProvider     lib.JWTProvider
)

func TestMain(m *testing.M) {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ltime | log.Lmsgprefix)
	log.SetPrefix("[TestMain] ")
	log.Println("========================================")
	log.Println("Starting Integration Tests")
	log.Println("========================================")

	log.Println("\n SETUP PHASE")
	if err := setupTestEnvironment(); err != nil {
		// teardownTestEnvironment()
		log.Fatalf("Setup failed: %v", err)

	}
	log.Println("Setup completed successfully")

	exitCode := m.Run()
	log.Println("\n TEARDOWN PHASE")
	if err := teardownTestEnvironment(); err != nil {
		log.Printf(" Teardown errors: %v", err)
	}
	log.Println("Teardown completed")

	// Summary
	log.Println("\n========================================")
	if exitCode == 0 {
		log.Println("All Integration Tests Passed")
	} else {
		log.Printf("Integration Tests Failed (code: %d)", exitCode)
	}
	log.Println("========================================")
	os.Exit(exitCode)
}

func TestIntegrationV1(t *testing.T) {
	seeder := seeder.NewTestDataSeeder(testBaseURL, testStore, jwtProvider)

	suites := []APITestSuite{
		v1.NewUserTestSuite(seeder, testBaseURL),
		v1.NewAuthTestSuite(seeder, testBaseURL),
		v1.NewNamespaceTestSuite(seeder, testBaseURL),
	}

	for _, suite := range suites {
		t.Run(fmt.Sprintf("%s: %s", suite.Name(), suite.APIVersion()), suite.Run)
	}

}

func setupTestEnvironment() error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("unable to detect test location")
	}

	testDir := filepath.Dir(filepath.Dir(filename))
	projectRoot := filepath.Dir(filepath.Dir(testDir))
	tempDir := filepath.Join(testDir, "temp")

	log.Printf("├─ Test directory: %s", testDir)
	log.Printf("├─ Project root: %s", projectRoot)

	log.Printf("├─ Cleaning temporary directories if exists: %s", tempDir)
	os.RemoveAll(tempDir) // we don't care about the errors if not exists

	configFileLocation := filepath.Join(testDir, "testdata", "test_config.yaml")
	log.Printf("├─ Loading config from: %s", configFileLocation)

	appConfig, err := config.LoadConfig(configFileLocation, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	appConfig.Server.Port = helpers.FindFreePort()
	appConfig.Development.Enable = true
	appConfig.Development.MockEmail = true

	testConfig = appConfig
	log.Printf("├─ Test server will run on port: %d", appConfig.Server.Port)

	log.Println("├─ Initializing database...")
	log.Println("DB file path = " + appConfig.Database.Path)
	store, err := sqlite.New(appConfig.Database)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	testStore = store
	log.Println("├─ Database initialized")

	log.Println("├─ Initializing storage...")
	if err := storage.Init(&appConfig.Storage); err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	log.Println("├─ Storage initialized")

	log.Println("├─ Creating admin user account...")
	if err := helpers.InitializeAdminUserAccount(store, &appConfig.Admin); err != nil {
		return fmt.Errorf("failed to create admin account: %w", err)
	}
	log.Printf("├─ Admin user created: %s", appConfig.Admin.Username)

	emailTemplatesDir := filepath.Join(projectRoot, "server", "templates")
	if appConfig.Development.MockEmail {
		emailClient, err := email.NewEmailClient(
			config.GetDefaultEmailSenderConfig(),
			emailTemplatesDir,
			"test-logo-url",
		)
		if err != nil {
			return fmt.Errorf("failed to create email client: %w", err)
		}
		testEmailClient = emailClient
		log.Println("├─ Email client initialized (mock)")
	}

	// Creating Resource Access Manager
	accessManager := resource.NewManager(store)

	authConfig := appConfig.Security.AuthToken
	jwtAuth := lib.NewOAuthEC256JWTAuthenticator(authConfig.GetPrivateKey(), authConfig.GetPublicKey(), authConfig.Issuer,
		time.Duration(authConfig.Expiry)*time.Second)

	jwtProvider = jwtAuth

	log.Println("├─ Creating HTTP server...")
	appRouter := rest.AppRouter(&appConfig.WebApp, store, jwtAuth, accessManager, testEmailClient)

	testServer = httptest.NewServer(appRouter)
	testBaseURL = testServer.URL

	if err := helpers.WaitForServer(testBaseURL, 10*time.Second); err != nil {
		return fmt.Errorf("server failed to start: %w", err)
	}
	log.Printf("└─ Server ready at: %s", testBaseURL)

	return nil
}

func teardownTestEnvironment() error {
	log.Println("├─ Shutting down HTTP server...")

	if testServer != nil {
		if testServer != nil {
			testServer.Close()
		}
	}

	if testConfig == nil {
		return nil
	}
	log.Println("└─ Cleaning up temporary files...")
	err := os.RemoveAll(filepath.Dir(testConfig.Storage.Path))
	if err != nil {
		log.Printf("├─ Removing temporary files(%s) failed: %v", tempDir, err)
	} else {
		log.Println("├─ Removed temporary files")
	}

	return nil
}