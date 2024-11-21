package tables

import (
	"time"

	"github.com/rs/xid"

	"github.com/turbot/tailpipe-plugin-gcp/mappers"
	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-gcp/sources"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
	"github.com/turbot/tailpipe-plugin-sdk/table"
)

const AuditLogTableIdentifier = "gcp_audit_log"

func init() {
	table.RegisterTable[*rows.AuditLog, *AuditLogTable]()
}

type AuditLogTable struct {
	table.TableImpl[*rows.AuditLog, *AuditLogTableConfig, *artifact_source.GcpConnection]
}

func (c *AuditLogTable) Identifier() string {
	return AuditLogTableIdentifier
}

func (c *AuditLogTable) SupportedSources() []*table.SourceMetadata[*rows.AuditLog] {
	return []*table.SourceMetadata[*rows.AuditLog]{
		{
			SourceName: sources.AuditLogAPISourceIdentifier,
			MapperFunc: mappers.NewAuditLogMapper,
		},
	}
}

func (c *AuditLogTable) EnrichRow(row *rows.AuditLog, sourceEnrichmentFields *enrichment.CommonFields) (*rows.AuditLog, error) {

	if sourceEnrichmentFields != nil {
		row.CommonFields = *sourceEnrichmentFields
	}

	row.TpID = xid.New().String()
	row.TpTimestamp = row.Timestamp
	row.TpIngestTimestamp = time.Now()
	row.TpIndex = *row.CommonFields.TpSourceLocation // project
	row.TpDate = row.Timestamp.Truncate(24 * time.Hour)

	if row.AuthenticationPrincipal != nil {
		row.TpUsernames = append(row.TpUsernames, *row.AuthenticationPrincipal)
	}
	if row.RequestCallerIp != nil {
		row.TpIps = append(row.TpIps, *row.RequestCallerIp)
		row.TpSourceIP = row.RequestCallerIp
	}

	// TODO: #finish payload.Request is a struct which has `Fields` property of map[string]*Value - how to handle? (common keys: @type / name - but this can literally contain anything!)
	// TODO: #finish payload.AuthorizationInfo is an array of structs with Resource (string), Permission (string), and Granted (bool) properties, seems to mostly be a single item but could be more - best way to handle?

	return row, nil
}
