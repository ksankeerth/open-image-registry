package constants

import "time"

// audit-events
const (
	DefaultAuditBucketLimit   = 4000000
	DefaultAuditBatchSize     = 100
	DefaultAuditBatchWaitTime = time.Minute * 2
	DefaultAuditSqliteBuckets = 10
)

// security
const (
	TokenSigningAlgoES256 = "ES256"
)