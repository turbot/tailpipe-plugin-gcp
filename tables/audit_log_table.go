package tables

import (
	"time"

	"github.com/rs/xid"

	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/tailpipe-plugin-gcp/mappers"
	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-gcp/sources"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source_config"
	"github.com/turbot/tailpipe-plugin-sdk/constants"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/schema"
	"github.com/turbot/tailpipe-plugin-sdk/table"
)

const AuditLogTableIdentifier string = "gcp_audit_log"

func init() {
	table.RegisterTable[*rows.AuditLog, *AuditLogTable]()
}

type AuditLogTable struct {
}

func (c *AuditLogTable) Identifier() string {
	return AuditLogTableIdentifier
}

func (c *AuditLogTable) GetSourceMetadata() []*table.SourceMetadata[*rows.AuditLog] {
	defaultArtifactConfig := &artifact_source_config.ArtifactSourceConfigBase{
		FileLayout: utils.ToStringPointer("cloudaudit\\.googleapis\\.com/(?P<type>\\w+)/(?P<year>\\d{4})/(?P<month>\\d{2})/(?P<day>\\d{2})/(?P<hour>\\d{2}).*\\.json"),
	}

	return []*table.SourceMetadata[*rows.AuditLog]{
		{
			SourceName: sources.AuditLogAPISourceIdentifier,
			Mapper:     &mappers.AuditLogMapper{},
		},
		{
			SourceName: constants.ArtifactSourceIdentifier,
			Mapper:     &mappers.AuditLogMapper{},
			Options: []row_source.RowSourceOption{
				artifact_source.WithDefaultArtifactSourceConfig(defaultArtifactConfig),
			},
		},
	}
}

func (c *AuditLogTable) EnrichRow(row *rows.AuditLog, sourceEnrichmentFields schema.SourceEnrichment) (*rows.AuditLog, error) {
	row.CommonFields = sourceEnrichmentFields.CommonFields

	row.TpID = xid.New().String()
	row.TpTimestamp = row.Timestamp
	row.TpIngestTimestamp = time.Now()
	row.TpIndex = *row.CommonFields.TpSourceLocation // project
	row.TpDate = row.Timestamp.Truncate(24 * time.Hour)

	if row.AuthenticationInfo != nil {
		if row.AuthenticationInfo.PrincipalEmail != "" {
			row.TpUsernames = append(row.TpUsernames, row.AuthenticationInfo.PrincipalEmail)
			row.TpEmails = append(row.TpEmails, row.AuthenticationInfo.PrincipalEmail)
		}
		if row.AuthenticationInfo.PrincipalSubject != "" {
			row.TpUsernames = append(row.TpUsernames, row.AuthenticationInfo.PrincipalSubject)
		}
	}

	if row.HttpRequest != nil {
		if row.HttpRequest.LocalIp != "" {
			row.TpIps = append(row.TpIps, row.HttpRequest.LocalIp)
			row.TpSourceIP = &row.HttpRequest.LocalIp
		}
		if row.HttpRequest.RemoteIp != "" {
			row.TpIps = append(row.TpIps, row.HttpRequest.RemoteIp)
			row.TpDestinationIP = &row.HttpRequest.RemoteIp
		}
	}

	if row.RequestMetadata != nil {
		if row.RequestMetadata.CallerIp != "" {
			row.TpIps = append(row.TpIps, row.RequestMetadata.CallerIp)
			row.TpSourceIP = &row.RequestMetadata.CallerIp
		}
	}

	return row, nil
}
