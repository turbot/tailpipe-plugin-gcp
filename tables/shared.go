package tables

import (
	"time"

	"github.com/rs/xid"

	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
)

type AuditLogType string

const (
	AuditLogTypeActivity    AuditLogType = "activity"
	AuditLogTypeSystemEvent AuditLogType = "system_event"
	AuditLogTypeDataAccess  AuditLogType = "data_access"
)

func EnrichAuditLogRow(row *rows.AuditLog, sourceEnrichmentFields enrichment.SourceEnrichment) (*rows.AuditLog, error) {
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
