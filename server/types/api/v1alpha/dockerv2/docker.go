package dockerv2

import "time"

// application/vnd.docker.distribution.manifest.v2+json
type ManifestV2 struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Config        struct {
		MediaType string `json:"mediaType"`
		Size      int64  `json:"size"`
		Digest    string `json:"digest"`
	} `json:"config"`
	Layers []struct {
		MediaType string   `json:"mediaType"`
		Size      int64    `json:"size"`
		Digest    string   `json:"digest"`
		URLs      []string `json:"urls,omitempty"`
	} `json:"layers"`
	Subject *Subject `json:"subject,omitempty"`
}

// ManifestListV2 represents a Docker Manifest List V2 (multi-platform)
// application/vnd.docker.distribution.manifest.list.v2+json
type ManifestListV2 struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Manifests     []struct {
		MediaType string `json:"mediaType"`
		Size      int64  `json:"size"`
		Digest    string `json:"digest"`
		Platform  struct {
			Architecture string   `json:"architecture"`
			OS           string   `json:"os"`
			OSVersion    string   `json:"os.version,omitempty"`
			OSFeatures   []string `json:"os.features,omitempty"`
			Variant      string   `json:"variant,omitempty"`
			Features     []string `json:"features,omitempty"`
		} `json:"platform"`
		URLs []string `json:"urls,omitempty"`
	} `json:"manifests"`
	Subject *Subject `json:"subject,omitempty"`
}

// Subject represents a subject reference (optional)
type Subject struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

type DockerRegistryLoginResponse struct {
	Token       string    `json:"token"`
	AccessToken string    `json:"access_token"`
	ExpiresIn   int64     `json:"expires_in"`
	IssuedAt    time.Time `json:"issued_at"`
}