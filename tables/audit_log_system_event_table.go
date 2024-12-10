package tables

import (
	"github.com/turbot/tailpipe-plugin-gcp/mappers"
	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-gcp/sources"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/table"
)

const AuditLogSystemEventTableIdentifier = "gcp_audit_log_system_event"

func init() {
	// Register the table, with type parameters:
	// 1. row struct
	// 2. table config struct
	// 3. table implementation
	table.RegisterTable[*rows.AuditLog, *AuditLogSystemEventTableConfig, *AuditLogSystemEventTable]()
}

type AuditLogSystemEventTable struct {
}

func (c *AuditLogSystemEventTable) Identifier() string {
	return AuditLogSystemEventTableIdentifier
}

func (c *AuditLogSystemEventTable) GetSourceMetadata(_ *AuditLogSystemEventTableConfig) []*table.SourceMetadata[*rows.AuditLog] {
	return []*table.SourceMetadata[*rows.AuditLog]{
		{
			SourceName: sources.AuditLogAPISourceIdentifier,
			Mapper:     &mappers.AuditLogMapper{},
			Options: []row_source.RowSourceOption{
				sources.WithLogType(string(AuditLogTypeSystemEvent)),
			},
		},
	}
}

func (c *AuditLogSystemEventTable) EnrichRow(row *rows.AuditLog, _ *AuditLogSystemEventTableConfig, sourceEnrichmentFields enrichment.SourceEnrichment) (*rows.AuditLog, error) {
	return EnrichAuditLogRow(row, sourceEnrichmentFields)
}
