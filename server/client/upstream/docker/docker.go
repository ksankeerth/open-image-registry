package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/ksankeerth/open-image-registry/client/upstream"
	"github.com/ksankeerth/open-image-registry/lib"
	"github.com/ksankeerth/open-image-registry/log"
)

const (
	defaultRegistryURL = "https://registry-1.docker.io"
	defaultTokenURL    = "https://auth.docker.io/token"
)

type Config struct {
	RegistryURL string
	TokenURL    string
	Anonymous   bool
	Username    string
	Password    string

	ConnectionTimeout time.Duration
	RequestTimeout    time.Duration

	MaxConnections     int
	MaxIdleConnections int

	MaxRetries             int
	RetryDelay             time.Duration
	RetryBackOffMultiplier float32

	// debug configuration
	LogHeaders bool
	LogBody    bool
}

type AuthConfig struct {
	Username      string `json:"username"`
	Credential    string `json:"credential"` // this could be password or PAT
	TokenEndpoint string `json:"token_endpoint"`
}

type loginResponse struct {
	Token       string    `json:"token"`
	AccessToken string    `json:"access_token"`
	ExpiresIn   int64     `json:"expires_in"`
	IssuedAt    time.Time `json:"issued_at"`
}

type dockerClient struct {
	config     *Config
	tokenCache *lib.Cache

	httpClient *http.Client
}

func NewClient(cfg *Config) upstream.UpstreamClient {
	if cfg == nil {
		cfg = &Config{}
	}

	// Set defaults
	if cfg.RegistryURL == "" {
		cfg.RegistryURL = defaultRegistryURL
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = defaultTokenURL
	}
	if cfg.ConnectionTimeout == 0 {
		cfg.ConnectionTimeout = 10 * time.Second
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 30 * time.Second
	}
	if cfg.MaxConnections == 0 {
		cfg.MaxConnections = 100
	}
	if cfg.MaxIdleConnections == 0 {
		cfg.MaxIdleConnections = 10
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 5 * time.Second
	}
	if cfg.RetryBackOffMultiplier == 0 {
		cfg.RetryBackOffMultiplier = 2.0
	}

	client := &dockerClient{
		config:     cfg,
		tokenCache: lib.NewCache(1 * time.Minute),
		httpClient: &http.Client{
			Timeout: cfg.RequestTimeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   cfg.ConnectionTimeout,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConnsPerHost: cfg.MaxConnections,
				MaxIdleConns:        cfg.MaxIdleConnections,
			},
		},
	}

	log.Logger().Info().Str("registry_url", cfg.RegistryURL).Msg("Docker client initialized")

	return client
}

func (d *dockerClient) GetManifest(namespace, repository, identifier string) (content []byte,
	mediaType string, err error) {

	log.Logger().Debug().
		Str("namespace", namespace).
		Str("repository", repository).
		Str("identifier", identifier).
		Msg("Fetching manifest")

	token, err := d.getToken(namespace, repository, "pull")
	if err != nil {
		log.Logger().Error().Err(err).
			Str("namespace", namespace).
			Str("repository", repository).
			Msg("Failed to get token for manifest fetch")
		return nil, "", fmt.Errorf("failed to get token: %w", err)
	}

	url := fmt.Sprintf("%s/v2/%s/%s/manifests/%s",
		d.config.RegistryURL, namespace, repository, identifier)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to create manifest request to %s", url)
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")
	req.Header.Set("Accept", "application/vnd.oci.image.manifest.v1+json")
	req.Header.Set("Accept", "application/vnd.oci.image.index.v1+json")

	resp, err := d.doWithRetry(req)
	if err != nil {
		log.Logger().Error().Err(err).
			Str("url", url).
			Msg("Failed to fetch manifest from upstream")
		return nil, "", fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Logger().Error().
			Int("status_code", resp.StatusCode).
			Str("url", url).
			Str("response_body", string(body)).
			Msg("Unexpected status code while fetching manifest")
		return nil, "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	content, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Logger().Error().Err(err).
			Str("url", url).
			Msg("Failed to read manifest response body")
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	mediaType = resp.Header.Get("Content-Type")

	if d.config.LogHeaders {
		log.Logger().Debug().Interface("headers", resp.Header).Msg("Manifest response headers")
	}
	if d.config.LogBody {
		log.Logger().Debug().Str("body", string(content)).Msg("Manifest response body")
	}

	log.Logger().Debug().
		Str("namespace", namespace).
		Str("repository", repository).
		Str("identifier", identifier).
		Str("media_type", mediaType).
		Int("size", len(content)).
		Msg("Manifest fetched successfully")

	return content, mediaType, nil
}

func (d *dockerClient) HeadManifest(namespace, repository, identifier string) (exists bool, err error) {
	log.Logger().Debug().
		Str("namespace", namespace).
		Str("repository", repository).
		Str("identifier", identifier).
		Msg("Checking manifest existence")

	token, err := d.getToken(namespace, repository, "pull")
	if err != nil {
		log.Logger().Error().Err(err).
			Str("namespace", namespace).
			Str("repository", repository).
			Msg("Failed to get token for manifest check")
		return false, fmt.Errorf("failed to get token: %w", err)
	}

	url := fmt.Sprintf("%s/v2/%s/%s/manifests/%s",
		d.config.RegistryURL, namespace, repository, identifier)

	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to create request to %s", url)
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := d.doWithRetry(req)
	if err != nil {
		log.Logger().Error().Err(err).
			Str("url", url).
			Msg("Failed to check manifest existence")
		return false, fmt.Errorf("failed to check manifest: %w", err)
	}
	defer resp.Body.Close()

	exists = resp.StatusCode == http.StatusOK

	log.Logger().Debug().
		Str("namespace", namespace).
		Str("repository", repository).
		Str("identifier", identifier).
		Bool("exists", exists).
		Int("status_code", resp.StatusCode).
		Msg("Manifest existence check completed")

	return exists, nil
}

func (d *dockerClient) GetBlob(namespace, repository, digest string) (content []byte, err error) {
	log.Logger().Debug().
		Str("namespace", namespace).
		Str("repository", repository).
		Str("digest", digest).
		Msg("Fetching blob")

	token, err := d.getToken(namespace, repository, "pull")
	if err != nil {
		log.Logger().Error().Err(err).
			Str("namespace", namespace).
			Str("repository", repository).
			Str("digest", digest).
			Msg("Failed to get token for blob fetch")
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	url := fmt.Sprintf("%s/v2/%s/%s/blobs/%s",
		d.config.RegistryURL, namespace, repository, digest)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to create blob request to %s", url)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := d.doWithRetry(req)
	if err != nil {
		log.Logger().Error().Err(err).
			Str("url", url).
			Msg("Failed to fetch blob from upstream")
		return nil, fmt.Errorf("failed to fetch blob: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Logger().Error().
			Int("status_code", resp.StatusCode).
			Str("url", url).
			Str("response_body", string(body)).
			Msg("Unexpected status code while fetching blob")
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	content, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to read blob response from upstream: %s", url)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if d.config.LogBody {
		log.Logger().Debug().Int("blob_size", len(content)).Msg("Blob fetched")
	}

	log.Logger().Debug().
		Str("namespace", namespace).
		Str("repository", repository).
		Str("digest", digest).
		Int("size", len(content)).
		Msg("Blob fetched successfully")

	return content, nil
}

func (d *dockerClient) HeadBlob(namespace, repository, digest string) (exists bool, err error) {
	log.Logger().Debug().
		Str("namespace", namespace).
		Str("repository", repository).
		Str("digest", digest).
		Msg("Checking blob existence")

	token, err := d.getToken(namespace, repository, "pull")
	if err != nil {
		log.Logger().Error().Err(err).
			Str("namespace", namespace).
			Str("repository", repository).
			Str("digest", digest).
			Msg("Failed to get token for blob check")
		return false, fmt.Errorf("failed to get token: %w", err)
	}

	url := fmt.Sprintf("%s/v2/%s/%s/blobs/%s",
		d.config.RegistryURL, namespace, repository, digest)

	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to create request to %s", url)
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := d.doWithRetry(req)
	if err != nil {
		log.Logger().Error().Err(err).
			Str("url", url).
			Msg("Failed to check blob existence")
		return false, fmt.Errorf("failed to check blob: %w", err)
	}
	defer resp.Body.Close()

	exists = resp.StatusCode == http.StatusOK

	log.Logger().Debug().
		Str("namespace", namespace).
		Str("repository", repository).
		Str("digest", digest).
		Bool("exists", exists).
		Int("status_code", resp.StatusCode).
		Msg("Blob existence check completed")

	return exists, nil
}

func (d *dockerClient) getToken(namespace, repository, scope string) (string, error) {
	cacheKey := fmt.Sprintf("token:%s:%s:%s", namespace, repository, scope)

	if token := d.tokenCache.Get(cacheKey); token != "" {
		log.Logger().Debug().
			Str("namespace", namespace).
			Str("repository", repository).
			Str("scope", scope).
			Msg("Using cached token")
		return token, nil
	}

	log.Logger().Debug().
		Str("namespace", namespace).
		Str("repository", repository).
		Str("scope", scope).
		Msg("Fetching new token from auth server")

	url := fmt.Sprintf("%s?service=registry.docker.io&scope=repository:%s/%s:%s", d.config.TokenURL,
		namespace, repository, scope)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to create token request to %s", url)
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	if d.config.Username != "" && d.config.Password != "" {
		req.SetBasicAuth(d.config.Username, d.config.Password)
		log.Logger().Debug().Str("username", d.config.Username).Msg("Using authenticated token request")
	} else {
		log.Logger().Debug().Msg("Using anonymous token request")
	}

	timeStart := time.Now()

	resp, err := d.httpClient.Do(req)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to fetch token from %s", url)
		return "", fmt.Errorf("failed to fetch token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Logger().Error().
			Int("status_code", resp.StatusCode).
			Str("url", url).
			Str("response_body", string(body)).
			Msg("Token request failed")
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to decode token response from %s", url)
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	token := loginResp.Token
	if token == "" {
		token = loginResp.AccessToken
	}

	if token == "" {
		log.Logger().Error().Msg("Received empty token from auth server")
		return "", fmt.Errorf("received empty token from auth server")
	}

	ttl := time.Second * time.Duration(loginResp.ExpiresIn-(time.Now().UnixNano()-timeStart.UnixNano())/1e9)

	d.tokenCache.Set(cacheKey, token, ttl)

	log.Logger().Debug().
		Str("namespace", namespace).
		Str("repository", repository).
		Str("scope", scope).
		Dur("ttl", ttl).
		Msg("Token fetched and cached successfully")

	return token, nil
}

func (d *dockerClient) doWithRetry(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	delay := d.config.RetryDelay

	for attempt := 0; attempt <= d.config.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Logger().Debug().
				Int("attempt", attempt).
				Dur("delay", delay).
				Str("url", req.URL.String()).
				Msg("Retrying request")
			time.Sleep(delay)
			delay = time.Duration((float32(delay) * d.config.RetryBackOffMultiplier) * float32(time.Second))
		}

		reqClone := req.Clone(context.Background())

		resp, err = d.httpClient.Do(reqClone)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		if resp != nil {
			resp.Body.Close()
		}

		// don't retry on 4xx errors
		if resp != nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
			log.Logger().Warn().
				Int("status_code", resp.StatusCode).
				Str("url", req.URL.String()).
				Msg("Client error, not retrying")
			return resp, err
		}

		if err != nil {
			log.Logger().Warn().Err(err).
				Int("attempt", attempt).
				Str("url", req.URL.String()).
				Msg("Request failed, will retry")
		} else if resp != nil {
			log.Logger().Warn().
				Int("status_code", resp.StatusCode).
				Int("attempt", attempt).
				Str("url", req.URL.String()).
				Msg("Server error, will retry")
		}
	}

	if err != nil {
		log.Logger().Error().Err(err).
			Str("url", req.URL.String()).
			Int("max_retries", d.config.MaxRetries).
			Msg("Request failed after max retries")
		return nil, fmt.Errorf("max retries exceeded: %w", err)
	}

	log.Logger().Error().
		Str("url", req.URL.String()).
		Int("max_retries", d.config.MaxRetries).
		Msg("Request failed after max retries with no error")

	return resp, nil
}