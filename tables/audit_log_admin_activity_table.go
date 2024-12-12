package tables

import (
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/tailpipe-plugin-gcp/mappers"
	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-gcp/sources"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source_config"
	"github.com/turbot/tailpipe-plugin-sdk/constants"
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

func (c *AuditLogAdminActivityTable) GetSourceMetadata(_ *AuditLogAdminActivityTableConfig) []*table.SourceMetadata[*rows.AuditLog] {
	// the default file layout for Admin Activity Logs in GCP Storage Buckets
	defaultArtifactConfig := &artifact_source_config.ArtifactSourceConfigBase{
		FileLayout: utils.ToStringPointer("cloudaudit\\.googleapis\\.com/activity/(?P<year>\\d{4})/(?P<month>\\d{2})/(?P<day>\\d{2})/(?P<hour>\\d{2}).*\\.json"),
	}

	return []*table.SourceMetadata[*rows.AuditLog]{
		{
			SourceName: sources.AuditLogAPISourceIdentifier,
			Mapper:     &mappers.AuditLogMapper{},
			Options: []row_source.RowSourceOption{
				sources.WithLogType(string(AuditLogTypeActivity)),
			},
		},
		{
			SourceName: constants.ArtifactSourceIdentifier,
			Mapper:     &mappers.AuditLogMapper{},
			Options: []row_source.RowSourceOption{
				artifact_source.WithDefaultArtifactSourceConfig(defaultArtifactConfig),
				artifact_source.WithRowPerLine(),
			},
		},
	}
}

func (c *AuditLogAdminActivityTable) EnrichRow(row *rows.AuditLog, _ *AuditLogAdminActivityTableConfig, sourceEnrichmentFields enrichment.SourceEnrichment) (*rows.AuditLog, error) {
	return EnrichAuditLogRow(row, sourceEnrichmentFields)
}
