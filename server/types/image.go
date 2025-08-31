package types

import "time"

type ImageManifestWithPlatform struct {
	Digest    string
	Content   []byte
	MediaType string
	Size      int
	OS        string
	Arch      string
}

// ----------- Docker V2 spec types ------------------------
// Reference: https://docker-docs.uclv.cu/registry/spec/manifest-v2-2/
type DockerV2ManifestList struct {
	SchemaVersion int                     `json:"schemaVersion"`
	MediaType     string                  `json:"mediaType"`
	Manifests     []DockerV2ManifestEntry `json:"manifests"`
}

type DockerV2ManifestEntry struct {
	MediaType string                `json:"mediaType"`
	Size      int                   `json:"size"`
	Digest    string                `json:"digest"`
	Platform  DockerV2PlatformEntry `json:"platform"`
}

type DockerV2PlatformEntry struct {
	Architecture string   `json:"architecture"`
	OS           string   `json:"os"`
	OSVersion    string   `json:"os.version,omitempty"`
	OSFeatures   []string `json:"os.features,omitempty"`
	Variant      string   `json:"variant,omitempty"`
	Features     []string `json:"features,omitempty"`
}

type V2ImageManifest struct {
	SchemaVersion int                         `json:"schemaVersion"`
	MediaType     string                      `json:"mediaType"`
	Config        V2ManifestPlatformConfigRef `json:"config"`
	Layers        []V2ImageManifestLayer      `json:"layers"`
}

type V2ManifestPlatformConfigRef struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int    `json:"size"`
}

type V2ImageManifestLayer struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int    `json:"size"`
}

type ContainerConfig struct {
	AttachStderr bool                `json:"AttachStderr,omitempty"`
	AttachStdin  bool                `json:"AttachStdin,omitempty"`
	AttachStdout bool                `json:"AttachStdout,omitempty"`
	Cmd          []string            `json:"Cmd,omitempty"`
	Domainname   string              `json:"Domainname,omitempty"`
	Entrypoint   []string            `json:"Entrypoint,omitempty"`
	Env          []string            `json:"Env,omitempty"`
	Hostname     string              `json:"Hostname,omitempty"`
	Image        string              `json:"Image,omitempty"`
	Labels       map[string]string   `json:"Labels,omitempty"`
	OnBuild      []string            `json:"OnBuild,omitempty"`
	OpenStdin    bool                `json:"OpenStdin,omitempty"`
	StdinOnce    bool                `json:"StdinOnce,omitempty"`
	Tty          bool                `json:"Tty,omitempty"`
	User         string              `json:"User,omitempty"`
	Volumes      map[string]struct{} `json:"Volumes,omitempty"`
	WorkingDir   string              `json:"WorkingDir,omitempty"`
	ExposedPorts map[string]struct{} `json:"ExposedPorts,omitempty"`
	StopSignal   string              `json:"StopSignal,omitempty"`
	StopTimeout  *int                `json:"StopTimeout,omitempty"`
	Shell        []string            `json:"Shell,omitempty"`
}

// RootFS represents the root filesystem configuration
type RootFS struct {
	Type    string   `json:"type"`
	DiffIDs []string `json:"diff_ids"`
}

// History represents a layer's history entry
type History struct {
	Created    time.Time `json:"created"`
	CreatedBy  string    `json:"created_by"`
	EmptyLayer bool      `json:"empty_layer,omitempty"`
	Comment    string    `json:"comment,omitempty"`
	Author     string    `json:"author,omitempty"`
}

// ImageConfig represents the complete Docker container image configuration
// Media type: application/vnd.docker.container.image.v1+json
type ImageConfig struct {
	Architecture    string          `json:"architecture"`
	Config          ContainerConfig `json:"config"`
	Container       string          `json:"container,omitempty"`
	ContainerConfig ContainerConfig `json:"container_config,omitempty"`
	Created         time.Time       `json:"created"`
	DockerVersion   string          `json:"docker_version,omitempty"`
	History         []History       `json:"history,omitempty"`
	OS              string          `json:"os"`
	RootFS          RootFS          `json:"rootfs"`
	Size            int64           `json:"size,omitempty"`
	VirtualSize     int64           `json:"virtual_size,omitempty"`
	RepoTags        []string        `json:"repo_tags,omitempty"`
	RepoDigests     []string        `json:"repo_digests,omitempty"`
	Parent          string          `json:"parent,omitempty"`
	Comment         string          `json:"comment,omitempty"`
	Author          string          `json:"author,omitempty"`
}