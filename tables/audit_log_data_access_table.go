package tables

import (
	"github.com/turbot/tailpipe-plugin-gcp/mappers"
	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-gcp/sources"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/table"
)

const AuditLogDataAccessTableIdentifier = "gcp_audit_log_data_access"

func init() {
	// Register the table, with type parameters:
	// 1. row struct
	// 2. table config struct
	// 3. table implementation
	table.RegisterTable[*rows.AuditLog, *AuditLogDataAccessTableConfig, *AuditLogDataAccessTable]()
}

type AuditLogDataAccessTable struct {
}

func (c *AuditLogDataAccessTable) Identifier() string {
	return AuditLogDataAccessTableIdentifier
}

func (c *AuditLogDataAccessTable) GetSourceMetadata(_ *AuditLogDataAccessTableConfig) []*table.SourceMetadata[*rows.AuditLog] {
	return []*table.SourceMetadata[*rows.AuditLog]{
		{
			SourceName: sources.AuditLogAPISourceIdentifier,
			Mapper:     &mappers.AuditLogMapper{},
			Options: []row_source.RowSourceOption{
				sources.WithLogType(string(AuditLogTypeDataAccess)),
			},
		},
	}
}

func (c *AuditLogDataAccessTable) EnrichRow(row *rows.AuditLog, _ *AuditLogDataAccessTableConfig, sourceEnrichmentFields enrichment.SourceEnrichment) (*rows.AuditLog, error) {
	return EnrichAuditLogRow(row, sourceEnrichmentFields)
}
