## [TBD]

_Breaking Changes_

- **Source Renamed**: The source `gcp_logging_log_entry` has been renamed to `gcp_logging_api` to align with naming conventions used in other Tailpipe plugins (e.g., `azure_activity_log_api`, `pipes_audit_log_api`). The old identifier is deprecated but will continue to work with a deprecation warning. See the [deprecation guide](DEPRECATION_GUIDE.md) for migration instructions. ([#93](https://github.com/turbot/tailpipe-plugin-gcp/pull/93))

_Enhancements_

- Migrated GCP logging source from legacy `logadmin` client to modern GAPIC client (`apiv2`) for improved performance and features. ([#93](https://github.com/turbot/tailpipe-plugin-gcp/pull/93))
- Added configurable retry mechanism with exponential backoff for GCP logging API calls. Retry parameters can be configured at the connection level: `min_retry_delay`, `max_retry_delay`, and `backoff_multiplier`. ([#93](https://github.com/turbot/tailpipe-plugin-gcp/pull/93))

## v0.6.0 [2025-08-14]

_What's new?_

- New tables added: ([#80](https://github.com/turbot/tailpipe-plugin-gcp/pull/80))
  - [gcp_billing_report](https://hub.tailpipe.io/plugins/turbot/gcp/tables/gcp_billing_report)

## v0.5.2 [2025-07-28]

_Dependencies_

- Recompiled plugin with [tailpipe-plugin-sdk v0.9.2](https://github.com/turbot/tailpipe-plugin-sdk/blob/develop/CHANGELOG.md#v092-2025-07-24) that fixes incorrect data ranges for zero‑granularity collections and prevents crashes in certain collection states. ([#81](https://github.com/turbot/tailpipe-plugin-gcp/pull/81))

## v0.5.1 [2025-07-02]

_Dependencies_

- Recompiled plugin with [tailpipe-plugin-sdk v0.9.1](https://github.com/turbot/tailpipe-plugin-sdk/blob/develop/CHANGELOG.md#v091-2025-07-02) that fixes collection state issues. ([#72](https://github.com/turbot/tailpipe-plugin-gcp/pull/72))

## v0.5.0 [2025-07-02]

_Dependencies_

- Recompiled plugin with [tailpipe-plugin-sdk v0.9.0](https://github.com/turbot/tailpipe-plugin-sdk/blob/develop/CHANGELOG.md#v090-2025-07-02) to support the `--to` flag and update tracking of collected data to support multiple separate time ranges. ([#66](https://github.com/turbot/tailpipe-plugin-gcp/pull/66))

## v0.4.2 [2025-06-05]

- Recompiled plugin with [tailpipe-plugin-sdk v0.7.2](https://github.com/turbot/tailpipe-plugin-sdk/blob/develop/CHANGELOG.md#v072-2025-06-04) that fixes an issue where the end time was not correctly recorded for collections using artifact sources. ([#57](https://github.com/turbot/tailpipe-plugin-gcp/pull/57))

## v0.4.1 [2025-06-04]

- Recompiled plugin with [tailpipe-plugin-sdk v0.7.1](https://github.com/turbot/tailpipe-plugin-sdk/blob/develop/CHANGELOG.md#v071-2025-06-04) that fixes an issue affecting collections using a file source. ([#55](https://github.com/turbot/tailpipe-plugin-gcp/pull/55))

## v0.4.0 [2025-06-03]

_Dependencies_

- Recompiled plugin with [tailpipe-plugin-sdk v0.7.0](https://github.com/turbot/tailpipe-plugin-sdk/blob/develop/CHANGELOG.md#v070-2025-06-03) that improves how collection end times are tracked, helping make future collections more accurate and reliable. ([#53](https://github.com/turbot/tailpipe-plugin-gcp/pull/53))

## v0.3.0 [2025-03-03]

_Enhancements_

- Standardized all example query titles to use `Title Case` for consistency. ([#43](https://github.com/turbot/tailpipe-plugin-gcp/pull/43))
- Added `folder` front matter to all queries for improved organization and discoverability in the Hub. ([#43](https://github.com/turbot/tailpipe-plugin-gcp/pull/43))

## v0.2.0 [2025-02-06]

_Enhancements_

- Updated documentation formatting and enhanced argument descriptions for `gcp_audit_log_api` and `gcp_storage_bucket` sources. ([#42](https://github.com/turbot/tailpipe-plugin-gcp/pull/42))

## v0.1.0 [2025-01-30]

_What's new?_

- New tables added
  - [gcp_audit_log](https://hub.tailpipe.io/plugins/turbot/gcp/tables/gcp_activity_log)
- New sources added
  - [gcp_audit_log_api](https://hub.tailpipe.io/plugins/turbot/gcp/sources/gcp_audit_log_api)
  - [gcp_storage_bucket](https://hub.tailpipe.io/plugins/turbot/gcp/sources/gcp_storage_bucket)
