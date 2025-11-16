package upstream

type UpstreamClient interface {
	GetManifest(namespace, repository, identifier string) (content []byte, mediaType string, err error)

	HeadManifest(namespace, repository, identifier string) (exists bool, err error)

	GetBlob(namespace, repository, digest string) (content []byte, err error)

	HeadBlob(namespace, repository, digest string) (exists bool, err error)
}

