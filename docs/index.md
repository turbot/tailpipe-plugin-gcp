---
organization: Turbot
category: ["public cloud"]
icon_url: "/images/plugins/turbot/gcp.svg"
brand_color: "#ea4335"
display_name: "GCP"
description: "Tailpipe plugin for collecting and querying various logs from GCP."
og_description: "Collect GCP logs and query them instantly with SQL! Open source CLI. No DB required."
og_image: "/images/plugins/turbot/gcp-social-graphic.png"
---

# GCP + Tailpipe

[Tailpipe](https://tailpipe.io) is an open-source CLI tool that allows you to collect logs and query them with SQL.

[GCP](https://cloud.google.com) provides on-demand cloud computing platforms and APIs to authenticated customers on a metered pay-as-you-go basis.

The [GCP Plugin for Tailpipe](https://hub.tailpipe.io/plugins/turbot/gcp) allows you to collect and query GCP logs using SQL to track activity, monitor trends, detect anomalies, and more!

- Documentation: [Table definitions & examples](https://hub.tailpipe.io/plugins/turbot/gcp/tables)
- Community: [Join #tailpipe on Slack →](https://turbot.com/community/join)
- Get involved: [Issues](https://github.com/turbot/tailpipe-plugin-gcp/issues)

![Image](https://raw.githubusercontent.com/turbot/tailpipe-plugin-gcp/main/docs/images/gcp_audit_log_terminal.png?type=thumbnail)  
![Image](https://raw.githubusercontent.com/turbot/tailpipe-plugin-gcp/main/docs/images/gcp_audit_log_mitre_dashboard.png?type=thumbnail)

## Getting Started

Install Tailpipe from the [downloads](https://tailpipe.io/downloads) page:

```sh
# MacOS
brew install turbot/tap/tailpipe
```

```sh
# Linux or Windows (WSL)
sudo /bin/sh -c "$(curl -fsSL https://tailpipe.io/install/tailpipe.sh)"
```

Install the plugin:

```sh
tailpipe plugin install gcp
```

Configure your [connection credentials](https://hub.tailpipe.io/plugins/turbot/gcp#connection-credentials), table partition, and data source ([examples](https://hub.tailpipe.io/plugins/turbot/gcp/tables/gcp_audit_log#example-configurations)):

```sh
vi ~/.tailpipe/config/gcp.tpc
```

```hcl
connection "gcp" "my_project" {
  project = "my-project"
}

partition "gcp_audit_log" "my_logs" {
  source "gcp_storage_bucket" {
    connection = connection.gcp.my_project
    bucket     = "gcp-audit-logs-bucket"
  }
}
```

Download, enrich, and save logs from your source ([examples](https://tailpipe.io/docs/reference/cli/collect)):

```sh
tailpipe collect gcp_audit_log
```

Enter interactive query mode:

```sh
tailpipe query
```

Run a query:

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
  event_count desc;
```

```sh
+-------------------------+-------------------------------------------------------+-------------+
| event_source            | event_name                                            | event_count |
+-------------------------+-------------------------------------------------------+-------------+
| storage.googleapis.com  | storage.objects.get                                   | 104349      |
| logging.googleapis.com  | google.logging.v2.LoggingServiceV2.ListLogEntries     | 28193       |
| compute.googleapis.com  | v1.compute.instanceGroupManagers.listManagedInstances | 27236       |
| storage.googleapis.com  | storage.objects.create                                | 11817       |
| cloudkms.googleapis.com | Decrypt                                               | 4171        |
+-------------------------+-------------------------------------------------------+-------------+
```

## Detections as Code with Powerpipe

Pre-built dashboards and detections for the GCP plugin are available in [Powerpipe](https://powerpipe.io) mods, helping you monitor and analyze activity across your GCP accounts.

For example, the [GCP CloudTrail Logs Detections mod](https://hub.powerpipe.io/mods/turbot/tailpipe-mod-gcp-cloudtrail-log-detections) scans your CloudTrail logs for anomalies, such as an Storage bucket being made public or a change in your Compute network infrastructure.

Dashboards and detections are [open source](https://github.com/topics/tailpipe-mod), allowing easy customization and collaboration.

To get started, choose a mod from the [Powerpipe Hub](https://hub.powerpipe.io/?engines=tailpipe&q=gcp).

## Connection Credentials

### Arguments

| Name                          | Type   | Required | Description                                                                                     |
| ----------------------------- | ------ | -------- | ----------------------------------------------------------------------------------------------- |
| `credentials`                 | String | No       | Path to the JSON credentials file or the contents of a service account key file in JSON format. |
| `impersonate_access_token`    | String | No       | An OAuth 2.0 access token used to impersonate a service account.                                |
| `impersonate_service_account` | String | No       | The email of the service account to impersonate for authentication.                             |
| `project`                     | String | No       | The project ID to connect to.                                                                   |
| `quota_project`               | String | No       | The project ID to use for quota usage and billing purposes.                                     |
| `min_retry_delay`             | Int    | No       | Initial retry delay in milliseconds for API rate limiting (default: 500ms).                     |
| `max_retry_delay`             | Int    | No       | Maximum retry delay in milliseconds for API rate limiting (default: 60000ms).                   |
| `backoff_multiplier`          | Float  | No       | Exponential growth multiplier for retry delays (default: 1.30).                                 |

### Application Default Credentials

By default, the GCP plugin uses your [Application Default Credentials](https://cloud.google.com/sdk/gcloud/reference/auth/application-default) to connect to GCP. If you have not set up ADC, simply run `gcloud auth application-default login`. This command will prompt you to log in, and then will download the application default credentials to `~/.config/gcloud/application_default_credentials.json`.

### Service Account Credentials

Generate and download a JSON key for an existing service account using: [create service account key page](https://console.cloud.google.com/apis/credentials/serviceaccountkey).

```hcl
connection "gcp" "gcp_my_other_project" {
  project     = "my-other-project"
  credentials = "/home/me/my-service-account-creds.json"
}
```

### Impersonation Access Token Credentials

Generate an impersonate access token using: [gcloud CLI command](https://cloud.google.com/iam/docs/create-short-lived-credentials-direct#gcloud_2).

```hcl
connection "gcp" "gcp_my_other_project" {
  project                  = "my-other-project"
  impersonate_access_token = "ya29.c.c0ASRK0GZ7mv8lIV0iiudmiGBs9m1gqGfBYZzV...aMYJd"
}
```

### Credentials from Environment Variables

The GCP plugin will use the standard GCP environment variables to obtain credentials **only if other arguments (`credentials`, `impersonate_access_token`, etc..) are not specified** in the connection:

```sh
export CLOUDSDK_CORE_PROJECT=myproject
export GOOGLE_CLOUD_QUOTA_PROJECT=billingproject
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/my/creds.json
```
