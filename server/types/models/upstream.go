package models

import "time"

type UpstreamRegistryEntity struct {
	Id          string    `json:"id,omitempty"`
	Name        string    `json:"name"`
	Port        int       `json:"port"`
	Status      string    `json:"status,omitempty"`
	UpstreamUrl string    `json:"upstream_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type UpstreamRegistryAuthConfig struct {
	// AuthType defines authentication methods. Possible values: `anonymous`, `basic`, `bearer`
	AuthType string `json:"auth_type"`
	// CredentialJson contains required data for the defined `AuthType`. eg:
	// 1. { grant_type:'client-credentials', basic_auth_header: ''}
	// 2. { username: 'admin', password: 'admin }
	// 3. {api_key: 'value' }
	CredentialJson map[string]interface{} `json:"credentials_json,omitempty"`
	// If `AuthType` is `oauth2`, the `TokenEndpoint` defines url to get the token.
	TokenEndpoint string `json:"token_endpoint"`

	// UpdatedAt is read only field.
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type UpstreamRegistryAccessConfig struct {
	ProxyEnabled               bool      `json:"proxy_enabled"`
	ProxyUrl                   string    `json:"proxy_url,omitempty"`
	ConnectionTimeoutInSeconds int       `json:"connection_timeout"`
	ReadTimeoutInSeconds       int       `json:"read_timeout"`
	MaxConnections             int       `json:"max_connections"`
	MaxRetries                 int       `json:"max_retries"`
	RetryDelayInSeconds        int       `json:"retry_delay"`
	UpdatedAt                  *time.Time `json:"updated_at"`
}

type UpstreamRegistryStorageConfig struct {
	StorageLimitInMbs float32   `json:"storage_limit"`
	CleanupPolicy     string    `json:"cleanup_policy"`
	CleanupThreshold  float32   `json:"cleanup_threshold"`
	UpdatedAt         *time.Time `json:"updated_at"`
}

type UpstreamRegistryCacheConfig struct {
	Enabled      bool      `json:"enabled"`
	TtlInSeconds int       `json:"ttl_seconds"`
	OfflineMode  bool      `json:"offline_mode"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

type UpstreamRegistrySummary struct {
	UpstreamRegistryEntity
	CachedImagesCount int
}

type UpstreamRegistryAddress struct {
	Id          string
	Name        string
	Port        int
	UpstreamUrl string
}