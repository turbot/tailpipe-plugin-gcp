package billing_report

import (
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/tailpipe-plugin-gcp/sources/storage_bucket"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source_config"
	"github.com/turbot/tailpipe-plugin-sdk/constants"
	"github.com/turbot/tailpipe-plugin-sdk/formats"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/schema"
	"github.com/turbot/tailpipe-plugin-sdk/table"
	"github.com/turbot/tailpipe-plugin-sdk/types"
)

const BillingReportTableIdentifier = "gcp_billing_report"

type BillingReportTable struct {
	table.CustomTableImpl
}

func (t *BillingReportTable) Identifier() string {
	return BillingReportTableIdentifier
}

func (t *BillingReportTable) GetSourceMetadata() ([]*table.SourceMetadata[*types.DynamicRow], error) {
	// Cloud Billing export to Cloud Storage does not include a default directory prefix inside your bucket. The files are saved directly at the root level.
	// https://groups.google.com/g/gce-discussion/c/4G_pLvcLSAA
	defaultStorageBucketArtifactConfig := &artifact_source_config.ArtifactSourceConfigImpl{
		FileLayout: utils.ToStringPointer("%{DATA:file_name}.json.gz"),
	}

	return []*table.SourceMetadata[*types.DynamicRow]{
		{
			SourceName: storage_bucket.GcpStorageBucketSourceIdentifier,
			Options: []row_source.RowSourceOption{
				artifact_source.WithDefaultArtifactSourceConfig(defaultStorageBucketArtifactConfig),
				artifact_source.WithRowPerLine(),
			},
		},
		{
			SourceName: constants.ArtifactSourceIdentifier,
			Options:    []row_source.RowSourceOption{},
		},
	}, nil
}

func (t *BillingReportTable) GetDefaultFormat() formats.Format {
	return formats.NewJsonLines()
}

func (t *BillingReportTable) GetTableDefinition() *schema.TableSchema {
	return &schema.TableSchema{
		Name: BillingReportTableIdentifier,
		Columns: []*schema.ColumnSchema{
			{
				ColumnName: "tp_timestamp",
				Type:       "timestamp",
				SourceName: "usage_start_time",
			},
			{
				ColumnName: "billing_account_id",
				Type:       "varchar",
			},
			{
				ColumnName: "cost",
				Type:       "float",
			},
			{
				ColumnName: "cost_at_list",
				Type:       "float",
			},
			{
				ColumnName: "cost_type",
				Type:       "varchar",
			},
			{
				ColumnName: "credits",
				Type:       "json",
				StructFields: []*schema.ColumnSchema{
					{
						ColumnName: "amount",
						Type:       "float",
					},
					{
						ColumnName: "id",
						Type:       "varchar",
					},
					{
						ColumnName: "name",
						Type:       "varchar",
					},
					{
						ColumnName: "type",
						Type:       "varchar",
					},
				},
			},
			{
				ColumnName: "currency",
				Type:       "varchar",
			},
			{
				ColumnName: "currency_conversion_rate",
				Type:       "float",
			},
			{
				ColumnName: "export_time",
				Type:       "timestamp",
				Transform:  "strptime(export_time, '%Y-%m-%d %H:%M:%S %Z')",
			},
			{
				ColumnName: "invoice_month",
				Type:       "integer",
				Transform:  "(invoice ->> 'month')::integer",
			},
			{
				ColumnName: "labels",
				Type:       "json",
			},
			{
				ColumnName: "location",
				Type:       "struct",
				StructFields: []*schema.ColumnSchema{
					{
						ColumnName: "country",
						Type:       "varchar",
					},
					{
						ColumnName: "location",
						Type:       "varchar",
					},
					{
						ColumnName: "region",
						Type:       "varchar",
					},
					{
						ColumnName: "zone",
						Type:       "varchar",
					},
				},
			},
			{
				ColumnName: "project_ancestors",
				Type:       "json",
				Transform:  "(project ->> 'ancestors')::json",
			},
			{
				ColumnName: "project_ancestry_numbers",
				Type:       "varchar",
				Transform:  "(project ->> 'ancestry_numbers')",
			},
			{
				ColumnName: "project_id",
				Type:       "varchar",
				Transform:  "(project ->> 'id')",
			},
			{
				ColumnName: "project_labels",
				Type:       "json",
				Transform:  "(project ->> 'labels')::json",
			},
			{
				ColumnName: "project_name",
				Type:       "varchar",
				Transform:  "(project ->> 'name')",
			},
			{
				ColumnName: "project_number",
				Type:       "varchar",
				Transform:  "(project ->> 'number')",
			},
			{
				ColumnName: "service_description",
				Type:       "varchar",
				Transform:  "(service ->> 'description')",
			},
			{
				ColumnName: "service_id",
				Type:       "varchar",
				Transform:  "(service ->> 'id')",
			},
			{
				ColumnName: "sku_description",
				Type:       "varchar",
				Transform:  "(sku ->> 'description')",
			},
			{
				ColumnName: "sku_id",
				Type:       "varchar",
				Transform:  "(sku ->> 'id')",
			},
			{
				ColumnName: "system_labels",
				Type:       "json",
			},
			{
				ColumnName: "transaction_type",
				Type:       "varchar",
			},
			{
				ColumnName: "usage_end_time",
				Type:       "timestamp",
				Transform:  "strptime(usage_end_time, '%Y-%m-%d %H:%M:%S %Z')",
			},
			{
				ColumnName: "usage_start_time",
				Type:       "timestamp",
				Transform:  "strptime(usage_start_time, '%Y-%m-%d %H:%M:%S %Z')",
			},
			{
				ColumnName: "usage",
				Type:       "struct",
				StructFields: []*schema.ColumnSchema{
					{
						ColumnName: "amount",
						Type:       "float",
					},
					{
						ColumnName: "amount_in_pricing_units",
						Type:       "float",
					},
					{
						ColumnName: "pricing_unit",
						Type:       "varchar",
					},
					{
						ColumnName: "unit",
						Type:       "varchar",
					},
				},
			},
		},
		Description: t.GetDescription(),
	}
}

func (t *BillingReportTable) GetDescription() string {
	return "GCP Billing Reports provide detailed cost and usage information for Google Cloud Platform resources. This table includes billing data such as costs, credits, usage metrics, and resource details, helping teams monitor spending, analyze cost trends, and optimize cloud resource usage across projects and services."
}
