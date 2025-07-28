---
title: "Source: gcp_cloud_logging_api - Collect logs from GCP Cloud Logging API"
description: "Allows users to collect logs from Google Cloud Platform (GCP) Cloud Logging API."
---

# Source: gcp_cloud_logging_api - Obtain logs from GCP Cloud Logging API

The Google Cloud Platform (GCP) Cloud Logging API provides access to all logs for all GCP services. It allows you to view and manage logs for your GCP projects, services, and applications.

This source is currently configured only for request logs from Google Load Balancer and Cloud Armor logs, from the source log name `projects/project-id/logs/requests`

Using this source, currently you can collect, filter, and analyze request logs that have been enriched with Cloud Armor rule findings, in order to collect metrics on blocking or analyze to eliminate false positive findings that would block wanted application requests.

Any other log type except for audit logs are of the `logEntry` type, and this source can potentially collect them with minor changes to the source code.  (Tables must still be created for each type of log)

## Example Configurations

### Collect request logs

Collect all of the request logs for a project.

```hcl
connection "gcp" "my_project" {
  project = "my-gcp-project"
}

partition "gcp_requests_log" "my_logs" {
  source "gcp_cloud_logging_api" {
    connection = connection.gcp.my_project
  }
}
```


## Arguments

| Argument   | Type             | Required | Default                  | Description                                                                                                                   |
|------------|------------------|----------|--------------------------|-------------------------------------------------------------------------------------------------------------------------------|
| connection | `connection.gcp` | No       | `connection.gcp.default` | The [GCP connection](https://hub.tailpipe.io/plugins/turbot/gcp#connection-credentials) to use to connect to the GCP account. |
| log_types  | List(String)     | No       | []                       | This could any type of non-audit log that confirms to the [logEntry data model](https://cloud.google.com/logging/docs/log-entry-data-model) and is stored in the `_Default` logging bucket in a GCP project.  The only restriction is what tables are supported in Tailpipe.  Currently the only supported type is `requests`
