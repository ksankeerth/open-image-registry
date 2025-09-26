package registry

// For Upstream registries, random UUID will be generated.
// For images which are supposed to be hosted in OpenImageRegistry,
// We'll use this ID.
const HostedRegistryID = "1"
const HostedRegistryName = "HostedRegistry"

// By default, the namespace can be username or organization name.
// If namespace is not provided, We'll use this namespace.
const DefaultNamespace = "library"

const UnknownBlobMediaType = "unknown_media_type"
