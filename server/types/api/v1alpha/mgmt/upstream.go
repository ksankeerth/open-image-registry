package mgmt

import "time"

type UpstreamAuthConfigDTO struct {
	// AuthType defines authentication methods. Possible values: `anonymous`, `basic`, `bearer`
	AuthType string `json:"auth_type"`
	// CredentialJson contains required data for the defined `AuthType`. eg:
	// 1. { grant_type:'client-credentials', basic_auth_header: ''}
	// 2. { username: 'admin', password: 'admin }
	// 3. {api_key: 'value' }
	CredentialJson map[string]interface{} `json:"credentials_json,omitempty"`
	// If `AuthType` is `oauth2`, the `TokenEndpoint` defines url to get the token.
	TokenEndpoint string `json:"token_endpoint"`
}

type UpstreamAccessConfigDTO struct {
	ProxyEnabled               bool   `json:"proxy_enabled"`
	ProxyUrl                   string `json:"proxy_url,omitempty"`
	ConnectionTimeoutInSeconds int    `json:"connection_timeout"`
	ReadTimeoutInSeconds       int    `json:"read_timeout"`
	MaxConnections             int    `json:"max_connections"`
	MaxRetries                 int    `json:"max_retries"`
	RetryDelayInSeconds        int    `json:"retry_delay"`
}

type UpstreamStorageConfigDTO struct {
	StorageLimitInMbs float32 `json:"storage_limit"`
	CleanupPolicy     string  `json:"cleanup_policy"`
	CleanupThreshold  float32 `json:"cleanup_threshold"`
}

type UpstreamCacheConfigDTO struct {
	Enabled      bool `json:"enabled"`
	TtlInSeconds int  `json:"ttl_seconds"`
	OfflineMode  bool `json:"offline_mode"`
}

type CreateUpstreamRegistryRequest struct {
	Name        string `json:"name"`
	Port        int    `json:"port"`
	Status      string `json:"status,omitempty"`
	UpstreamUrl string `json:"upstream_url"`

	AuthConfig    UpstreamAuthConfigDTO    `json:"auth_config"`
	AccessConfig  UpstreamAccessConfigDTO  `json:"access_config"`
	StorageConfig UpstreamStorageConfigDTO `json:"storage_config"`
	CacheConfig   UpstreamCacheConfigDTO   `json:"cache_config"`
}

type CreateUpstreamRegistryResponse struct {
	RegId   string `json:"reg_id"`
	RegName string `json:"reg_name"`
}

type UpdateUpstreamRegistryRequest struct {
	RegId string `json:"reg_id"`
	CreateUpstreamRegistryRequest
}

type UpstreamRegistrySummaryDTO struct {
	Id                string `json:"id"`
	Name              string `json:"name"`
	Port              int    `json:"port"`
	Status            string `json:"status,omitempty"`
	UpstreamUrl       string `json:"upstream_url"`
	CachedImagesCount int    `json:"cached_images_count"`
}

type ListUpstreamsResponse struct {
	Total      int                           `json:"total"`
	Page       int                           `json:"page"`
	Limit      int                           `json:"limit"`
	Registries []*UpstreamRegistrySummaryDTO `json:"registries"`
}

type UpstreamAuthConfigResponse struct {
	UpstreamAuthConfigDTO
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpstreamAccessConfigResponse struct {
	UpstreamAccessConfigDTO
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpstreamStorageConfigResponse struct {
	UpstreamStorageConfigDTO
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpstreamCacheConfigResponse struct {
	UpstreamCacheConfigDTO
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpstreamRegistryResponse struct {
	Id            string                        `json:"id"`
	Name          string                        `json:"name"`
	Port          int                           `json:"port"`
	Status        string                        `json:"status,omitempty"`
	UpstreamUrl   string                        `json:"upstream_url"`
	CreatedAt     time.Time                     `json:"created_at"`
	UpdatedAt     time.Time                     `json:"updated_at"`
	AuthConfig    UpstreamAuthConfigResponse    `json:"auth_config"`
	AccessConfig  UpstreamAccessConfigResponse  `json:"access_config"`
	StorageConfig UpstreamStorageConfigResponse `json:"storage_config"`
	CacheConfig   UpstreamCacheConfigResponse   `json:"cache_config"`
}