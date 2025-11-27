---
title: "Tailpipe Table: gcp_billing_report - Query GCP Billing Reports"
description: "GCP billing reports provide detailed cost and usage information for Google Cloud Platform resources, including costs, credits, usage metrics, and resource details."
---

# Table: gcp_billing_report - Query GCP Billing Reports

The `gcp_billing_report` table allows you to query data from GCP billing reports. This table provides detailed information about cloud costs, usage patterns, credits, and resource consumption across your Google Cloud environment.

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

**[Explore 10+ example queries for this table →](https://hub.tailpipe.io/plugins/turbot/gcp/queries/gcp_billing_report)**

### Daily costs

Get daily cost breakdown across all projects.

```sql
select
  date_trunc('day', usage_start_time) as billing_date,
  sum(cost) as total_cost,
  currency,
  count(*) as line_items
from
  gcp_billing_report
group by
  billing_date,
  currency
order by
  billing_date desc;
```

### Top 10 most expensive services

List the 10 services with the highest costs.

```sql
select
  service_description,
  sum(cost) as total_cost,
  currency,
  count(*) as usage_records
from
  gcp_billing_report
group by
  service_description,
  currency
order by
  total_cost desc
limit 10;
```

### Cost by project

Analyze costs grouped by project to identify high-spending areas.

```sql
select
  project_name,
  project_id,
  sum(cost) as total_cost,
  sum(cost_at_list) as list_cost,
  currency,
  count(*) as line_items
from
  gcp_billing_report
group by
  project_name,
  project_id,
  currency
order by
  total_cost desc;
```

### Credits analysis

Review available credits and their impact on costs. This query properly iterates through all credits in the array regardless of how many credits are applied to each line item.

```sql
with credit_expanded as (
  select
    project_name,
    currency,
    trim(json_extract(credits, '$[' || (i.generate_series - 1) || '].name'), '"') as credit_name,
    trim(json_extract(credits, '$[' || (i.generate_series - 1) || '].type'), '"') as credit_type,
    cast(json_extract(credits, '$[' || (i.generate_series - 1) || '].amount') as float) as credit_amount
  from
    gcp_billing_report,
    generate_series(1::bigint, greatest(cast(json_array_length(credits) as bigint), 1::bigint)) as i
  where
    credits is not null
    and json_array_length(credits) > 0
    and i.generate_series <= json_array_length(credits)
)
select
  project_name,
  credit_name,
  credit_type,
  sum(credit_amount) as total_credit_amount,
  currency
from
  credit_expanded
where
  credit_name is not null
group by
  project_name,
  credit_name,
  credit_type,
  currency
order by
  total_credit_amount desc;
```

## Example Configurations

### Collect billing data from a Storage bucket

Collect billing reports stored in a Storage bucket with the standard export format.

```hcl
connection "gcp" "billing_account" {
  project = "my-gcp-project"
}

partition "gcp_billing_report" "my_billing" {
  source "gcp_storage_bucket" {
    connection = connection.gcp.billing_account
    bucket     = "gcp-billing-export-bucket"
  }
}
```

### Collect billing data with a prefix

Collect billing reports stored with a GCS key prefix.

```hcl
partition "gcp_billing_report" "my_billing_prefix" {
  source "gcp_storage_bucket" {
    connection = connection.gcp.billing_account
    bucket     = "gcp-billing-export-bucket"
    prefix     = "billing-exports/"
  }
}
```

### Filter by specific project

Collect billing data for a specific project only.

```hcl
partition "gcp_billing_report" "single_project" {
  filter = "project_id = 'my-specific-project'"

  source "gcp_storage_bucket" {
    connection = connection.gcp.billing_account
    bucket     = "gcp-billing-export-bucket"
  }
}
```

### Filter by cost threshold

Only collect billing records above a certain cost threshold to focus on significant expenses.

```hcl
partition "gcp_billing_report" "high_cost_only" {
  filter = "cost > 10.0"

  source "gcp_storage_bucket" {
    connection = connection.gcp.billing_account
    bucket     = "gcp-billing-export-bucket"
  }
}
```

### Filter by date range

Collect billing data for a specific time period.

```hcl
partition "gcp_billing_report" "monthly_billing" {
  filter = "usage_start_time >= '2024-01-01' AND usage_start_time < '2024-02-01'"

  source "gcp_storage_bucket" {
    connection = connection.gcp.billing_account
    bucket     = "gcp-billing-export-bucket"
  }
}
```

## Source Defaults

### gcp_storage_bucket

This table sets the following defaults for the [gcp_storage_bucket](https://hub.tailpipe.io/plugins/turbot/gcp/sources/gcp_storage_bucket#arguments):

| Argument    | Default                     |
| ----------- | --------------------------- |
| file_layout | `%{DATA:file_name}.json.gz` |
