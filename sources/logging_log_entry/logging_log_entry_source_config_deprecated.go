package logging_log_entry

// DeprecatedLoggingLogEntrySourceConfig is a deprecated config that uses the old identifier.
// It's functionally identical to LoggingLogEntrySourceConfig but uses the deprecated identifier.
//
// Deprecated: Use LoggingLogEntrySourceConfig with identifier "gcp_logging_api" instead.
type DeprecatedLoggingLogEntrySourceConfig struct {
	LoggingLogEntrySourceConfig
}

func (a *DeprecatedLoggingLogEntrySourceConfig) Identifier() string {
	return DeprecatedLoggingLogEntrySourceIdentifier
}

