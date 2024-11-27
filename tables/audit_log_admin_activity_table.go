package tables

import (
	"github.com/turbot/tailpipe-plugin-gcp/mappers"
	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-gcp/sources"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/table"
)

const AuditLogAdminActivityTableIdentifier = "gcp_audit_log_admin_activity"

func init() {
	// Register the table, with type parameters:
	// 1. row struct
	// 2. table config struct
	// 3. table implementation
	table.RegisterTable[*rows.AuditLog, *AuditLogAdminActivityTableConfig, *AuditLogAdminActivityTable]()
}

type AuditLogAdminActivityTable struct {
}

func (c *AuditLogAdminActivityTable) Identifier() string {
	return AuditLogAdminActivityTableIdentifier
}

func (c *AuditLogAdminActivityTable) SupportedSources(_ *AuditLogAdminActivityTableConfig) []*table.SourceMetadata[*rows.AuditLog] {
	return []*table.SourceMetadata[*rows.AuditLog]{
		{
			SourceName: sources.AuditLogAPISourceIdentifier,
			MapperFunc: mappers.NewAuditLogMapper,
			Options: []row_source.RowSourceOption{
				sources.WithLogType(string(AuditLogTypeActivity)),
			},
		},
	}
}

func (c *AuditLogAdminActivityTable) EnrichRow(row *rows.AuditLog, sourceEnrichmentFields *enrichment.CommonFields) (*rows.AuditLog, error) {
	return EnrichAuditLogRow(row, sourceEnrichmentFields)
}
