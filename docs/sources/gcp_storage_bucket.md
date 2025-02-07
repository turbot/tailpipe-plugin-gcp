---
title: "Source: gcp_storage_bucket - Collect logs from GCP Storage buckets"
description: "Allows users to collect logs from GCP Storage buckets."
---

# Source: gcp_storage_bucket - Collect logs from GCP Storage buckets

A GCP Storage bucket is a cloud storage resource used to store objects like data files and metadata. It serves as a central repository for logs from GCP services such as audit logs, VPC Flow Logs, Cloud Functions, and more.

Using this source, you can collect, filter, and analyze logs stored in GCP Storage buckets, enabling system monitoring, security investigations, and compliance reporting.

Most GCP tables define a default `file_path` for the `gcp_storage_bucket` source, so if your GCP logs are stored in default log locations, you don't need to override the `file_path` argument.

## Example Configurations

### Collect audit logs

Collect audit logs for all projects.

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

### Collect audit logs with a prefix

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

### Collect audit logs for a single project

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

## Arguments

| Argument    | Type             | Required | Default                  | Description                                                                                                                   |
|-------------|------------------|----------|--------------------------|-------------------------------------------------------------------------------------------------------------------------------|
| bucket      | String           | Yes      |                          | The name of the GCP Storage bucket to collect logs from.                                                                      |
| connection  | `connection.gcp` | No       | `connection.gcp.default` | The [GCP connection](https://hub.tailpipe.io/plugins/turbot/gcp#connection-credentials) to use to connect to the GCP account. |
| file_layout | String           | No       |                          | The Grok pattern that defines the log file structure.                                                                         |
| prefix      | String           | No       |                          | The GCS key prefix that comes after the name of the bucket you have designated for log file delivery.                         |

### Table Defaults

The following tables define their own default values for certain source arguments:

- **[gcp_audit_log](https://hub.tailpipe.io/plugins/turbot/gcp/tables/gcp_audit_log#gcp_storage_bucket)**
