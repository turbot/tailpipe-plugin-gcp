# GCP Plugin for Tailpipe

[Tailpipe](https://tailpipe.io) is an open-source CLI tool that allows you to collect logs and query them with SQL.

[GCP](https://cloud.google.com) provides on-demand cloud computing platforms and APIs to authenticated customers on a metered pay-as-you-go basis.

The [GCP Plugin for Tailpipe](https://hub.tailpipe.io/plugins/turbot/gcp) allows you to collect and query GCP logs using SQL to track activity, monitor trends, detect anomalies, and more!

- **[Get started →](https://hub.tailpipe.io/plugins/turbot/gcp)**
- Documentation: [Table definitions & examples](https://hub.tailpipe.io/plugins/turbot/gcp/tables)
- Community: [Join #tailpipe on Slack →](https://turbot.com/community/join)
- Get involved: [Issues](https://github.com/turbot/tailpipe-plugin-gcp/issues)

Collect and query logs:
![image](docs/images/gcp_audit_log_terminal.png)

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

For example, the [GCP Audit Log Detections mod](https://hub.powerpipe.io/mods/turbot/tailpipe-mod-gcp-caudit-log-detections) scans your CloudTrail logs for anomalies, such as an Storage bucket being made public or a change in your Compute network infrastructure.

Dashboards and detections are [open source](https://github.com/topics/tailpipe-mod), allowing easy customization and collaboration.

To get started, choose a mod from the [Powerpipe Hub](https://hub.powerpipe.io/?engines=tailpipe&q=gcp).

![image](docs/images/gcp_audit_log_mitre_dashboard.png)

## Developing

Prerequisites:

- [Tailpipe](https://tailpipe.io/downloads)
- [Golang](https://golang.org/doc/install)

Clone:

```sh
git clone https://github.com/turbot/tailpipe-plugin-gcp.git
cd tailpipe-plugin-gcp
```

After making your local changes, build the plugin, which automatically installs the new version to your `~/.tailpipe/plugins` directory:

```
make
```

Re-collect your data:

```
tailpipe collect gcp_audit_log
```

Try it!

```
tailpipe query
> .inspect gcp_audit_log
```

## Open Source & Contributing

This repository is published under the [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0) (source code) and [CC BY-NC-ND](https://creativecommons.org/licenses/by-nc-nd/2.0/) (docs) licenses. Please see our [code of conduct](https://github.com/turbot/.github/blob/main/CODE_OF_CONDUCT.md). We look forward to collaborating with you!

[Tailpipe](https://tailpipe.io) is a product produced from this open source software, exclusively by [Turbot HQ, Inc](https://turbot.com). It is distributed under our commercial terms. Others are allowed to make their own distribution of the software, but cannot use any of the Turbot trademarks, cloud services, etc. You can learn more in our [Open Source FAQ](https://turbot.com/open-source).

## Get Involved

**[Join #tailpipe on Slack →](https://turbot.com/community/join)**

Want to help but don't know where to start? Pick up one of the `help wanted` issues:

- [Tailpipe](https://github.com/turbot/tailpipe/labels/help%20wanted)
- [GCP Plugin](https://github.com/turbot/tailpipe-plugin-gcp/labels/help%20wanted)
