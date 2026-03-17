package requests_log

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
			SourceName: logging_log_entry.LoggingLogEntrySourceIdentifier,
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
	row.TpIngestTimestamp = time.Now()

	// Use Timestamp, fallback to ReceiveTimestamp, or current time if both are zero
	timestamp := row.Timestamp
	if timestamp.IsZero() {
		timestamp = row.ReceiveTimestamp
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	row.TpTimestamp = timestamp
	row.TpDate = timestamp.Truncate(24 * time.Hour)

	// Ensure TpIps is initialized before appending
	if row.TpIps == nil {
		row.TpIps = []string{}
	}

	if row.HttpRequest != nil {
		if row.HttpRequest.RemoteIp != "" {
			row.TpIps = append(row.TpIps, row.HttpRequest.RemoteIp)
			row.TpSourceIP = &row.HttpRequest.RemoteIp
		}
		if row.HttpRequest.ServerIp != "" {
			row.TpIps = append(row.TpIps, row.HttpRequest.ServerIp)
			row.TpDestinationIP = &row.HttpRequest.ServerIp
		}
	}

	return row, nil
}

func (c *RequestsLogTable) GetDescription() string {
	return "GCP Request Logs track requests to Google Cloud services including application load balancer logs and Cloud Armor logs, capturing request events for security and compliance monitoring."
}
