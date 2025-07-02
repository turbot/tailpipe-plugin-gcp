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
  source "gcp_cloud_logging_api" {
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

Count how many requests were blocked by Cloud Armor

```sql
> SELECT
  enforced_security_policy ->> 'name' AS policy_name,
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

List the 10 most frequently called events.

```sql
select
  service_name,
  method_name,
  count(*) as event_count
from
  gcp_audit_log
group by
  service_name,
  method_name
order by
  event_count desc
limit 10;
```

### High Volume IAM Access Token Generation

Find users generating a high volume of IAM access tokens within a short period, which may indicate potential privilege escalation or compromised credentials.

```sql
select
  authentication_info ->> 'principal_email' as user_email,
  count(*) as event_count,
  date_trunc('minute', timestamp) as event_minute
from
  gcp_audit_log
where
  service_name = 'iamcredentials.googleapis.com'
  and method_name ilike 'generateaccesstoken'
group by
  user_email,
  event_minute
having
  count(*) > 10
order by
  event_count desc;
```

## Example Configurations

### Collect logs from a Storage bucket

Collect audit logs stored in a Storage bucket that use the [default log file name format](https://hub.tailpipe.io/plugins/turbot/gcp/tables/gcp_audit_log#gcp_storage_bucket).

```hcl
connection "gcp" "logging_account" {
  project = "my-gcp-project"
}

partition "gcp_audit_log" "my_logs" {
  source "gcp_storage_bucket" {
    connection = connection.gcp.logging_account
    bucket     = "gcp-audit-logs-bucket"
  }
}
```

### Collect logs from a Storage bucket with a prefix

Collect audit logs stored with a GCS key prefix.

```hcl
partition "gcp_audit_log" "my_logs_prefix" {
  source "gcp_storage_bucket" {
    connection = connection.gcp.logging_account
    bucket     = "gcp-audit-logs-bucket"
    prefix     = "my/prefix/"
  }
}
```

### Collect logs from a Storage Bucket for a single project

Collect audit logs for a specific project.

```hcl
partition "gcp_audit_log" "my_logs_prefix" {
  filter = "log_name like 'projects/my-project-name/logs/cloudaudit.googleapis.com/%'"

  source "gcp_storage_bucket" {
    connection = connection.gcp.logging_account
    bucket     = "gcp-audit-logs-bucket"
  }
}
```

### Collect logs from audit logs API

Collect audit logs stored in a Storage bucket that use the [default log file name format](https://hub.tailpipe.io/plugins/turbot/gcp/tables/gcp_audit_log#gcp_storage_bucket).

```hcl
connection "gcp" "my_project" {
  project = "my-gcp-project"
}

partition "gcp_audit_log" "my_logs" {
  source "gcp_audit_log_api" {
    connection = connection.gcp.my_project
  }
}
```

### Collect specific types of audit logs from audit logs API

Collect admin activity and data access audit logs for a project.

```hcl
partition "gcp_audit_log" "my_logs_admin_data_access" {
  source "gcp_audit_log_api" {
    connection = connection.gcp.my_project
    log_types = ["activity", "data_access"]
  }
}
```

### Exclude INFO level events

Use the filter argument in your partition to exclude INFO severity level events and reduce log storage size.

```hcl
partition "gcp_audit_log" "my_logs_severity" {
  # Avoid saving specific severity levels
  filter = "severity != 'INFO'"

  source "gcp-storage_bucket" {
    connection = connection.gcp.logging_account
    bucket     = "gcp-audit-logs-bucket"
  }
}
```

## Source Defaults

### gcp_storage_bucket

This table sets the following defaults for the [gcp_storage_bucket](https://hub.tailpipe.io/plugins/turbot/gcp/sources/gcp_storage_bucket#arguments):

| Argument      | Default |
|--------------|---------|
| file_layout   | `cloudaudit.googleapis.com/%{DATA:type}/%{YEAR:year}/%{MONTHNUM:month}/%{MONTHDAY:day}/%{HOUR:hour}:%{MINUTE:minute}:%{SECOND:second}_%{DATA:end_time}_%{DATA:suffix}.json` |
