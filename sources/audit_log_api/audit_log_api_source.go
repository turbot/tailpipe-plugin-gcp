package audit_log_api

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

const AuditLogAPISourceIdentifier = "gcp_audit_log_api"

// AuditLogAPISource source is responsible for collecting audit logs from GCP
type AuditLogAPISource struct {
	row_source.RowSourceImpl[*AuditLogAPISourceConfig, *config.GcpConnection]
}

func (s *AuditLogAPISource) Init(ctx context.Context, params *row_source.RowSourceParams, opts ...row_source.RowSourceOption) error {
	// set the collection state ctor
	s.NewCollectionStateFunc = collection_state.NewTimeRangeCollectionState

	// call base init
	return s.RowSourceImpl.Init(ctx, params, opts...)
}

func (s *AuditLogAPISource) Identifier() string {
	return AuditLogAPISourceIdentifier
}

func (s *AuditLogAPISource) Collect(ctx context.Context) error {
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

	sourceName := AuditLogAPISourceIdentifier
	sourceEnrichmentFields := &schema.SourceEnrichment{
		CommonFields: schema.CommonFields{
			TpSourceName:     &sourceName,
			TpSourceType:     AuditLogAPISourceIdentifier,
			TpSourceLocation: &project,
		},
	}

	// build the filter to fetch the logs for the given project, log types and time range
	filter := s.getLogFilter(project, logTypes, s.CollectionTimeRange)

	// TODO: #ratelimit implement rate limiting
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

		if logEntry != nil {
			if s.CollectionState.ShouldCollect(logEntry.InsertID, logEntry.Timestamp) {
				row := &types.RowData{
					Data:             *logEntry,
					SourceEnrichment: sourceEnrichmentFields,
				}

				if err = s.CollectionState.OnCollected(logEntry.InsertID, logEntry.Timestamp); err != nil {
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

func (s *AuditLogAPISource) getClient(ctx context.Context, project string) (*logadmin.Client, error) {
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

func (s *AuditLogAPISource) getLogFilter(projectId string, logTypes []string, timeRange collection_state.DirectionalTimeRange) string {
	activity := fmt.Sprintf(`"projects/%s/logs/cloudaudit.googleapis.com%sactivity"`, projectId, "%2F")
	dataAccess := fmt.Sprintf(`"projects/%s/logs/cloudaudit.googleapis.com%sdata_access"`, projectId, "%2F")
	systemEvent := fmt.Sprintf(`"projects/%s/logs/cloudaudit.googleapis.com%ssystem_event"`, projectId, "%2F")
	policyDenied := fmt.Sprintf(`"projects/%s/logs/cloudaudit.googleapis.com%spolicy"`, projectId, "%2F")
	cloudRunRequest := fmt.Sprintf(`"projects/%s/logs/run.googleapis.com%srequests"`, projectId, "%2F")
	appEngineRequest := fmt.Sprintf(`"projects/%s/logs/appengine.googleapis.com%request_log"`, projectId, "%2F")
	requests := fmt.Sprintf(`"projects/%s/logs/requests"`, projectId)
	// construct filter for time range
	timePart := fmt.Sprintf(`AND (timestamp >= "%s") AND (timestamp < "%s")`,
		timeRange.StartTime().Format(time.RFC3339Nano),
		timeRange.EndTime().Format(time.RFC3339Nano))

	// short-circuit default
	if len(logTypes) == 0 {
		return fmt.Sprintf("logName=(%s OR %s OR %s OR %s OR %s OR %s OR %s) %s", activity, dataAccess, systemEvent, policyDenied, cloudRunRequest, appEngineRequest, requests, timePart)
	}

	var selected []string
	for _, logType := range logTypes {
		switch logType {
		case "activity":
			selected = append(selected, activity)
		case "data_access":
			selected = append(selected, dataAccess)
		case "system_event":
			selected = append(selected, systemEvent)
		case "policy":
			selected = append(selected, policyDenied)
		case "cloud_run_request":
			selected = append(selected, cloudRunRequest)
		case "app_engine_request":
			selected = append(selected, appEngineRequest)
		case "requests":
			selected = append(selected, requests)
		}
	}

	switch len(selected) {
	case 0:
		return fmt.Sprintf("logName=(%s OR %s OR %s or %s OR %s OR %s OR %s) %s", activity, dataAccess, systemEvent, policyDenied, cloudRunRequest, appEngineRequest, requests, timePart)
	case 1:
		return fmt.Sprintf("logName=%s %s", selected[0], timePart)
	default:
		return fmt.Sprintf("logName=(%s) %s", strings.Join(selected, " OR "), timePart)
	}
}
