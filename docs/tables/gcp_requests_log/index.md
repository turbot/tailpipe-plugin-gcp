---
title: "Tailpipe Table: gcp_requests_log - Query GCP request logs"
description: "GCP request logs capture network requests from GCP Load Balancers containing fields for the results of Cloud Armor analysis."
---

# Table: gcp_requests_log - Query GCP request logs

The `gcp_requests_log` table allows you to query data from GCP audit logs. This table provides detailed information about API calls made within your Google Cloud environment, including the event name, resource affected, user identity, and more.

## Configure

Create a [partition](https://tailpipe.io/docs/manage/partition) for `gcp_requests_log`:

```sh
vi ~/.tailpipe/config/gcp.tpc
```

```hcl
connection "gcp" "logging_account" {
  project = "my-gcp-project"
}

partition "gcp_requests_log" "my_logs" {
  source "gcp_storage_bucket" {
    connection = connection.gcp.logging_account
  }
}
```
OR

```hcl
connection "gcp" "logging_account" {
  project = "my-gcp-project"
}

partition "gcp_requests_log" "my_logs" {
  source "gcp_logging_api" {
    connection = connection.gcp.logging_account
  }
}
```

## Collect

[Collect](https://tailpipe.io/docs/manage/collection) logs for all `gcp_requests_log` partitions:

```sh
tailpipe collect gcp_requests_log
```

Or for a single partition:

```sh
tailpipe collect gcp_requests_log.my_logs
```

## Query


### Blocked Requests

Count how many requests were blocked by Cloud Armor across all policies

```sql
SELECT
  enforced_security_policy.name AS policy_name,
  count(*) AS total_blocked_requests
FROM
  gcp_requests_log
WHERE
  enforced_security_policy ->> 'outcome' = 'DENY'
GROUP BY
  policy_name
ORDER BY
  total_blocked_requests DESC
```

### Top 10 events

List the 10 most blocked OWASP Core Rule Set rules

```sql
SELECT
  enforced_security_policy ->> 'preconfigured_expr_id' AS rule_id,
  COUNT(*) AS total_occurrences
FROM
  gcp_requests_log
WHERE
  enforced_security_policy ->> 'outcome' = 'DENY'
  AND LENGTH(enforced_security_policy ->> 'preconfigured_expr_id') > 0
GROUP BY
  rule_id
ORDER BY
  total_occurrences DESC
LIMIT 10
```


## Example Configurations

### Collect logs from a Storage bucket

Collect request logs stored in a Storage bucket that use the [default log file name format](https://hub.tailpipe.io/plugins/turbot/gcp/tables/gcp_audit_log#gcp_storage_bucket).

```hcl
connection "gcp" "logging_account" {
  project = "my-gcp-project"
}

partition "gcp_requests_log" "my_logs" {
  source "gcp_storage_bucket" {
    connection = connection.gcp.logging_account
    bucket     = "gcp-cloudarmor-logs-bucket"
  }
}
```

### Collect logs from a Storage bucket with a prefix

Collect audit logs stored with a GCS key prefix.

```hcl
partition "gcp_requests_log" "my_logs_prefix" {
  source "gcp_storage_bucket" {
    connection = connection.gcp.logging_account
    bucket     = "gcp-cloudarmor-logs-bucket"
    prefix     = "my/prefix/"
  }
}
```

### Collect logs from a Storage Bucket for a single project

Collect audit logs for a specific project.

```hcl
partition "gcp_requests_log" "my_logs_prefix" {
  filter = "log_name like 'projects/my-project-name/logs/requests/%'"

  source "gcp_storage_bucket" {
    connection = connection.gcp.logging_account
    bucket     = "gcp-cloudarmor-logs-bucket"
  }
}
```

### Collect logs from Cloud Logging API

Collect request logs directly via the Cloud Logging API.  *Note that rate limiting is currently not implemented and this could impact ability to collect a large number of logs*

```hcl
connection "gcp" "my_project" {
  project = "my-gcp-project"
}

partition "gcp_requests_log" "my_logs" {
  source "gcp_logging_api" {
    connection = connection.gcp.my_project
  }
}
```

### Collect other types of logs from Cloud Logging API

The Cloud Logging API source can be expanded upon to retrieve logs other than requests, if the appropriate table is created to enrich and save them.  The log name attribute is the filter used for this, and it assumes that a table / partition have been made that match the data type.

Example: Collecting GCP Dataflow logs

```hcl
partition "gcp_dataflow_log" "my_logs_prefix" {
  filter = "log_name like 'projects/my-project-name/logs/dataflow.googleapis.com%'"
  source "gcp_logging_api" {
    connection = connection.gcp.my_project
  }
}
```

## Source Defaults

### gcp_storage_bucket

This table sets the following defaults for the [gcp_storage_bucket](https://hub.tailpipe.io/plugins/turbot/gcp/sources/gcp_storage_bucket#arguments):

| Argument      | Default |
|--------------|---------|
| file_layout   | `requests/%{YEAR:year}/%{MONTHNUM:month}/%{MONTHDAY:day}/%{HOUR:hour}:%{MINUTE:minute}:%{SECOND:second}_%{DATA:end_time}_%{DATA:suffix}.json` |
