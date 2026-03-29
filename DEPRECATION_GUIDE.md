# Deprecation Guide: gcp_logging_log_entry → gcp_logging_api

## Overview

The source identifier `gcp_logging_log_entry` has been **deprecated** and replaced with `gcp_logging_api` to align with naming conventions used in other Tailpipe plugins (e.g., `azure_activity_log_api`, `pipes_audit_log_api`).

## Migration Timeline

- **Current Version**: Both identifiers are supported
- **Future Version**: `gcp_logging_log_entry` will be removed (exact version TBD)

## What Changed

### Source Identifier
- **Old**: `gcp_logging_log_entry`
- **New**: `gcp_logging_api`

### Functionality
No functional changes - this is purely a naming update. All features, configuration options, and behavior remain identical.

## How to Migrate

### Step 1: Update Your Partition Configuration

**Before:**
```hcl
partition "gcp_audit_log" "my_logs" {
  source "gcp_logging_log_entry" {
    connection = connection.gcp.my_project
  }
}
```

**After:**
```hcl
partition "gcp_audit_log" "my_logs" {
  source "gcp_logging_api" {
    connection = connection.gcp.my_project
  }
}
```

### Step 2: Test Your Configuration

After updating, test your configuration:
```bash
tailpipe collect gcp_audit_log
```

## Deprecation Warnings

When you use the deprecated `gcp_logging_log_entry` identifier, you'll see a warning message:

```
WARN Source 'gcp_logging_log_entry' is deprecated and will be removed in a future version
     deprecated_source=gcp_logging_log_entry
     new_source=gcp_logging_api
     migration=Update your partition configuration to use 'gcp_logging_api' instead of 'gcp_logging_log_entry'
```

This warning appears in your logs but does not prevent the source from working.

## Implementation Details

### For Developers

The deprecation is implemented using a wrapper pattern:

1. **Main Source**: `LoggingLogEntrySource` with identifier `gcp_logging_api`
2. **Deprecated Source**: `DeprecatedLoggingLogEntrySource` with identifier `gcp_logging_log_entry`

The deprecated source:
- Embeds the main source (inherits all functionality)
- Overrides `Identifier()` to return the old identifier
- Logs a deprecation warning in `Init()`
- Delegates all operations to the embedded source

Both sources are registered in `gcp/plugin.go`:
```go
row_source.RegisterRowSource[*logging_log_entry.LoggingLogEntrySource]()
row_source.RegisterRowSource[*logging_log_entry.DeprecatedLoggingLogEntrySource]()
```

## Why This Change?

1. **Consistency**: Aligns with naming patterns in Azure (`azure_activity_log_api`) and Pipes (`pipes_audit_log_api`) plugins
2. **Clarity**: `gcp_logging_api` better reflects that this source uses the GCP Logging API service
3. **Future-proofing**: Allows for potential future sources like `gcp_logging_metrics_api` or `gcp_logging_sinks_api`


