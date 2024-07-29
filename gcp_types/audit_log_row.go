package gcp_types

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
)

// AuditLogRow represents an enriched row ready for parquet writing
type AuditLogRow struct {
	// embed required enrichment fields
	enrichment.CommonFields

	Timestamp time.Time `json:"timestamp"`
	LogName   string    `json:"logName"`

	// TODO: #finish add the rest of the fields
}
