package cloud_logging_api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/logging/logadmin"
	"google.golang.org/api/iterator"

	"github.com/turbot/tailpipe-plugin-gcp/config"
	"github.com/turbot/tailpipe-plugin-sdk/collection_state"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/schema"
	"github.com/turbot/tailpipe-plugin-sdk/types"
)

const CloudLoggingAPISourceIdentifier = "gcp_cloud_logging_api"

// CloudLoggingAPISource source is responsible for collecting cloud logs from GCP
type CloudLoggingAPISource struct {
	row_source.RowSourceImpl[*CloudLoggingAPISourceConfig, *config.GcpConnection]
}

func (s *CloudLoggingAPISource) Init(ctx context.Context, params *row_source.RowSourceParams, opts ...row_source.RowSourceOption) error {
	// set the collection state ctor
	s.NewCollectionStateFunc = collection_state.NewTimeRangeCollectionState

	// call base init
	return s.RowSourceImpl.Init(ctx, params, opts...)
}

func (s *CloudLoggingAPISource) Identifier() string {
	return CloudLoggingAPISourceIdentifier
}

func (s *CloudLoggingAPISource) Collect(ctx context.Context) error {
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

	sourceName := CloudLoggingAPISourceIdentifier
	sourceEnrichmentFields := &schema.SourceEnrichment{
		CommonFields: schema.CommonFields{
			TpSourceName:     &sourceName,
			TpSourceType:     CloudLoggingAPISourceIdentifier,
			TpSourceLocation: &project,
		},
	}

	filter := s.getLogNameFilter(project, logTypes, s.CollectionTimeRange)

	// TODO: #ratelimit Implement rate limiting to ensure compliance with GCP API quotas.
	//       Use a token bucket algorithm with a maximum of 100 requests per second and a burst capacity of 200.
	//       Refer to the GCP API rate-limiting documentation: https://cloud.google.com/apis/docs/rate-limits
	//       This feature should be implemented by Q4 2023 to prevent potential throttling issues.

	// logEntry will now be the higher-level logging.Entry
	var logEntry *logging.Entry
	it := client.Entries(ctx, logadmin.Filter(filter), logadmin.PageSize(250))
	for {
		logEntry, err = it.Next()
		if err != nil && errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return fmt.Errorf("error fetching log entries, %w", err)
		}

		if logEntry.Payload != nil {
			logEntry.LogName = "" // remove logName from the log entry due to ToLogEntry requirements of an empty string
			protoLogEntry, err := logging.ToLogEntry(*logEntry, project)
			if err != nil {
				return fmt.Errorf("error converting log entry to loggingpb.LogEntry: %w", err)
			}

			if s.CollectionState.ShouldCollect(protoLogEntry.GetInsertId(), protoLogEntry.GetTimestamp().AsTime()) {
				row := &types.RowData{
					Data:             protoLogEntry,
					SourceEnrichment: sourceEnrichmentFields,
				}

				if err = s.CollectionState.OnCollected(protoLogEntry.GetInsertId(), protoLogEntry.GetTimestamp().AsTime()); err != nil {
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

func (s *CloudLoggingAPISource) getClient(ctx context.Context, project string) (*logadmin.Client, error) {
	opts, err := s.Connection.GetClientOptions(ctx)
	if err != nil {
		return nil, err
	}

	if project == "" {
		return nil, errors.New("unable to determine active project, please set project in configuration or env var CLOUDSDK_CORE_PROJECT / GCP_PROJECT")
	}

	client, err := logadmin.NewClient(ctx, project, opts...)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (s *CloudLoggingAPISource) getLogNameFilter(projectId string, logTypes []string, timeRange collection_state.DirectionalTimeRange) string {
	requestsLog := fmt.Sprintf(`"projects/%s/logs/requests"`, projectId)
	timePart := fmt.Sprintf(`AND (timestamp >= "%s") AND (timestamp < "%s")`,
		timeRange.StartTime().Format(time.RFC3339Nano),
		timeRange.EndTime().Format(time.RFC3339Nano))

	// short-circuit default
	if len(logTypes) == 0 {
		return fmt.Sprintf("logName=%s %s", requestsLog, timePart)
	}

	// Only request logs supported at implementation.  Append additional cases for other log types as needed
	var selected []string
	for _, logType := range logTypes {
		switch logType {
		case "requests":
			selected = append(selected, requestsLog)
		}
	}

	switch len(selected) {
	case 0:
		return fmt.Sprintf("logName=%s %s", requestsLog, timePart)
	case 1:
		return fmt.Sprintf("logName=%s %s", selected[0], timePart)
	default:
		return fmt.Sprintf("logName=(%s) %s", strings.Join(selected, " OR "), timePart)
	}
}
