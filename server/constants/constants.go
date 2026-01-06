package constants

const NamespacePurposeTeam = "Team"
const NamespacePurposeProject = "Project"

const ResourceStateActive = "Active"
const ResourceStateDeprecated = "Deprecated"
const ResourceStateDisabled = "Disabled"

const ResourceTypeNamespace = "namespace"
const ResourceTypeRepository = "repository"
const ResourceTypeUpstream = "upstream"

const AccessLevelMaintainer = "Maintainer"
const AccessLevelDeveloper = "Developer"
const AccessLevelGuest = "Guest"

// For Upstream registries, random UUID will be generated.
// For images which are supposed to be hosted in OpenImageRegistry,
// We'll use this ID.
const HostedRegistryID = "1"
const HostedRegistryName = "HostedRegistry"

// By default, the namespace can be username or organization name.
// If namespace is not provided, We'll use this namespace.
const DefaultNamespace = "library"

const UnknownBlobMediaType = "unknown_media_type"

const (
	RegistryVendorDockerHub   = "docker_hub"
	RegistryVendorGCR         = "gcr"
	RegistryVendorECR         = "ecr"
	RegistryVendorACR         = "acr"
	RegistryVendorGHCR        = "ghcr"
	RegistryVendorGitLab      = "gitlab"
	RegistryVendorQuay        = "quay"
	RegistryVendorHarbor      = "harbor"
	RegistryVendorArtifactory = "artifactory"
	RegistryVendorNexus       = "nexus"
	RegistryVendorCustom      = "custom"
)