package requests_log

import (
	"time"

	"github.com/rs/xid"

	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/tailpipe-plugin-gcp/sources/audit_log_api"
	"github.com/turbot/tailpipe-plugin-gcp/sources/storage_bucket"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source_config"
	"github.com/turbot/tailpipe-plugin-sdk/constants"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/schema"
	"github.com/turbot/tailpipe-plugin-sdk/table"
)

const RequestsLogTableIdentifier string = "gcp_requests_log"

type RequestsLogTable struct {
}

func (c *RequestsLogTable) Identifier() string {
	return RequestsLogTableIdentifier
}

func (c *RequestsLogTable) GetSourceMetadata() ([]*table.SourceMetadata[*RequestsLog], error) {
	defaultArtifactConfig := &artifact_source_config.ArtifactSourceConfigImpl{
		FileLayout: utils.ToStringPointer("requests/%{YEAR:year}/%{MONTHNUM:month}/%{MONTHDAY:day}/%{HOUR:hour}:%{MINUTE:minute}:%{SECOND:second}_%{DATA:end_time}_%{DATA:suffix}.json"),
	}

	return []*table.SourceMetadata[*RequestsLog]{
		{
			SourceName: audit_log_api.AuditLogAPISourceIdentifier,
			Mapper:     &RequestsLogMapper{},
		},
		{
			SourceName: storage_bucket.GcpStorageBucketSourceIdentifier,
			Mapper:     &RequestsLogMapper{},
			Options: []row_source.RowSourceOption{
				artifact_source.WithDefaultArtifactSourceConfig(defaultArtifactConfig),
				artifact_source.WithRowPerLine(),
			},
		},
		{
			SourceName: constants.ArtifactSourceIdentifier,
			Mapper:     &RequestsLogMapper{},
			Options: []row_source.RowSourceOption{
				artifact_source.WithDefaultArtifactSourceConfig(defaultArtifactConfig),
			},
		},
	}, nil
}

func (c *RequestsLogTable) EnrichRow(row *RequestsLog, sourceEnrichmentFields schema.SourceEnrichment) (*RequestsLog, error) {
	if row == nil {
		return nil, nil
	}

	row.CommonFields = sourceEnrichmentFields.CommonFields

	row.TpID = xid.New().String()
	row.TpTimestamp = row.Timestamp
	row.TpIngestTimestamp = time.Now()
	row.TpDate = row.Timestamp.Truncate(24 * time.Hour)

	// Ensure TpIps is always initialized (even if empty)
	if row.TpIps == nil {
		row.TpIps = []string{}
	}

	// Set TpDestinationIP and TpSourceIP to non-nil pointers, even if empty
	emptyStr := ""
	if row.HttpRequest != nil {
		if row.HttpRequest.RemoteIp != "" {
			row.TpIps = append(row.TpIps, row.HttpRequest.RemoteIp)
			row.TpSourceIP = &row.HttpRequest.RemoteIp
		} else {
			row.TpSourceIP = &emptyStr
		}
		if row.HttpRequest.ServerIp != "" {
			row.TpIps = append(row.TpIps, row.HttpRequest.ServerIp)
			row.TpDestinationIP = &row.HttpRequest.ServerIp
		} else {
			row.TpDestinationIP = &emptyStr
		}
	} else {
		row.TpDestinationIP = &emptyStr
		row.TpSourceIP = &emptyStr
	}

	return row, nil
}

func (c *RequestsLogTable) GetDescription() string {
	return "GCP Request Logs track requests to Google Cloud services including application load balancer logs and Cloud Armor logs, capturing request events for security and compliance monitoring."
}
