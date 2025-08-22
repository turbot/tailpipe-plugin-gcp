---
title: "Tailpipe Table: gcp_requests_log - Query GCP request logs"
description: "GCP request logs capture HTTP(S) request flows with security policy enforcement details, threat intelligence, and comprehensive request/response metadata."
---

# Table: gcp_requests_log - Query GCP request logs

The `gcp_requests_log` table allows you to query data from GCP request logs. This table provides detailed information about HTTP(S) request flows with security policy enforcement details, threat intelligence, and comprehensive request/response metadata.

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

**[Explore 12+ example queries for this table →](https://hub.tailpipe.io/plugins/turbot/gcp/queries/gcp_requests_log)**

### Blocked requests by Cloud Armor

Find requests that were blocked by Cloud Armor security policies.

```sql
select
  timestamp,
  http_request.remote_ip,
  http_request.url,
  security_policy.action,
  security_policy.rule_name,
  threat_info.threat_type
from
  gcp_requests_log
where
  security_policy.action = 'deny'
order by
  timestamp desc;
```

### Threat pattern analysis

Analyze threat patterns and their frequency.

```sql
select
  threat_info.threat_type,
  threat_info.threat_category,
  threat_info.threat_severity,
  count(*) as request_count
from
  gcp_requests_log
where
  threat_info.threat_type is not null
group by
  threat_info.threat_type,
  threat_info.threat_category,
  threat_info.threat_severity
order by
  request_count desc;
```

### Cache performance monitoring

Monitor cache hit rates and performance metrics.

```sql
select
  date_trunc('hour', timestamp) as hour,
  cache_info.cache_hit,
  count(*) as request_count,
  avg(cast(replace(http_request.latency, 's', '') as float)) as avg_latency
from
  gcp_requests_log
where
  cache_info.cache_lookup = true
group by
  hour,
  cache_info.cache_hit
order by
  hour desc;
```

### High-latency requests

Find requests with high latency for performance analysis.

```sql
select
  timestamp,
  http_request.remote_ip,
  http_request.url,
  http_request.latency,
  verbose_logging.backend_latency,
  verbose_logging.load_balancer_latency
from
  gcp_requests_log
where
  cast(replace(http_request.latency, 's', '') as float) > 1.0
order by
  timestamp desc
limit 100;
```

### Security policy rule analysis

Analyze security policy rule effectiveness and preview mode usage.

```sql
select
  security_policy.rule_name,
  security_policy.rule_type,
  security_policy.action,
  count(*) as matches,
  count(case when security_policy.preview_mode = true then 1 end) as preview_matches
from
  gcp_requests_log
where
  security_policy.rule_evaluated = true
group by
  security_policy.rule_name,
  security_policy.rule_type,
  security_policy.action
order by
  matches desc;
```

## Example Configurations

### Collect logs from a Storage bucket

Collect request logs stored in a Storage bucket that use the [default log file name format](https://hub.tailpipe.io/plugins/turbot/gcp/tables/gcp_requests_log#gcp_storage_bucket).

```hcl
connection "gcp" "logging_account" {
  project = "my-gcp-project"
}

partition "gcp_requests_log" "my_logs" {
  source "gcp_storage_bucket" {
    connection = connection.gcp.logging_account
    bucket     = "gcp-request-logs-bucket"
  }
}
```

### Collect logs from a Storage bucket with a prefix

Collect request logs stored with a GCS key prefix.

```hcl
partition "gcp_requests_log" "my_logs_prefix" {
  source "gcp_storage_bucket" {
    connection = connection.gcp.logging_account
    bucket     = "gcp-request-logs-bucket"
    prefix     = "requests/"
  }
}
```

### Collect logs from a Storage Bucket for a single project

Collect request logs for a specific project.

```hcl
partition "gcp_requests_log" "my_logs_prefix" {
  filter = "log_name like 'projects/my-project-name/logs/requests/%'"

  source "gcp_storage_bucket" {
    connection = connection.gcp.logging_account
    bucket     = "gcp-request-logs-bucket"
  }
}
```

### Collect logs from logging API

Collect request logs using the GCP logging API.

```hcl
connection "gcp" "my_project" {
  project = "my-gcp-project"
}

partition "gcp_requests_log" "my_logs" {
  source "logging_log_entry" {
    connection = connection.gcp.my_project
  }
}
```

### Collect specific types of request logs from logging API

Collect specific types of request logs for a project.

```hcl
partition "gcp_requests_log" "my_logs_requests" {
  source "logging_log_entry" {
    connection = connection.gcp.my_project
    log_types = ["requests"]
  }
}
```

### Filter for security events

Use the filter argument in your partition to only include security-related events.

```hcl
partition "gcp_requests_log" "security_events" {
  # Only include security-related events
  filter = "security_policy.action = 'deny' or threat_info.threat_type is not null"

  source "logging_log_entry" {
    connection = connection.gcp.my_project
    log_types = ["requests"]
  }
}
```

## Source Defaults

### gcp_storage_bucket

This table sets the following defaults for the [gcp_storage_bucket](https://hub.tailpipe.io/plugins/turbot/gcp/sources/gcp_storage_bucket#arguments):

| Argument    | Default                                                                                                                                       |
| ----------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| file_layout | `requests/%{YEAR:year}/%{MONTHNUM:month}/%{MONTHDAY:day}/%{HOUR:hour}:%{MINUTE:minute}:%{SECOND:second}_%{DATA:end_time}_%{DATA:suffix}.json` |
