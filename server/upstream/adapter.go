package upstream

import (
	"time"

	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/types/models"
)

type UpstreamAdapter struct{}

// ToUpstreamEntity converts CreateUpstreamRegistryRequest DTO → UpstreamRegistryEntity (model)
func (ua *UpstreamAdapter) ToUpstreamEntity(req *mgmt.CreateUpstreamRegistryRequest) *models.UpstreamRegistryEntity {
	model := &models.UpstreamRegistryEntity{
		Name:        req.Name,
		Port:        req.Port,
		Status:      req.Status,
		UpstreamUrl: req.UpstreamUrl,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if model.Status == "" {
		model.Status = "active"
	}
	return model
}

// ToUpstreamAuthConfig converts CreateUpstreamRegistryRequest.AuthConfig DTO → UpstreamRegistryAuthConfig (model)
func (ua *UpstreamAdapter) ToUpstreamAuthConfig(req *mgmt.CreateUpstreamRegistryRequest) *models.UpstreamRegistryAuthConfig {
	return &models.UpstreamRegistryAuthConfig{
		AuthType:       req.AuthConfig.AuthType,
		CredentialJson: req.AuthConfig.CredentialJson,
		TokenEndpoint:  req.AuthConfig.TokenEndpoint,
		UpdatedAt:      time.Now(),
	}
}

// ToUpstreamAccessConfig converts DTO → model
func (ua *UpstreamAdapter) ToUpstreamAccessConfig(req *mgmt.CreateUpstreamRegistryRequest) *models.UpstreamRegistryAccessConfig {
	return &models.UpstreamRegistryAccessConfig{
		ProxyEnabled:               req.AccessConfig.ProxyEnabled,
		ProxyUrl:                   req.AccessConfig.ProxyUrl,
		ConnectionTimeoutInSeconds: req.AccessConfig.ConnectionTimeoutInSeconds,
		ReadTimeoutInSeconds:       req.AccessConfig.ReadTimeoutInSeconds,
		MaxConnections:             req.AccessConfig.MaxConnections,
		MaxRetries:                 req.AccessConfig.MaxRetries,
		RetryDelayInSeconds:        req.AccessConfig.RetryDelayInSeconds,
		UpdatedAt:                  time.Now(),
	}
}

// ToUpstreamStorageConfig converts DTO → model
func (ua *UpstreamAdapter) ToUpstreamStorageConfig(req *mgmt.CreateUpstreamRegistryRequest) *models.UpstreamRegistryStorageConfig {
	return &models.UpstreamRegistryStorageConfig{
		StorageLimitInMbs: req.StorageConfig.StorageLimitInMbs,
		CleanupPolicy:     req.StorageConfig.CleanupPolicy,
		CleanupThreshold:  req.StorageConfig.CleanupThreshold,
		UpdatedAt:         time.Now(),
	}
}

// ToUpstreamCacheConfig converts DTO → model
func (ua *UpstreamAdapter) ToUpstreamCacheConfig(req *mgmt.CreateUpstreamRegistryRequest) *models.UpstreamRegistryCacheConfig {
	return &models.UpstreamRegistryCacheConfig{
		Enabled:      req.CacheConfig.Enabled,
		TtlInSeconds: req.CacheConfig.TtlInSeconds,
		OfflineMode:  req.CacheConfig.OfflineMode,
		UpdatedAt:    time.Now(),
	}
}

func (ua *UpstreamAdapter) ToUpstreamRegistrySummaryDTO(model *models.UpstreamRegistrySummary) *mgmt.UpstreamRegistrySummaryDTO {
	if model == nil {
		return nil
	}

	return &mgmt.UpstreamRegistrySummaryDTO{
		Id:                model.Id,
		Name:              model.Name,
		Port:              model.Port,
		Status:            model.Status,
		UpstreamUrl:       model.UpstreamUrl,
		CachedImagesCount: model.CachedImagesCount,
	}
}