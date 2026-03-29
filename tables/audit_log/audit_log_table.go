package audit_log

import (
	"time"

	"github.com/rs/xid"

	"github.com/turbot/pipe-fittings/v2/utils"
	logging_log_entry "github.com/turbot/tailpipe-plugin-gcp/sources/logging_log_entry"
	"github.com/turbot/tailpipe-plugin-gcp/sources/storage_bucket"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source_config"
	"github.com/turbot/tailpipe-plugin-sdk/constants"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/schema"
	"github.com/turbot/tailpipe-plugin-sdk/table"
)

const AuditLogTableIdentifier string = "gcp_audit_log"

type AuditLogTable struct {
}

func (c *AuditLogTable) Identifier() string {
	return AuditLogTableIdentifier
}

func (c *AuditLogTable) GetSourceMetadata() ([]*table.SourceMetadata[*AuditLog], error) {
	defaultArtifactConfig := &artifact_source_config.ArtifactSourceConfigImpl{
		FileLayout: utils.ToStringPointer("cloudaudit.googleapis.com/%{DATA:type}/%{YEAR:year}/%{MONTHNUM:month}/%{MONTHDAY:day}/%{HOUR:hour}:%{MINUTE:minute}:%{SECOND:second}_%{DATA:end_time}_%{DATA:suffix}.json"),
	}

	return []*table.SourceMetadata[*AuditLog]{
		{
			SourceName: logging_log_entry.LoggingLogEntrySourceIdentifier,
			Mapper:     &AuditLogMapper{},
			// The options here are required to validate and pull the different types of logs per table.
			// The NewLogTypeFilter function maps the supported log types for each table, as defined in
			// logging_log_type_filter_map.go.
			Options: []row_source.RowSourceOption{
				logging_log_entry.WithTableName(AuditLogTableIdentifier),
			},
		},
		{
			SourceName: storage_bucket.GcpStorageBucketSourceIdentifier,
			Mapper:     &AuditLogMapper{},
			Options: []row_source.RowSourceOption{
				artifact_source.WithDefaultArtifactSourceConfig(defaultArtifactConfig),
				artifact_source.WithRowPerLine(),
			},
		},
		{
			SourceName: constants.ArtifactSourceIdentifier,
			Mapper:     &AuditLogMapper{},
			Options: []row_source.RowSourceOption{
				artifact_source.WithDefaultArtifactSourceConfig(defaultArtifactConfig),
			},
		},
	}, nil
}

func (c *AuditLogTable) EnrichRow(row *AuditLog, sourceEnrichmentFields schema.SourceEnrichment) (*AuditLog, error) {
	row.CommonFields = sourceEnrichmentFields.CommonFields

	row.TpID = xid.New().String()
	row.TpTimestamp = row.Timestamp
	row.TpIngestTimestamp = time.Now()
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

func (c *AuditLogTable) GetDescription() string {
	return "GCP Audit Logs track administrative and data access activities across Google Cloud services, capturing user actions and system events for security and compliance monitoring."
}
