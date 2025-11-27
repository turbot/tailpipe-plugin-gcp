---
title: "Tailpipe Table: gcp_audit_log - Query GCP audit logs"
description: "GCP audit logs capture API activity, administrative actions, and security-related events within your Google Cloud environment."
---

# Table: gcp_audit_log - Query GCP audit logs

The `gcp_audit_log` table allows you to query data from GCP audit logs. This table provides detailed information about API calls made within your Google Cloud environment, including the event name, resource affected, user identity, and more.

## Configure

Create a [partition](https://tailpipe.io/docs/manage/partition) for `gcp_audit_log`:

```sh
vi ~/.tailpipe/config/gcp.tpc
```

```hcl
connection "gcp" "logging_account" {
  project = "my-gcp-project"
}

partition "gcp_audit_log" "my_logs" {
  source "gcp_storage_bucket" {
    connection = connection.gcp.logging_account
  }
}
```

## Collect

[Collect](https://tailpipe.io/docs/manage/collection) logs for all `gcp_audit_log` partitions:

```sh
tailpipe collect gcp_audit_log
```

Or for a single partition:

```sh
tailpipe collect gcp_audit_log.my_logs
```

## Query

**[Explore 50+ example queries for this table →](https://hub.tailpipe.io/plugins/turbot/gcp/queries/gcp_audit_log)**

### User activity

Find any actions taken by a user.

```sql
select
  timestamp,
  method_name,
  service_name,
  authentication_info ->> 'principal_email' as user_email,
  resource_name
from
  gcp_audit_log
where
  (authentication_info ->> 'principal_email') = 'jane_doe@domain.com'
order by
  timestamp desc;
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

### Collect logs from logging API

Collect audit logs using the GCP logging API.

```hcl
connection "gcp" "my_project" {
  project = "my-gcp-project"
}

partition "gcp_audit_log" "my_logs" {
  source "gcp_logging_api" {
    connection = connection.gcp.my_project
  }
}
```

### Collect specific types of audit logs from logging API

Collect admin activity and data access audit logs for a project.

```hcl
partition "gcp_audit_log" "my_logs_admin_data_access" {
  source "gcp_logging_api" {
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

| Argument    | Default                                                                                                                                                                     |
| ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| file_layout | `cloudaudit.googleapis.com/%{DATA:type}/%{YEAR:year}/%{MONTHNUM:month}/%{MONTHDAY:day}/%{HOUR:hour}:%{MINUTE:minute}:%{SECOND:second}_%{DATA:end_time}_%{DATA:suffix}.json` |
