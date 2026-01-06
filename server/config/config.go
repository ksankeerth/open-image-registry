package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/utils"
	"gopkg.in/yaml.v3"
)

var appConfiguration *AppConfig

type AppConfig struct {
	Server           MgmtServerConfig       `yaml:"server"`
	ImageRegistry    ImageRegistryConfig    `yaml:"image_registry"`
	UpstreamRegistry UpstreamRegistryConfig `yaml:"upstream_registry"`
	Admin            AdminUserAccountConfig `yaml:"admin"`
	Database         DatabaseConfig         `yaml:"database"`
	Storage          StorageConfig          `yaml:"storage"`
	Notification     NotificationConfig     `yaml:"notification"`
	WebApp           WebAppConfig           `yaml:"webapp"`
	Audit            AuditEventsConfig      `yaml:"audit"`
	Development      DevelopmentConfig      `yaml:"development"`
	Testing          TestingConfig          `yaml:"testing"`
}

func LoadConfig(configPath, appHome string) (*AppConfig, error) {
	appConfig := defaultConfig(filepath.Join(appHome, "server"))

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &appConfig); err != nil {
		return nil, err
	}

	resolveAppHomeVars(appConfig, appHome)

	valid, errMsg := validateConfig(appConfig)
	if !valid {
		return nil, fmt.Errorf("invalid config: %s", errMsg)
	}

	appConfiguration = appConfig
	return appConfig, nil
}

func resolveAppHomeVars(cfg *AppConfig, appHome string) {
	replace := func(path string) string {
		return filepath.Clean(os.Expand(path, func(key string) string {
			if key == "app_home" {
				return appHome
			}
			return ""
		}))
	}

	cfg.Database.Path = replace(cfg.Database.Path)
	cfg.Database.ScriptsPath = replace(cfg.Database.ScriptsPath)
	cfg.Storage.Path = replace(cfg.Storage.Path)
	cfg.WebApp.DistPath = replace(cfg.WebApp.DistPath)
}

func validateConfig(cfg *AppConfig) (bool, string) {
	// --- Server Validation ---
	if cfg.Server.Hostname == "" {
		return false, "server.hostname cannot be empty"
	}
	if cfg.Server.Port == 0 {
		return false, "server.port must be greater than 0"
	}

	// --- Image Registry ---
	if cfg.ImageRegistry.Enabled {
		if cfg.ImageRegistry.Hostname == "" {
			return false, "image_registry.hostname cannot be empty when image_registry.enabled = true"
		}
		if cfg.ImageRegistry.Port == 0 {
			return false, "image_registry.port must be greater than 0 when image_registry.enabled = true"
		}
	}

	// --- Admin Account ---
	if cfg.Admin.CreateAccount {
		if cfg.Admin.Username == "" {
			return false, "admin.username cannot be empty"
		}
		if cfg.Admin.Password == "" {
			return false, "admin.password cannot be empty"
		}
		if cfg.Admin.Email == "" {
			return false, "admin.email cannot be empty"
		}
		if !utils.IsValidEmail(cfg.Admin.Email) {
			return false, "invalid email provided for admin.email"
		}
	}

	// --- Database ---
	if cfg.Database.Type == "" {
		return false, "database.type cannot be empty"
	}
	if cfg.Database.Path == "" && !cfg.Database.AutoCreate {
		return false, "database.path cannot be empty"
	}
	if _, err := os.Stat(filepath.Dir(cfg.Database.Path)); err != nil {
		if cfg.Database.AutoCreate {
			dbDir := filepath.Dir(cfg.Database.Path)
			err = os.MkdirAll(dbDir, 0755)
			if err != nil {
				return false, fmt.Sprintf("unable to create database path: %s due to errors: %v", dbDir, err)
			}
		} else {
			return false, fmt.Sprintf("database path directory does not exist: %s", filepath.Dir(cfg.Database.Path))
		}
	}

	// --- Storage ---
	if cfg.Storage.Type == "" {
		return false, "storage.type cannot be empty"
	}
	if cfg.Storage.Path == "" && !cfg.Storage.AutoCreate {
		return false, "storage.path cannot be empty"
	}
	_, err := os.Stat(cfg.Storage.Path)
	if err != nil && cfg.Storage.AutoCreate {
		err1 := os.MkdirAll(cfg.Storage.Path, 0755)
		if err1 != nil {
			return false, fmt.Sprintf("unable to create storage directory: %s due to errors: %v", cfg.Storage.Path, err1)
		}
	} else if err != nil {
		return false, fmt.Sprintf("storage path does not exist: %s", cfg.Storage.Path)
	}

	// --- WebApp ---
	if cfg.WebApp.EnableUI {
		if cfg.WebApp.DistPath == "" {
			return false, "webapp.dist_path cannot be empty when enable_ui = true"
		}
		if _, err := os.Stat(cfg.WebApp.DistPath); err != nil {
			return false, fmt.Sprintf("webapp.dist_path does not exist: %s", cfg.WebApp.DistPath)
		}
	}

	// --- Notification / Email ---
	if cfg.Notification.Email.Enabled {
		emailCfg := cfg.Notification.Email
		if emailCfg.SmtpHost == "" {
			return false, "notification.email.smtp_host cannot be empty when email.enabled = true"
		}
		if emailCfg.SmtpPort == 0 {
			return false, "notification.email.smtp_port must be greater than 0 when email.enabled = true"
		}
		if emailCfg.FromAddress == "" {
			return false, "notification.email.from_address cannot be empty when email.enabled = true"
		}
	}
	return true, ""
}

func defaultConfig(severHome string) *AppConfig {
	return &AppConfig{
		Server: MgmtServerConfig{
			Hostname: "localhost",
			Port:     8000,
		},
		ImageRegistry: ImageRegistryConfig{
			Enabled:  true,
			Hostname: "localhost",
			Port:     5000,
		},
		UpstreamRegistry: UpstreamRegistryConfig{
			Enabled: true,
		},
		Admin: AdminUserAccountConfig{
			Username:      "admin",
			Password:      "admin",
			CreateAccount: true,
		},
		Database: DatabaseConfig{
			Type:       "sqlite",
			AutoCreate: true,
			Path:       filepath.Join(severHome, "registry_sqlite.db"),
		},
		Storage: StorageConfig{
			Type: "lfs",
			Path: filepath.Join(severHome, "temp"),
		},
		Notification: NotificationConfig{
			Email: EmailSenderConfig{
				Enabled: false,
			},
		},
		Audit: AuditEventsConfig{
			Enable:               true,
			VerifyBucketsOnStart: true,
			RecordsPerBucket:     constants.DefaultAuditBucketLimit,
			AvailableBucketCount: constants.DefaultAuditSqliteBuckets,
			BatchInsertSize:      constants.DefaultAuditBatchSize,
			BatchInsertWaitTime:  constants.DefaultAuditBatchWaitTime,
		},
	}
}

type MgmtServerConfig struct {
	Hostname string `yaml:"hostname"`
	Port     uint   `yaml:"port"`
}

type ImageRegistryConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Hostname string `yaml:"hostname"`
	Port     uint   `yaml:"port"`
	// if this is true, it allows developers to create namespace on docker push
	CreateNamespaceOnPush bool `yaml:"create_namespace_on_push"`
	// if this is true, it allows developers to create repository on docker push
	CreateRepositoryOnPush bool `yaml:"create_repository_on_push"`
}

type UpstreamRegistryConfig struct {
	Enabled bool `yaml:"enabled"`
}

type AdminUserAccountConfig struct {
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
	Email         string `yaml:"email"`
	CreateAccount bool   `yaml:"create_account"`
}

type DatabaseConfig struct {
	Type        string `yaml:"type"`
	Path        string `yaml:"path"`
	AutoCreate  bool   `yaml:"auto_create"`
	ScriptsPath string `yaml:"scripts_path"`
}

type StorageConfig struct {
	Type       string `yaml:"type"`
	Path       string `yaml:"path"`
	AutoCreate bool   `yaml:"auto_create"`
}

type NotificationConfig struct {
	Email EmailSenderConfig `yaml:"email"`
}

type EmailSenderConfig struct {
	Enabled      bool   `yaml:"enabled"`
	SmtpHost     string `yaml:"smtp_host"`
	SmtpPort     uint   `yaml:"smtp_port"`
	SmtpUser     string `yaml:"smtp_user"`
	SmtpPassword string `yaml:"smtp_password"`
	FromAddress  string `yaml:"from_address"`
}

type WebAppConfig struct {
	EnableUI bool   `yaml:"enable_ui"`
	DistPath string `yaml:"dist_path"`
}

type DevelopmentConfig struct {
	Enable    bool `yaml:"enable"`
	MockEmail bool `yaml:"mock_email"`
}

type TestingConfig struct {
	AllowDeleteAll bool `yaml:"allow_delete_all"`
}

type AuditEventsConfig struct {
	Enable               bool          `yaml:"enable"`
	VerifyBucketsOnStart bool          `yaml:"verify_buckets_on_start"`
	RecordsPerBucket     uint64        `yaml:"records_per_bucket"`
	AvailableBucketCount uint16        `yaml:"available_bucket_count"`
	BatchInsertSize      uint32        `yaml:"batch_insert_size"`
	BatchInsertWaitTime  time.Duration `yaml:"batch_insert_wait_time"`
	// TODO: for now, by default we persist audit events in database and logs.
	// later consider supporting multiple types
}

func GetDevelopmentConfig() DevelopmentConfig {
	if appConfiguration == nil {
		return DevelopmentConfig{}
	}
	return appConfiguration.Development
}

func GetTestingConfig() TestingConfig {
	if appConfiguration == nil {
		return TestingConfig{}
	}
	return appConfiguration.Testing
}

func GetImageRegistryConfig() ImageRegistryConfig {
	if appConfiguration == nil {
		return ImageRegistryConfig{}
	}

	return appConfiguration.ImageRegistry
}

func GetDefaultEmailSenderConfig() *EmailSenderConfig {
	return &EmailSenderConfig{
		Enabled:      false,
		SmtpHost:     "localhost",
		SmtpPort:     25,
		SmtpUser:     "",
		SmtpPassword: "",
		FromAddress:  "noreply@test.com",
	}
}