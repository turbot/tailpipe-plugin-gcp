package gcp_types

import (
	"cloud.google.com/go/logging"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
)

type AuditLogRow struct {
	// embed required enrichment fields
	enrichment.CommonFields

	logging.Entry // TODO: Verify if this is the correct type and OK to use
}
