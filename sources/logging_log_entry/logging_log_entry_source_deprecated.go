package logging_log_entry

import (
	"context"
	"log/slog"

	"github.com/turbot/tailpipe-plugin-sdk/row_source"
)

// DeprecatedLoggingLogEntrySource is a deprecated wrapper around LoggingLogEntrySource
// that uses the old identifier "gcp_logging_log_entry" for backward compatibility.
// This source will be removed in a future version. Users should migrate to "gcp_logging_api".
//
// Deprecated: Use LoggingLogEntrySource with identifier "gcp_logging_api" instead.
type DeprecatedLoggingLogEntrySource struct {
	LoggingLogEntrySource
}

func (s *DeprecatedLoggingLogEntrySource) Identifier() string {
	return DeprecatedLoggingLogEntrySourceIdentifier
}

func (s *DeprecatedLoggingLogEntrySource) Init(ctx context.Context, params *row_source.RowSourceParams, opts ...row_source.RowSourceOption) error {
	// Log deprecation warning
	slog.Warn(
		"Source 'gcp_logging_log_entry' is deprecated and will be removed in a future version",
		"deprecated_source", DeprecatedLoggingLogEntrySourceIdentifier,
		"new_source", LoggingLogEntrySourceIdentifier,
		"migration", "Update your partition configuration to use 'gcp_logging_api' instead of 'gcp_logging_log_entry'",
	)

	// Delegate to the embedded source's Init
	return s.LoggingLogEntrySource.Init(ctx, params, opts...)
}

