---
title: "Tailpipe Table: gcp_billing_report - Query GCP billing exports"
description: "Query detailed GCP billing and cost data exported from BigQuery, including service charges, credits, adjustments, and resource-level spend."
---

# Table: gcp_billing_report - Query GCP billing exports

The `gcp_billing_report` table allows you to query detailed cost and usage data exported from Google Cloud Billing to BigQuery. This table provides resource-level breakdowns of GCP service costs, credits, adjustments, and pricing details, enabling cost analysis, budget tracking, and optimization insights across your GCP environment.

## Configure

Create a [partition](https://tailpipe.io/docs/manage/partition) for `gcp_billing_report`:

```sh
vi ~/.tailpipe/config/gcp.tpc
```

```hcl
connection "gcp" "billing_account" {
  project = "my-gcp-project"
}

partition "gcp_billing_report" "my_billing" {
  source "gcp_storage_bucket" {
    connection = connection.gcp.billing_account
    bucket     = "gcp-billing-export-bucket"
    # prefix  = "optional/prefix/"
  }
}
```

## Collect

[Collect](https://tailpipe.io/docs/manage/collection) billing data for all `gcp_billing_report` partitions:

```sh
tailpipe collect gcp_billing_report
```

Or for a single partition:

```sh
tailpipe collect gcp_billing_report.my_billing
```

## Query

**[See example queries for this table →](../gcp_billing_report/queries.md)**

### Example: Monthly cost by service

```sql
select
  invoice_month,
  service_description,
  sum(cost) as total_cost
from
  gcp_billing_report
where
  usage_start_time >= date_sub(current_date(), interval 1 year)
group by
  invoice_month, service_description
order by
  invoice_month, total_cost desc;
```

## Example Configurations

### Collect billing exports from a Storage bucket

Collect billing data exported to a GCS bucket by BigQuery.

```hcl
connection "gcp" "billing_account" {
  project = "my-gcp-project"
}

partition "gcp_billing_report" "my_billing" {
  source "gcp_storage_bucket" {
    connection = connection.gcp.billing_account
    bucket     = "gcp-billing-export-bucket"
    # prefix  = "optional/prefix/"
  }
}
```

### Collect billing exports from an Artifact

```hcl
partition "gcp_billing_report" "my_billing_artifact" {
  source "artifact" {
    path = "/path/to/billing-export.json"
  }
}
```

## Source Defaults

### gcp_storage_bucket

This table sets the following defaults for the [gcp_storage_bucket](https://hub.tailpipe.io/plugins/turbot/gcp/sources/gcp_storage_bucket#arguments):

| Argument      | Default |
|--------------|---------|
| file_layout   | `*/*.json` |

## Key Fields

| Field                   | Description                                                      |
|------------------------|------------------------------------------------------------------|
| billing_account_id      | The Cloud Billing account ID                                      |
| invoice_month           | Invoice year and month (YYYYMM)                                   |
| service_id              | The ID of the GCP service                                         |
| service_description     | The name of the GCP service                                       |
| sku_id                  | The ID of the SKU (resource type)                                 |
| sku_description         | Description of the SKU                                            |
| usage_start_time        | Start time of the usage window                                    |
| usage_end_time          | End time of the usage window                                      |
| cost                    | The cost of the usage before credits                              |
| currency                | The currency for the cost                                         |
| usage_amount            | The quantity of usage                                             |
| usage_unit              | The unit of usage (e.g., seconds, bytes)                          |
| credits                 | Array of credits applied to the usage                             |
| project                 | Project details (id, number, name, labels, ancestors)             |
| location                | Location details (region, country, zone)                          |
| tags                    | Array of resource tags                                            |
| export_time             | Time the record was exported from BigQuery                        |
| cost_type               | Type of cost (regular, tax, adjustment, etc.)                     |
| transaction_type        | Transaction type (GOOGLE, THIRD_PARTY_RESELLER, etc.)             |
| seller_name             | Name of the seller                                                |
| price                   | Pricing details (effective price, tier, unit, etc.)               |
| resource                | Resource details (global name, name)                              |
| adjustment_info         | Adjustment details (id, description, type, mode)                  |
| subscription            | Subscription/commitment details                                   |

## References
- [GCP Billing Export to BigQuery - Standard Usage](https://cloud.google.com/billing/docs/how-to/export-data-bigquery-tables/standard-usage)
- [GCP Billing Export to BigQuery - Detailed Usage](https://cloud.google.com/billing/docs/how-to/export-data-bigquery-tables/detailed-usage) 