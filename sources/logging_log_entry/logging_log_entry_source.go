package logging_log_entry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	loggingCore "cloud.google.com/go/logging"
	logging "cloud.google.com/go/logging/apiv2"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"

	"github.com/turbot/tailpipe-plugin-gcp/config"
	"github.com/turbot/tailpipe-plugin-sdk/collection_state"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/schema"
	"github.com/turbot/tailpipe-plugin-sdk/types"
)

const (
	// LoggingLogEntrySourceIdentifier is the current (new) identifier for this source
	LoggingLogEntrySourceIdentifier = "gcp_logging_api"
	// DeprecatedLoggingLogEntrySourceIdentifier is the deprecated identifier (kept for backward compatibility)
	DeprecatedLoggingLogEntrySourceIdentifier = "gcp_logging_log_entry"
	DefaultLogEntriesPageSize                 = 100000
)

// WithTableName is a custom option to pass the table name to the source
func WithTableName(tableName string) row_source.RowSourceOption {
	return func(source row_source.RowSource) error {
		if s, ok := source.(*LoggingLogEntrySource); ok {
			s.tableName = tableName
		}
		return nil
	}
}

// LoggingLogEntrySource source is responsible for collecting logs from GCP using Audit log log entries API.
type LoggingLogEntrySource struct {
	row_source.RowSourceImpl[*LoggingLogEntrySourceConfig, *config.GcpConnection]
	tableName     string
	logTypeFilter *LogTypeFilter
}

func (s *LoggingLogEntrySource) Init(ctx context.Context, params *row_source.RowSourceParams, opts ...row_source.RowSourceOption) error {
	// set the collection state ctor
	s.NewCollectionStateFunc = collection_state.NewTimeRangeCollectionState

	// Initialize the log type filter
	s.logTypeFilter = NewLogTypeFilter()

	// call base init first to apply options
	if err := s.RowSourceImpl.Init(ctx, params, opts...); err != nil {
		return err
	}

	// Validate log types for the specific table if table name is available
	if s.tableName != "" && s.Config != nil && len(s.Config.LogTypes) > 0 {
		if err := s.Config.ValidateForTable(s.tableName); err != nil {
			return fmt.Errorf("validation failed for table %s: %w", s.tableName, err)
		}
	}

	return nil
}

func (s *LoggingLogEntrySource) Identifier() string {
	return LoggingLogEntrySourceIdentifier
}

// GetTableName returns the table name extracted from the collection state path
func (s *LoggingLogEntrySource) GetTableName() string {
	return s.tableName
}

func (s *LoggingLogEntrySource) Collect(ctx context.Context) error {
	project := s.Connection.GetProject()
	var logTypes []string
	if s.Config != nil && s.Config.LogTypes != nil {
		logTypes = s.Config.LogTypes
	}

	client, err := s.getClient(ctx, project)
	if err != nil {
		return err
	}
	defer client.Close()

	sourceName := LoggingLogEntrySourceIdentifier
	sourceEnrichmentFields := &schema.SourceEnrichment{
		CommonFields: schema.CommonFields{
			TpSourceName:     &sourceName,
			TpSourceType:     LoggingLogEntrySourceIdentifier,
			TpSourceLocation: &project,
		},
	}

	// build the filter to fetch the logs for the given project, table name, log types and time range
	filter := s.getLogFilter(project, s.tableName, logTypes, s.CollectionTimeRange)

	req := &loggingpb.ListLogEntriesRequest{
		ResourceNames: []string{fmt.Sprintf("projects/%s", project)},
		Filter:        filter,
		PageSize:      DefaultLogEntriesPageSize,
	}

	it := client.ListLogEntries(ctx, req)
	for {
		logEntry, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// If we get here, retries were already applied & exhausted.
			return fmt.Errorf("ListLogEntries failed after retries: %w", err)
		}

		if logEntry != nil {
			if s.CollectionState.ShouldCollect(logEntry.GetInsertId(), logEntry.GetTimestamp().AsTime()) {
				row := &types.RowData{
					Data:             logEntry,
					SourceEnrichment: sourceEnrichmentFields,
				}

				if err = s.CollectionState.OnCollected(logEntry.GetInsertId(), logEntry.GetTimestamp().AsTime()); err != nil {
					return fmt.Errorf("error updating collection state: %w", err)
				}
				if err = s.OnRow(ctx, row); err != nil {
					return fmt.Errorf("error processing row: %w", err)
				}
			}
		}
	}

	return nil
}

func (s *LoggingLogEntrySource) getClient(ctx context.Context, project string) (*logging.Client, error) {
	if project == "" {
		return nil, errors.New("unable to determine active project, please set project in configuration or env var CLOUDSDK_CORE_PROJECT / GCP_PROJECT")
	}

	// Get the connection options that were working with logadmin
	opts, err := s.Connection.GetClientOptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting client options: %v", err)
	}

	// Add the logging scope to ensure we have the right permissions
	opts = append(opts, option.WithScopes(loggingCore.AdminScope))

	// Create GAPIC client so we can override per-method CallOptions.
	client, err := logging.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("logging/apiv2.NewClient: %v", err)
	}

	// Get retry configuration to calculate timeout
	backOff := s.getConfigurableBackoff()
	// Set timeout to maximum delay + 1
	timeout := backOff.Max + time.Millisecond

	// Configure retry mechanism for ListLogEntries with configurable parameters
	// Retry mechanism has been implemented as suggested by GCP: https://cloud.google.com/storage/docs/retry-strategy#customize-retries
	client.CallOptions.ListLogEntries = []gax.CallOption{
		gax.WithTimeout(timeout),
		gax.WithRetry(func() gax.Retryer {
			return gax.OnCodes(
				[]codes.Code{
					codes.Unavailable,
					codes.DeadlineExceeded,
					codes.ResourceExhausted,
					codes.Aborted,
				},
				backOff,
			)
		}),
	}

	return client, nil
}

// getConfigurableBackoff returns a gax.Backoff configured with connection-level parameters
// or defaults if not specified
func (s *LoggingLogEntrySource) getConfigurableBackoff() gax.Backoff {
	// Get retry configuration from connection
	initial, max, multiplier := s.Connection.GetRetryConfig()

	return gax.Backoff{
		Initial:    initial,
		Max:        max,
		Multiplier: multiplier,
	}
}

func (s *LoggingLogEntrySource) getLogFilter(projectId string, tableName string, logTypes []string, timeRange collection_state.DirectionalTimeRange) string {
	// construct filter for time range
	timePart := fmt.Sprintf(`AND (timestamp >= "%s") AND (timestamp < "%s")`,
		timeRange.StartTime().Format(time.RFC3339Nano),
		timeRange.EndTime().Format(time.RFC3339Nano))

	// If specific log types are provided, use them
	if len(logTypes) > 0 {
		var selected []string
		for _, logType := range logTypes {
			if s.logTypeFilter.IsValidLogTypeForTable(tableName, logType) {
				filters := s.logTypeFilter.GetLogFiltersForTableAndType(tableName, logType, projectId)
				selected = append(selected, filters...)
			}
		}

		switch len(selected) {
		case 0:
			// If no valid log types found, use all available for the table
			filters := s.logTypeFilter.GetLogFiltersForTable(tableName, projectId)
			if len(filters) > 0 {
				return fmt.Sprintf("logName=(%s) %s", strings.Join(filters, " OR "), timePart)
			}
		case 1:
			return fmt.Sprintf("logName=%s %s", selected[0], timePart)
		default:
			return fmt.Sprintf("logName=(%s) %s", strings.Join(selected, " OR "), timePart)
		}
	}

	// Default: include all log types for the table
	filters := s.logTypeFilter.GetLogFiltersForTable(tableName, projectId)
	if len(filters) > 0 {
		return fmt.Sprintf("logName=(%s) %s", strings.Join(filters, " OR "), timePart)
	}

	// Fallback: include all log types from all tables if table not found
	allFilters := []string{}
	for tableName := range s.logTypeFilter.TableLogTypeMap {
		filters := s.logTypeFilter.GetLogFiltersForTable(tableName, projectId)
		allFilters = append(allFilters, filters...)
	}
	return fmt.Sprintf("logName=(%s) %s", strings.Join(allFilters, " OR "), timePart)
}
