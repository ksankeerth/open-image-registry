package models

import "time"

type UpstreamRegistry struct {
	ID          string
	Name        string
	Description string
	Vendor      string
	State       string
	Port        uint
	UpstreamURL string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type UpstreamRegistryAuthConfig struct {
	RegistryID string
	AuthType   string
	ConfigJSON []byte
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}

type UpstreamRegistryCacheStoreConfig struct {
	RegistryID       string
	CacheEnabled     bool
	TTLSeconds       int
	StorageLimit     float32
	CleanupThreshold float32
	CreatedAt        time.Time
	UpdatedAt        *time.Time
}

type UpstreamRegistryNetworkConfig struct {
	RegistryID             string
	ConnectionTimeout      int
	ReadTimeout            int
	WriteTimeout           int
	MaxConnections         int
	MaxIdleConnections     int
	MaxRetries             int
	RetryDelay             int
	RetryBackOffMultiplier float32
	CreatedAt              time.Time
	UpdatedAt              *time.Time
}
