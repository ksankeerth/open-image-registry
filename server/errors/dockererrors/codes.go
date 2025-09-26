package dockererrors

// Docker Registry v2 API Error Codes
// Based on: https://distribution.github.io/distribution/spec/api/#errors-2
const (
	// Blob related errors
	ErrCodeBlobUnknown         = "BLOB_UNKNOWN"
	ErrCodeBlobUploadInvalid   = "BLOB_UPLOAD_INVALID"
	ErrCodeBlobUploadUnknown   = "BLOB_UPLOAD_UNKNOWN"
	ErrCodeManifestBlobUnknown = "MANIFEST_BLOB_UNKNOWN"

	// Digest and content validation errors
	ErrCodeDigestInvalid = "DIGEST_INVALID"
	ErrCodeSizeInvalid   = "SIZE_INVALID"
	ErrCodeRangeInvalid  = "RANGE_INVALID"

	// Manifest related errors
	ErrCodeManifestInvalid    = "MANIFEST_INVALID"
	ErrCodeManifestUnknown    = "MANIFEST_UNKNOWN"
	ErrCodeManifestUnverified = "MANIFEST_UNVERIFIED"

	// Repository and naming errors
	ErrCodeNameInvalid = "NAME_INVALID"
	ErrCodeNameUnknown = "NAME_UNKNOWN"
	ErrCodeTagInvalid  = "TAG_INVALID"

	// Authentication and authorization errors
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeDenied       = "DENIED"

	// Operational errors
	ErrCodeUnsupported             = "UNSUPPORTED"
	ErrCodeTooManyRequests         = "TOOMANYREQUESTS"
	ErrCodePaginationNumberInvalid = "PAGINATION_NUMBER_INVALID"
)

// ErrorMessages maps error codes to their standard messages
var ErrorMessages = map[string]string{
	ErrCodeBlobUnknown:             "blob unknown to registry",
	ErrCodeBlobUploadInvalid:       "blob upload invalid",
	ErrCodeBlobUploadUnknown:       "blob upload unknown to registry",
	ErrCodeManifestBlobUnknown:     "blob unknown to registry",
	ErrCodeDigestInvalid:           "provided digest did not match uploaded content",
	ErrCodeSizeInvalid:             "provided length did not match content length",
	ErrCodeRangeInvalid:            "invalid content range",
	ErrCodeManifestInvalid:         "manifest invalid",
	ErrCodeManifestUnknown:         "manifest unknown",
	ErrCodeManifestUnverified:      "manifest failed signature verification",
	ErrCodeNameInvalid:             "invalid repository name",
	ErrCodeNameUnknown:             "repository name not known to registry",
	ErrCodeTagInvalid:              "manifest tag did not match URI",
	ErrCodeUnauthorized:            "authentication required",
	ErrCodeDenied:                  "requested access to the resource is denied",
	ErrCodeUnsupported:             "The operation is unsupported",
	ErrCodeTooManyRequests:         "too many requests",
	ErrCodePaginationNumberInvalid: "invalid number of results requested",
}

// StatusCodes maps error codes to HTTP status codes
var StatusCodes = map[string]int{
	ErrCodeBlobUnknown:             404,
	ErrCodeBlobUploadInvalid:       400,
	ErrCodeBlobUploadUnknown:       404,
	ErrCodeManifestBlobUnknown:     404,
	ErrCodeDigestInvalid:           400,
	ErrCodeSizeInvalid:             400,
	ErrCodeRangeInvalid:            416,
	ErrCodeManifestInvalid:         400,
	ErrCodeManifestUnknown:         404,
	ErrCodeManifestUnverified:      400,
	ErrCodeNameInvalid:             400,
	ErrCodeNameUnknown:             404,
	ErrCodeTagInvalid:              400,
	ErrCodeUnauthorized:            401,
	ErrCodeDenied:                  403,
	ErrCodeUnsupported:             405,
	ErrCodeTooManyRequests:         429,
	ErrCodePaginationNumberInvalid: 400,
}