package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ksankeerth/open-image-registry/errors/client"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/types"
)

// TODO: First, We will develop only for anonymous access and later we'll extend it.
type DockerV2RegistryClient struct {
	registryUrl   string
	authUrl       string
	enableWireLog bool
}

func NewDockerV2RegistryClient(registryUrl, authUrl string) *DockerV2RegistryClient {
	if strings.Contains(registryUrl, "docker") && authUrl == "" {
		log.Logger().Info().
			Msgf("Token endpoint of %s is empty. Therefore, DockerHub authentication endpoint will be used %s",
				registryUrl, "https://auth.docker.io/token")
		authUrl = "https://auth.docker.io/token"
	}
	return &DockerV2RegistryClient{
		registryUrl:   registryUrl,
		authUrl:       authUrl,
		enableWireLog: false,
	}
}

// logRequest logs the outgoing HTTP request details
func (c *DockerV2RegistryClient) logRequest(req *http.Request, operation string) {
	if !c.enableWireLog {
		return
	}

	logger := log.Logger().Debug().
		Str("operation", operation).
		Str("method", req.Method).
		Str("url", req.URL.String()).
		Str("protocol", req.Proto)

	headers := make(map[string]string)
	for name, values := range req.Header {
		if name == "Authorization" {
			// Mask authorization header for security
			if len(values) > 0 && strings.HasPrefix(values[0], "Bearer ") {
				headers[name] = "Bearer ***MASKED***"
			} else {
				headers[name] = "***MASKED***"
			}
		} else {
			headers[name] = strings.Join(values, ", ")
		}
	}

	logger.Interface("headers", headers).
		Msg("HTTP Request")
}

// logResponse logs the incoming HTTP response details
func (c *DockerV2RegistryClient) logResponse(resp *http.Response, operation string, duration time.Duration) {
	if !c.enableWireLog {
		return
	}

	logger := log.Logger().Debug().
		Str("operation", operation).
		Str("method", resp.Request.Method).
		Str("url", resp.Request.URL.String()).
		Int("status_code", resp.StatusCode).
		Str("status", resp.Status).
		Str("protocol", resp.Proto).
		Dur("duration", duration)

	headers := make(map[string]string)
	for name, values := range resp.Header {
		headers[name] = strings.Join(values, ", ")
	}

	logger.Interface("headers", headers).
		Int64("content_length", resp.ContentLength).
		Msg("HTTP Response")
}

// executeRequestWithLogging performs HTTP request and log
func (c *DockerV2RegistryClient) executeRequestWithLogging(req *http.Request, operation string) (*http.Response, error) {
	c.logRequest(req, operation)

	httpClient := &http.Client{}
	startTime := time.Now()

	resp, err := httpClient.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		log.Logger().Error().Err(err).
			Str("operation", operation).
			Str("url", req.URL.String()).
			Dur("duration", duration).
			Msg("HTTP Request Failed")
		return resp, err
	}
	defer resp.Body.Close()

	c.logResponse(resp, operation, duration)
	return resp, nil
}

func (c *DockerV2RegistryClient) getToken(scope string) (string, error) {
	url := fmt.Sprintf("%s?service=registry.docker.io&scope=%s", c.authUrl, scope)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", client.ClassifyError(err, fmt.Sprintf("GET %s", url), nil)
	}

	resp, err := c.executeRequestWithLogging(req, "getToken")
	if err != nil {
		return "", client.ClassifyError(err, fmt.Sprintf("GET %s", url), resp)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Logger().Error().Msgf("Unable to get token from url: %s", url)
		return "", client.ClassifyError(nil, fmt.Sprintf("GET %s", url), resp)
	}

	var loginResp types.DockerRegistryLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", client.ClassifyError(err, fmt.Sprintf("GET %s", url), resp)
	}

	// Docker may return "token" or "access_token"
	if loginResp.Token != "" {
		// log.Logger().Debug().Msgf("Received token from %s: %s", c.authUrl, loginResp.Token)
		return loginResp.Token, nil
	}
	if loginResp.AccessToken != "" {
		// log.Logger().Debug().Msgf("Received token from %s: %s", c.authUrl, loginResp.AccessToken)
		return loginResp.AccessToken, nil
	}
	return "", client.NewProxyClientError(fmt.Sprintf("GET %s", url), nil, client.CodeProxyResponseBodyMismatch)
}

func (c *DockerV2RegistryClient) GetManifest(namespace, repository, tagOrDigest string) (content []byte,
	contentType string, err error) {
	token, err := c.getToken(fmt.Sprintf("repository:%s/%s:pull", namespace, repository))
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to get token from %s", c.authUrl)
		return nil, "", err
	}

	url := fmt.Sprintf("%s/v2/%s/%s/manifests/%s", c.registryUrl, namespace, repository, tagOrDigest)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", client.ClassifyError(err, fmt.Sprintf("GET %s", url), nil)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", strings.Join([]string{
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.oci.image.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.list.v2+json",
		"application/vnd.oci.image.index.v1+json",
	}, ", "))

	resp, err := c.executeRequestWithLogging(req, "GetManifest")
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to retrieve manifest from %s", c.registryUrl)
		return nil, "", client.ClassifyError(err, fmt.Sprintf("GET %s", url), resp)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Logger().Error().Msgf("Unexpected status code was received for url(%s); Body = %s", req.RequestURI, body)
		return nil, "", client.ClassifyError(nil, fmt.Sprintf("GET %s", url), resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", client.ClassifyError(err, fmt.Sprintf("GET %s", url), resp)
	}

	contentType = resp.Header.Get("Content-Type")
	return body, contentType, nil
}

func (c *DockerV2RegistryClient) HeadManifest(namespace, repository, tagOrDigest string) (exists bool, digest,
	contentType string, err error) {
	token, err := c.getToken(fmt.Sprintf("repository:%s/%s:pull", namespace, repository))
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to get token from %s", c.authUrl)
		return false, "", "", err
	}

	url := fmt.Sprintf("%s/v2/%s/%s/manifests/%s", c.registryUrl, namespace, repository, tagOrDigest)

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false, "", "", client.ClassifyError(err, fmt.Sprintf("HEAD %s", url), nil)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", strings.Join([]string{
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.oci.image.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.list.v2+json",
		"application/vnd.oci.image.index.v1+json",
	}, ", "))

	resp, err := c.executeRequestWithLogging(req, "HeadManifest")
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to retrieve manifest from %s", c.registryUrl)
		return false, "", "", client.ClassifyError(err, fmt.Sprintf("HEAD %s", url), resp)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, "", "", nil
	}

	if resp.StatusCode != http.StatusOK {
		log.Logger().Error().Msgf("Unexpected status code was received for url(%s);", req.RequestURI)
		return false, "", "", client.ClassifyError(nil, fmt.Sprintf("HEAD %s", url), resp)
	}

	digest = resp.Header.Get("docker-content-digest")
	contentType = resp.Header.Get("Content-Type")
	return true, digest, contentType, nil
}

func (c *DockerV2RegistryClient) GetBlob(namespace, repository, digest string) (content []byte, err error) {
	// Get an auth token for pulling the blob
	token, err := c.getToken(fmt.Sprintf("repository:%s/%s:pull", namespace, repository))
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to get token from %s", c.authUrl)
		return nil, err
	}

	url := fmt.Sprintf("%s/v2/%s/%s/blobs/%s", c.registryUrl, namespace, repository, digest)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, client.ClassifyError(err, fmt.Sprintf("GET %s", url), nil)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.executeRequestWithLogging(req, "GetBlob")
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to retrieve blob from %s", c.registryUrl)
		return nil, client.ClassifyError(err, fmt.Sprintf("GET %s", url), resp)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Logger().Error().Msgf("Unexpected status code received for URL %s; Body = %s", url, body)
		return nil, client.ClassifyError(nil, fmt.Sprintf("GET %s", url), resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, client.ClassifyError(err, fmt.Sprintf("GET %s", url), resp)
	}

	return body, nil
}

func (c *DockerV2RegistryClient) HeadBlob(namespace, repository, digest string) (exists bool, err error) {
	// Get an auth token for pulling the blob
	token, err := c.getToken(fmt.Sprintf("repository:%s/%s:pull", namespace, repository))
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to get token from %s", c.authUrl)
		return false, err
	}

	url := fmt.Sprintf("%s/v2/%s/%s/blobs/%s", c.registryUrl, namespace, repository, digest)

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false, client.ClassifyError(err, fmt.Sprintf("HEAD %s", url), nil)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.executeRequestWithLogging(req, "HeadBlob")
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to retrieve blob from %s", c.registryUrl)
		return false, client.ClassifyError(err, fmt.Sprintf("HEAD %s", url), resp)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Logger().Error().Msgf("Unexpected status code received for URL %s; Body = %s", url, body)
		return false, client.ClassifyError(nil, fmt.Sprintf("HEAD %s", url), resp)
	}

	return true, nil
}

func (c *DockerV2RegistryClient) SetWireLogging(enabled bool) {
	c.enableWireLog = enabled
}

func (c *DockerV2RegistryClient) IsWireLoggingEnabled() bool {
	return c.enableWireLog
}
