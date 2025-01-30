---
title: "Source: gcp_storage_bucket - Collect logs from GCP Storage Buckets"
description: "Allows users to collect logs from GCP Storage Buckets."
---

# Source: gcp_storage_bucket - Collect logs from GCP Storage Buckets

A GCP Storage Bucket is a cloud storage resource used to store objects like data files and metadata. It serves as a central repository for logs from GCP services such as Audit Logs, VPC Flow Logs, Cloud Functions, and more.

Using this source, you can collect, filter, and analyze logs stored in GCP Storage Buckets, enabling system monitoring, security investigations, and compliance reporting.

Most GCP tables define a default `file_path` for the `gcp_storage_bucket` source, so if your GCP logs are stored in default log locations, you don't need to override the `file_path` argument.

## Example Configurations

### Collect Audit Logs

Collect Audit Logs for all projects and regions.

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

### Collect Audit Logs with a prefix

Collect Audit Logs stored with a GCS key prefix.

```hcl
partition "gcp_audit_log" "my_logs_prefix" {
  source "gcp_storage_bucket" {
    connection = connection.gcp.logging_account
    bucket     = "gcp-audit-logs-bucket"
    prefix     = "my/prefix/"
  }
}
```

### Collect Audit Logs with a custom path

```hcl
partition "gcp_audit_log" "my_logs_custom_path" {
  source "gcp_storage_bucket" {
    connection  = connection.gcp.logging_account
    bucket      = "gcp-audit-logs-bucket"
    file_layout = "cloudaudit.googleapis.com/%{DATA:type}/%{YEAR:year}/%{MONTHNUM:month}/%{MONTHDAY:day}/%{HOUR:hour}:%{MINUTE:minute}:%{SECOND:second}_%{DATA:end_time}_%{DATA:suffix}.json"
  }
}
```

## Arguments

| Argument      | Required | Default                  | Description                                                                                                                |
|---------------|----------|--------------------------|----------------------------------------------------------------------------------------------------------------------------|
| bucket        | Yes      |                          | The name of the GCP Storage Bucket to collect logs from.                                                                   |
| connection    | No       | `connection.gcp.default` | The [GCP connection](https://tailpipe.io/docs/reference/config-files/connection/gcp) to use to connect to the GCP account. |
| file_layout   | No       |                          | The Grok pattern that defines the log file structure.                                                                      |
| prefix        | No       |                          | The GCS key prefix that comes after the name of the bucket you have designated for log file delivery.                      |

### Table Defaults

The following tables define their own default values for certain source arguments:

- **[gcp_audit_log](https://tailpipe.io/plugins/turbot/gcp/tables/gcp_audit_log#gcp_storage_bucket)**

