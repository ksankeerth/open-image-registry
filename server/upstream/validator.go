package upstream

import (
	"crypto/x509"
	"encoding/pem"
	"net/url"
	"strings"

	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/utils"
)

// ValidateCreateRequest validates the request payload for creating/updating an upstream registry
func ValidateCreateRequest(req *mgmt.CreateUpstreamRegistryRequest) (bool, string) {
	if req.Name == "" {
		return false, "Registry Name is not provided"
	}
	if !utils.IsValidRegistry(req.Name) {
		return false, "Registry Name has invalid characters. Allowed characters: a-zA-Z0-9_-"
	}

	if req.Port <= 1024 || req.Port > 65535 {
		return false, "Port must be in range [1025-65535]"
	}

	if req.UpstreamUrl == "" {
		return false, "Upstream Registry URL is not provided"
	}

	// TODO: support Oauth2 later
	if !(req.AuthConfig.AuthType == "anonymous" ||
		req.AuthConfig.AuthType == "basic" ||
		req.AuthConfig.AuthType == "bearer" ||
		req.AuthConfig.AuthType == "mtls") {
		return false, "Unsupported Auth type provided"
	}

	// Auth-specific checks
	switch req.AuthConfig.AuthType {
	case "anonymous":
		if req.AuthConfig.CredentialJson != nil && len(req.AuthConfig.CredentialJson) != 0 {
			return false, "Unnecessary credentials are provided for anonymous access"
		}
	case "basic":
		if req.AuthConfig.CredentialJson == nil ||
			req.AuthConfig.CredentialJson["username"] == "" ||
			req.AuthConfig.CredentialJson["password"] == "" {
			return false, "Username and password are mandatory for Basic auth access"
		}
	case "bearer":
		if req.AuthConfig.CredentialJson == nil ||
			req.AuthConfig.CredentialJson["token"] == "" {
			return false, "Token is mandatory for Bearer auth access"
		}
	}

	// AccessConfig
	if req.AccessConfig.ProxyEnabled && !isValidUrl(req.AccessConfig.ProxyUrl) {
		return false, "Proxy is enabled but proxy server URL is invalid"
	}
	if !isInRange(req.AccessConfig.ConnectionTimeoutInSeconds, 1, 300) {
		return false, "Connection timeout must be in range 1–300 seconds"
	}
	if !isInRange(req.AccessConfig.ReadTimeoutInSeconds, 1, 600) {
		return false, "Read timeout must be in range 1–600 seconds"
	}
	if !isInRange(req.AccessConfig.MaxRetries, 0, 10) {
		return false, "Max retries must be in range 0–10"
	}
	if !isInRange(req.AccessConfig.RetryDelayInSeconds, 1, 60) {
		return false, "Retry delay must be in range 1–60 seconds"
	}

	// StorageConfig
	if !isInRangeFloat(req.StorageConfig.StorageLimitInMbs, 100, 102400) {
		return false, "Storage limit must be between 100MB – 100GB"
	}
	if !isOneOf(req.StorageConfig.CleanupPolicy, []string{"lru_1m", "lru_3m", "lp"}) {
		return false, "Invalid cleanup policy"
	}

	// CacheConfig
	if req.CacheConfig.Enabled && !isInRange(req.CacheConfig.TtlInSeconds, 600, 31536000) {
		return false, "Cache TTL must be between 600 – 31536000 seconds"
	}

	return true, ""
}

//
// --- helper functions ---
//

func isValidCertificate(certificate string) bool {
	certificate = strings.TrimSpace(certificate)

	block, _ := pem.Decode([]byte(certificate))
	if block == nil || block.Type != "CERTIFICATE" {
		return false
	}

	_, err := x509.ParseCertificate(block.Bytes)
	return err == nil
}

func isValidUrl(urlStr string) bool {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	if parsedUrl.Scheme == "" || parsedUrl.Host == "" {
		return false
	}
	return isOneOf(strings.ToLower(parsedUrl.Scheme), []string{"http", "https"})
}

func isInRange(value int, min int, max int) bool {
	return value >= min && value <= max
}

func isInRangeFloat(value float32, min float32, max float32) bool {
	return value >= min && value <= max
}

func isOneOf(value string, allowedValues []string) bool {
	for _, av := range allowedValues {
		if av == value {
			return true
		}
	}
	return false
}