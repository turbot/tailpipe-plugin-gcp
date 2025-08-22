---
title: "Source: gcp_logging_log_entry - Collect logs from GCP logging API"
description: "Allows users to collect logs from Google Cloud Platform (GCP) logging API."
---

# Source: gcp_logging_log_entry - Obtain logs from GCP logging API

The Google Cloud Platform (GCP) logging API provides access to various types of logs for GCP services. It allows you to view and manage logs for your GCP projects, including logs for administrative actions, data access, system events, request logs, and more.

Using this source, you can collect, filter, and analyze logs retrieved from the GCP logging API, enabling system monitoring, security investigations, and compliance reporting.

## Automatic Table Detection

The source automatically detects which table is being used based on the partition configuration. When you run a command like:

```bash
tpc gcp_requests_log.my_logs_requests --from T-30d
```

The source will:

1. Automatically detect the table name (`gcp_requests_log`) from the partition configuration
2. Automatically use the appropriate log types for that table
3. For `gcp_requests_log` table: uses `requests`, `cloud_run_request`, and `app_engine_request` logs
4. For `gcp_audit_log` table: uses `activity`, `data_access`, `system_event`, and `policy` logs

This means you don't need to specify log types manually - the source will automatically filter logs based on the table you're collecting for.

## Example Configurations

### Collect all types of logs

Collect all types of logs for a project.

```hcl
connection "gcp" "my_project" {
  project = "my-gcp-project"
}

partition "gcp_audit_log" "my_logs" {
  source "gcp_logging_log_entry" {
    connection = connection.gcp.my_project
  }
}
```

### Collect specific types of logs

Collect admin activity and data access logs for a project.

```hcl
partition "gcp_audit_log" "my_logs_admin_data_access" {
  source "gcp_logging_log_entry" {
    connection = connection.gcp.my_project
    log_types = ["activity", "data_access"]
  }
}
```

### Collect request logs

Collect request logs for Cloud Armor and Application Load Balancer analysis.

```hcl
partition "gcp_requests_log" "my_request_logs" {
  source "gcp_logging_log_entry" {
    connection = connection.gcp.my_project
    log_types = ["requests"]
  }
}
```

### Collect specific types of request logs

Collect specific types of request logs.

```hcl
partition "gcp_requests_log" "my_request_logs" {
  source "gcp_logging_log_entry" {
    connection = connection.gcp.my_project
    log_types = ["requests", "cloud_run_request"]
  }
}
```

### Collect specific log type

Collect only activity logs.

```hcl
partition "gcp_audit_log" "my_activity_logs" {
  source "gcp_logging_log_entry" {
    connection = connection.gcp.my_project
    log_types = ["activity"]
  }
}
```

## Arguments

| Argument   | Type             | Required | Default                  | Description                                                                                                                                                                                                        |
| ---------- | ---------------- | -------- | ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| connection | `connection.gcp` | No       | `connection.gcp.default` | The [GCP connection](https://hub.tailpipe.io/plugins/turbot/gcp#connection-credentials) to use to connect to the GCP account.                                                                                      |
| log_types  | List(String)     | No       | []                       | A list of log types to retrieve. If no types are specified, all log types for the table are retrieved. Valid values: activity, data_access, system_event, policy, cloud_run_request, app_engine_request, requests. |
