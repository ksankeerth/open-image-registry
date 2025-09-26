package oci

// -------- application/vnd.oci.image.index.v1+json ---------------
// OCIImageIndex represents the top-level OCI image index structure
type OCIImageIndex struct {
	SchemaVersion int               `json:"schemaVersion"`
	MediaType     string            `json:"mediaType"`
	Manifests     []OCIDescriptor   `json:"manifests"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}

// OCIDescriptor represents a content descriptor in the OCI spec
type OCIDescriptor struct {
	MediaType    string            `json:"mediaType"`
	Size         int64             `json:"size"`
	Digest       string            `json:"digest"`
	URLs         []string          `json:"urls,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	Data         []byte            `json:"data,omitempty"`
	Platform     *OCIPlatform      `json:"platform,omitempty"`
	ArtifactType string            `json:"artifactType,omitempty"`
}

// OCIPlatform represents the platform information
type OCIPlatform struct {
	Architecture string   `json:"architecture"`
	OS           string   `json:"os"`
	OSVersion    string   `json:"os.version,omitempty"`
	OSFeatures   []string `json:"os.features,omitempty"`
	Variant      string   `json:"variant,omitempty"`
	Features     []string `json:"features,omitempty"`
}

// ---------- application/vnd.oci.image.manifest.v1+json --------

type OCIImageManifest struct {
	SchemaVersion int               `json:"schemaVersion"`
	MediaType     string            `json:"mediaType"`
	Config        OCIDescriptor     `json:"config"`
	Layers        []OCIDescriptor   `json:"layers"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	Subject       *OCIDescriptor    `json:"subject,omitempty"`
	ArtifactType  string            `json:"artifactType,omitempty"`
}