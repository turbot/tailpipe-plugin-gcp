package tables

import (
	"github.com/turbot/tailpipe-plugin-gcp/mappers"
	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-gcp/sources"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/table"
)

const DataAccessAuditLogTableIdentifier = "gcp_data_access_audit_log"

func init() {
	// Register the table, with type parameters:
	// 1. row struct
	// 2. table config struct
	// 3. table implementation
	table.RegisterTable[*rows.AuditLog, *DataAccessAuditLogTableConfig, *DataAccessAuditLogTable]()
}

type DataAccessAuditLogTable struct {
}

func (c *DataAccessAuditLogTable) Identifier() string {
	return DataAccessAuditLogTableIdentifier
}

func (c *DataAccessAuditLogTable) SupportedSources(_ *DataAccessAuditLogTableConfig) []*table.SourceMetadata[*rows.AuditLog] {
	return []*table.SourceMetadata[*rows.AuditLog]{
		{
			SourceName: sources.AuditLogAPISourceIdentifier,
			MapperFunc: mappers.NewAuditLogMapper,
			Options: []row_source.RowSourceOption{
				sources.WithLogType(string(AuditLogTypeDataAccess)),
			},
		},
	}
}

func (c *DataAccessAuditLogTable) EnrichRow(row *rows.AuditLog, sourceEnrichmentFields *enrichment.CommonFields) (*rows.AuditLog, error) {
	return EnrichAuditLogRow(row, sourceEnrichmentFields)
}
