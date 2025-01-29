---
title: "Tailpipe Table: gcp_audit_log - Query GCP Audit Logs"
description: "GCP Audit Logs capture API activity, administrative actions, and security-related events within your Google Cloud environment."
---

# Table: gcp_audit_log - Query GCP Audit Logs

The `gcp_audit_log` table allows you to query data from GCP Audit Logs. This table provides detailed information about API calls made within your Google Cloud environment, including the event name, resource affected, user identity, and more.

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
  source "gcp_logging" {
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

**[Explore 100+ example queries for this table →](https://hub.tailpipe.io/plugins/turbot/gcp/queries/gcp_audit_log)**

### User Activity

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
  authentication_info ->> 'principal_email' = 'jane_doe@domain.com'
order by
  timestamp desc;
```

### Top 10 Events

List the 10 most frequently called events.

```sql
select
  service_name as event_source,
  method_name as event_name,
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

### High Volume IAM Policy Changes

Find users generating a high volume of IAM policy changes to detect potential privilege escalations.

```sql
select
  authentication_info ->> 'principal_email' as user_email,
  count(*) as event_count,
  date_trunc('minute', timestamp) as event_minute
from
  gcp_audit_log
where
  service_name = 'iam.googleapis.com'
  and method_name like '%setiampolicy%'
group by
  user_email,
  event_minute
having
  count(*) > 50
order by
  event_count desc;
```

## Example Configurations

### Collect logs from GCP Logging

Collect GCP Audit Logs using GCP Logging:

```hcl
partition "gcp_audit_log" "my_logs" {
  source "gcp_logging" {
    connection = connection.gcp.logging_account
  }
}
```

### Exclude Read-Only Events

Use the filter argument in your partition to exclude read-only events and reduce log storage size.

```hcl
partition "gcp_audit_log" "my_logs_write" {
  # Avoid saving read-only events
  filter = "severity != 'INFO'"

  source "gcp_logging" {
    connection = connection.gcp.logging_account
  }
}
```

### Collect logs for all projects in an organization

For a specific organization, collect logs for all projects.

```hcl
partition "gcp_audit_log" "my_logs_org" {
  source "gcp_logging"  {
    connection  = connection.gcp.logging_account
  }
}
```

### Collect logs for a single project

For a specific project, collect logs for all resources.

```hcl
partition "gcp_audit_log" "my_logs_project" {
  source "gcp_logging"  {
    connection  = connection.gcp.logging_account
    project = "my-gcp-project"
  }
}
```

## Source Defaults

### gcp_logging

This table sets the following defaults for the [gcp_logging source](https://tailpipe.io/plugins/turbot/gcp/sources/gcp_logging#arguments):

| Argument      | Default |
|--------------|---------|
| log_type     | `cloudaudit.googleapis.com/activity` |

