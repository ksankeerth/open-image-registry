package types

import "time"

type UpstreamOCIRegEntity struct {
	Id          string    `json:"id,omitempty"`
	Name        string    `json:"name"`
	Url         string    `json:"url"`
	Port        int       `json:"port"`
	Status      string    `json:"status,omitempty"`
	UpstreamUrl string    `json:"upstream_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UpstreamOCIRegAuthConfig defines auth config to access upstream registry.
type UpstreamOCIRegAuthConfig struct {
	// AuthType defines authentication methods. Possible values: `anonymous`, `basic`, `bearer`
	AuthType string `json:"auth_type"`
	// CredentialJson contains required data for the defined `AuthType`. eg:
	// 1. { grant_type:'client-credentials', basic_auth_header: ''}
	// 2. { username: 'admin', password: 'admin }
	// 3. {api_key: 'value' }
	CredentialJson map[string]interface{} `json:"credentials_json,omitempty"`
	// If `AuthType` is `oauth2`, the `TokenEndpoint` defines url to get the token.
	TokenEndpoint string `json:"token_endpoint"`
	// Certificate hold RSA certificate if the `AuthType` is mtls.
	Certificate string `json:"certificate"`
	// UpdatedAt is read only field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type UpstreamOCIRegAccessConfig struct {
	ProxyEnabled               bool      `json:"proxy_enabled"`
	ProxyUrl                   string    `json:"proxy_url,omitempty"`
	ConnectionTimeoutInSeconds int       `json:"connection_timeout"`
	ReadTimeoutInSeconds       int       `json:"read_timeout"`
	MaxConnections             int       `json:"max_connections"`
	MaxRetries                 int       `json:"max_retries"`
	RetryDelayInSeconds        int       `json:"retry_delay"`
	UpdatedAt                  time.Time `json:"updated_at"`
}

type UpstreamOCIRegStorageConfig struct {
	StorageLimitInMbs float32   `json:"storage_limit"`
	CleanupPolicy     string    `json:"cleanup_policy"`
	CleanupThreshold  float32   `json:"cleanup_threshold"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type UpstreamOCIRegCacheConfig struct {
	TtlInSeconds int       `json:"ttl_seconds"`
	OfflineMode  bool      `json:"offline_mode"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// -------------------------------- types of request and response --------------------------------------

type CreateUpstreamRegRequestMsg struct {
	UpstreamOCIRegEntity
	AuthConfig    UpstreamOCIRegAuthConfig    `json:"auth_config"`
	AccessConfig  UpstreamOCIRegAccessConfig  `json:"access_config"`
	StorageConfig UpstreamOCIRegStorageConfig `json:"storage_config"`
	CacheConfig   UpstreamOCIRegCacheConfig   `json:"cache_config"`
}

type UpdateUpstreamRegRequestMsg = CreateUpstreamRegRequestMsg

type UpstreamOCIRegEntityWithAdditionalInfo struct {
	UpstreamOCIRegEntity
	CachedImagesCount int `json:"cached_images_count"`
}

type ListUpstreamRegistriesResponseMsg struct {
	Total       int
	Page        int
	Limit       int
	Registeries []*UpstreamOCIRegEntityWithAdditionalInfo
}

type UpstreamOCIRegResMsg struct {
	UpstreamOCIRegEntity
	AuthConfig    UpstreamOCIRegAuthConfig    `json:"auth_config"`
	AccessConfig  UpstreamOCIRegAccessConfig  `json:"access_config"`
	StorageConfig UpstreamOCIRegStorageConfig `json:"storage_config"`
	CacheConfig   UpstreamOCIRegCacheConfig   `json:"cache_config"`
}
