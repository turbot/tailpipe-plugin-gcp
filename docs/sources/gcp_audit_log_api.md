---
title: "Source: gcp_audit_log_api - Collect logs from GCP audit log API"
description: "Allows users to collect logs from Google Cloud Platform (GCP) audit log API."
---

# Source: gcp_audit_log_api - Obtain logs from GCP audit log API

The Google Cloud Platform (GCP) audit log API provides access to audit logs for GCP services. It allows you to view and manage logs for your GCP projects, including logs for administrative actions, data access, and system events.

Using this source, you can collect, filter, and analyze logs retrieved from the GCP audit log API, enabling system monitoring, security investigations, and compliance reporting.

## Example Configurations

### Collect audit logs

Collect all types of audit logs for a project.

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

### Collect specific types of audit logs

Collect admin activity and data access audit logs for a project.

```hcl
partition "gcp_audit_log" "my_logs_admin_data_access" {
  source "gcp_audit_log_api" {
    connection = connection.gcp.my_project
    log_types = ["activity", "data_access"]
  }
}
```

## Arguments

| Argument   | Required | Default                  | Description                                                                                                                   |
|------------|----------|--------------------------|-------------------------------------------------------------------------------------------------------------------------------|
| connection | No       | `connection.gcp.default` | The [GCP connection](https://hub.tailpipe.io/plugins/turbot/gcp#connection-credentials) to use to connect to the GCP account. |
| log_types  | No       | []                       | A list of [audit log types](https://cloud.google.com/logging/docs/audit#types) to retrieve. If no types are specified, all log types are retrieved. Valid values: `activity`, `data_access`, `system_event`.                                                                       |
