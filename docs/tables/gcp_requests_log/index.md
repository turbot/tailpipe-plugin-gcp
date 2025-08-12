---
title: "Table: gcp_requests_log - Cloud Armor Request Logs"
description: "Query Cloud Armor-augmented Application Load Balancer request logs for security monitoring and threat detection."
---

# Table: gcp_requests_log

Query Cloud Armor-augmented Application Load Balancer request logs to analyze HTTP(S) request flows with security policy enforcement details, threat intelligence, and comprehensive request/response metadata.

This table captures detailed information about requests processed through Google Cloud Load Balancers with Cloud Armor security policies, including:

- **HTTP Request Details**: Method, URL, headers, status codes, and timing information
- **Security Policy Information**: Cloud Armor policy actions, rule evaluations, and enforcement decisions
- **Threat Intelligence**: Threat types, categories, severity, and confidence levels
- **Cache Information**: Cache hit/miss status and performance metrics
- **Verbose Logging**: Detailed request/response headers and body content (when enabled)

## Example Queries

### Find blocked requests by Cloud Armor

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

### Analyze threat patterns

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

### Monitor cache performance

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

### Find high-latency requests

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

### Collect request logs from audit log API

Collect request logs for a project using the audit log API source.

```hcl
connection "gcp" "my_project" {
  project = "my-gcp-project"
}

partition "gcp_requests_log" "my_request_logs" {
  source "gcp_audit_log_api" {
    connection = connection.gcp.my_project
    log_types = ["requests"]
  }
}
```

### Collect request logs from storage bucket

Collect request logs stored in a GCP Storage bucket.

```hcl
connection "gcp" "logging_account" {
  project = "my-gcp-project"
}

partition "gcp_requests_log" "my_request_logs" {
  source "gcp_storage_bucket" {
    connection = connection.gcp.logging_account
    bucket     = "gcp-request-logs-bucket"
    prefix     = "requests/"
  }
}
```

### Filter for specific security events

Filter the partition to only include security-related events.

```hcl
partition "gcp_requests_log" "security_events" {
  filter = "security_policy.action = 'deny' or threat_info.threat_type is not null"

  source "gcp_audit_log_api" {
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

## Column Descriptions

| Column                | Type        | Description                                                                               |
| --------------------- | ----------- | ----------------------------------------------------------------------------------------- |
| `timestamp`           | `timestamp` | The date and time when the request occurred, in ISO 8601 format.                          |
| `log_name`            | `text`      | The name of the log that recorded the request.                                            |
| `insert_id`           | `text`      | A unique identifier for the log entry, used to prevent duplicate log entries.             |
| `severity`            | `text`      | The severity level of the log entry (e.g., 'INFO', 'WARNING', 'ERROR', 'CRITICAL').       |
| `trace`               | `text`      | The unique trace ID associated with the request, used for distributed tracing.            |
| `trace_sampled`       | `boolean`   | Indicates whether the request trace was sampled for analysis (true or false).             |
| `span_id`             | `text`      | The span ID for the request, used in distributed tracing to identify specific operations. |
| `http_request`        | `jsonb`     | Details about the HTTP request associated with the log entry.                             |
| `security_policy`     | `jsonb`     | Cloud Armor security policy information and enforcement details.                          |
| `threat_info`         | `jsonb`     | Threat intelligence information related to the request.                                   |
| `cache_info`          | `jsonb`     | Cache-related information for the request.                                                |
| `verbose_logging`     | `jsonb`     | Verbose logging information including headers and body content.                           |
| `labels`              | `jsonb`     | Key-value labels associated with the log entry for filtering and analysis.                |
| `tp_id`               | `text`      | A unique identifier for the row.                                                          |
| `tp_timestamp`        | `timestamp` | The timestamp when the row was created.                                                   |
| `tp_ingest_timestamp` | `timestamp` | The timestamp when the row was ingested.                                                  |
| `tp_date`             | `date`      | The date when the row was created.                                                        |
| `tp_source_name`      | `text`      | The name of the source that provided the data.                                            |
| `tp_source_type`      | `text`      | The type of the source that provided the data.                                            |
| `tp_source_location`  | `text`      | The location of the source that provided the data.                                        |
| `tp_ips`              | `text[]`    | Array of IP addresses extracted from the log entry.                                       |
| `tp_source_ip`        | `text`      | The source IP address extracted from the log entry.                                       |
| `tp_destination_ip`   | `text`      | The destination IP address extracted from the log entry.                                  |
