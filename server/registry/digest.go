package registry

import (
	"encoding/json"
	"fmt"

	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/dockerv2"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/oci"
	"github.com/ksankeerth/open-image-registry/utils"
)

func UniqueDigest(mediaType string, content []byte) (digest string, err error) {
	var uniqueDigest string

	switch mediaType {
	case "application/vnd.docker.distribution.manifest.v2+json":
		var manifest dockerv2.ManifestV2
		if err := json.Unmarshal(content, &manifest); err != nil {
			return "", err
		}
		inputs := []string{manifest.Config.Digest}
		for _, entry := range manifest.Layers {
			inputs = append(inputs, entry.Digest)
		}
		uniqueDigest = utils.CombineAndCalculateSHA256Digest(inputs...)

	case "application/vnd.docker.distribution.manifest.list.v2+json":
		var manifest dockerv2.ManifestListV2
		if err := json.Unmarshal(content, &manifest); err != nil {
			return "", err
		}
		inputs := []string{}
		for _, entry := range manifest.Manifests {
			inputs = append(inputs, entry.Digest)
		}
		uniqueDigest = utils.CombineAndCalculateSHA256Digest(inputs...)

	case "application/vnd.oci.image.manifest.v1+json":
		var manifest oci.OCIImageManifest
		if err := json.Unmarshal(content, &manifest); err != nil {
			return "", err
		}
		inputs := []string{manifest.Config.Digest}
		for _, layer := range manifest.Layers {
			inputs = append(inputs, layer.Digest)
		}
		uniqueDigest = utils.CombineAndCalculateSHA256Digest(inputs...)

	case "application/vnd.oci.image.index.v1+json":
		var manifest oci.OCIImageIndex
		if err := json.Unmarshal(content, &manifest); err != nil {
			return "", err
		}
		if manifest.Manifests == nil {
			return "", fmt.Errorf("no manifests found in OCI index")
		}
		manifestDigests := []string{}
		for _, m := range manifest.Manifests {
			manifestDigests = append(manifestDigests, m.Digest)
		}
		uniqueDigest = utils.CombineAndCalculateSHA256Digest(manifestDigests...)

	default:
		return "", fmt.Errorf("unsupported mediaType: %s", mediaType)
	}

	return uniqueDigest, nil
}