package billing_report

import (
	"time"

	"github.com/rs/xid"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/tailpipe-plugin-gcp/sources/storage_bucket"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source_config"
	"github.com/turbot/tailpipe-plugin-sdk/constants"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/schema"
	"github.com/turbot/tailpipe-plugin-sdk/table"
)

const BillingReportTableIdentifier = "gcp_billing_report"

type BillingReportTable struct{}

func (t *BillingReportTable) Identifier() string {
	return BillingReportTableIdentifier
}

func (t *BillingReportTable) GetSourceMetadata() ([]*table.SourceMetadata[*BillingReport], error) {
	defaultGCSArtifactConfig := &artifact_source_config.ArtifactSourceConfigImpl{
		// Pattern for BigQuery JSON export files (customize as needed)
		FileLayout: utils.ToStringPointer("%{DATA:file_name}.json.gz"),
	}

	return []*table.SourceMetadata[*BillingReport]{
		{
			SourceName: storage_bucket.GcpStorageBucketSourceIdentifier,
			Mapper:     NewBillingReportMapper(),
			Options: []row_source.RowSourceOption{
				artifact_source.WithDefaultArtifactSourceConfig(defaultGCSArtifactConfig),
				artifact_source.WithRowPerLine(),
			},
		},
		{
			SourceName: constants.ArtifactSourceIdentifier,
			Mapper:     NewBillingReportMapper(),
			Options: []row_source.RowSourceOption{
				artifact_source.WithRowPerLine(),
			},
		},
	}, nil
}

func (t *BillingReportTable) EnrichRow(row *BillingReport, sourceEnrichmentFields schema.SourceEnrichment) (*BillingReport, error) {
	row.CommonFields = sourceEnrichmentFields.CommonFields

	row.TpID = xid.New().String()
	row.TpIngestTimestamp = time.Now()

	if row.UsageStartTime != nil {
		row.TpTimestamp = *row.UsageStartTime
		row.TpDate = row.UsageStartTime.Truncate(24 * time.Hour)
	} else if row.UsageEndTime != nil {
		row.TpTimestamp = *row.UsageEndTime
		row.TpDate = row.UsageEndTime.Truncate(24 * time.Hour)
	}

	// TpIndex: Use Project.ID if available, else fallback
	if row.Project != nil && row.Project.ID != nil {
		row.TpIndex = typehelpers.SafeString(row.Project.ID)
	} else {
		row.TpIndex = schema.DefaultIndex
	}

	return row, nil
}

func (t *BillingReportTable) GetDescription() string {
	return "Google Cloud Billing Reports exported from BigQuery provide a detailed, resource-level breakdown of GCP service costs and usage. This table enables cost analysis, budget tracking, and optimization insights across GCP projects, including service charges, credits, adjustments, and pricing details."
}
