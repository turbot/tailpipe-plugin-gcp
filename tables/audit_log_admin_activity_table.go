package tables

import (
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/tailpipe-plugin-gcp/extractors"
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
	defaultArtifactConfig := &artifact_source_config.ArtifactSourceConfigBase{
		FileLayout: utils.ToStringPointer("/activity/\\d{4}/\\d{2}/\\d{2}/\\d{2}:\\d{2}:\\d{2}_\\d{2}:\\d{2}:\\d{2}_S\\d\\.json"),
	}
	return []*table.SourceMetadata[*rows.AuditLog]{
		{
			// any artifact source
			SourceName: constants.ArtifactSourceIdentifier,
			Options: []row_source.RowSourceOption{
				artifact_source.WithDefaultArtifactSourceConfig(defaultArtifactConfig),
				artifact_source.WithArtifactExtractor(extractors.NewActivityLogExtractor()),
			},
		},
		{
			SourceName: sources.AuditLogAPISourceIdentifier,
			Mapper:     &mappers.AuditLogMapper{},
			Options: []row_source.RowSourceOption{
				sources.WithLogType(string(AuditLogTypeActivity)),
			},
		},
	}
}

func (c *AuditLogAdminActivityTable) EnrichRow(row *rows.AuditLog, _ *AuditLogAdminActivityTableConfig, sourceEnrichmentFields enrichment.SourceEnrichment) (*rows.AuditLog, error) {
	return EnrichAuditLogRow(row, sourceEnrichmentFields)
}
