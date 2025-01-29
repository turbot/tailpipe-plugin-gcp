---
title: "Source: gcp_audit_log_api - Collect logs from GCP Audit Log API"
description: "Allows users to collect logs from Google Cloud Platform (GCP) Audit Log API."
---

# Source: gcp_audit_log_api - Obtain logs from GCP Audit Log API

The Google Cloud Platform (GCP) Audit Log API provides access to audit logs for GCP services. It allows you to view and manage logs for your GCP projects, including logs for administrative actions, data access, and system events.

Using this source, you can collect, filter, and analyze logs retrieved from the GCP Audit Log API, enabling system monitoring, security investigations, and compliance reporting.

## Example Configurations

### Collect Audit Logs for all projects

```hcl
connection "gcp" "logging_account" {
  project = "my-gcp-project"
}

partition "gcp_audit_log" "my_logs" {
  source "gcp_audit_log_api" {
    connection = connection.gcp.logging_account
  }
}
```

### Collect Audit Logs for a specific project

```hcl
partition "gcp_audit_log" "my_project_logs" {
  source "gcp_audit_log_api" {
    connection = connection.gcp.logging_account
    project    = "my-gcp-project"
  }
}
```

## Arguments

| Argument      | Required | Default                  | Description                                                                                                                |
|---------------|----------|--------------------------|----------------------------------------------------------------------------------------------------------------------------|
| connection    | No       | `connection.gcp.default` | The [GCP connection](https://tailpipe.io/docs/reference/config-files/connection/gcp) to use to connect to the GCP account. |
| project       | No       |                          | The GCP project ID from which logs should be retrieved.                                                                    |

### Table Defaults

The following tables define their own default values for certain source arguments:

- **[gcp_audit_log](https://tailpipe.io/plugins/turbot/gcp/tables/gcp_audit_log#gcp_audit_log_api)**

