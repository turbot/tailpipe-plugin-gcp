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
