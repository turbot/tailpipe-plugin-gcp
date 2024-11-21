package sources

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/logging/logadmin"
	"google.golang.org/api/iterator"

	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/collection_state"
	"github.com/turbot/tailpipe-plugin-sdk/config_data"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/types"
)

const AuditLogAPISourceIdentifier = "gcp_audit_log_api"

func init() {
	row_source.RegisterRowSource[*AuditLogAPISource]()
}

// AuditLogAPISource source is responsible for collecting audit logs from GCP
type AuditLogAPISource struct {
	row_source.RowSourceImpl[*AuditLogAPISourceConfig, *artifact_source.GcpConnection]
}

func (s *AuditLogAPISource) Init(ctx context.Context, configData, connectionData config_data.ConfigData, opts ...row_source.RowSourceOption) error {
	// set the collection state ctor
	s.NewCollectionStateFunc = collection_state.NewTimeRangeCollectionState

	// call base init
	return s.RowSourceImpl.Init(ctx, configData, connectionData, opts...)
}

func (s *AuditLogAPISource) Identifier() string {
	return AuditLogAPISourceIdentifier
}

func (s *AuditLogAPISource) Collect(ctx context.Context) error {
	collectionState := s.CollectionState.(*collection_state.TimeRangeCollectionState[*AuditLogAPISourceConfig])
	// TODO: #config the below should be settable via a config option
	collectionState.IsChronological = true
	collectionState.HasContinuation = true

	startTime := collectionState.GetLatestEndTime()
	project := s.Connection.GetProject()
	logTypes := s.Config.LogTypes

	client, err := s.getClient(ctx, project)
	if err != nil {
		return err
	}
	defer client.Close()

	if startTime == nil {
		// TODO: #config figure out how to determine an appropriate start time if collectionState doesn't provide one, defaulting to 30 days in meantime
		st := time.Now().Add(-720 * time.Hour)
		startTime = &st
	}

	sourceName := AuditLogAPISourceIdentifier
	sourceEnrichmentFields := &enrichment.CommonFields{
		TpSourceName:     &sourceName,
		TpSourceType:     AuditLogAPISourceIdentifier,
		TpSourceLocation: &project,
	}

	filter := s.getLogNameFilter(project, logTypes, *startTime)
	collectionState.StartCollection()
	// TODO: #ratelimit implement rate limiting
	it := client.Entries(ctx, logadmin.Filter(filter))
	for {
		logEntry, err := it.Next()
		if err != nil && errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return fmt.Errorf("error fetching log entries, %w", err)
		}

		if logEntry != nil {
			if collectionState.ShouldCollectRow(logEntry.Timestamp, logEntry.InsertID) {
				row := &types.RowData{
					Data:     *logEntry,
					Metadata: sourceEnrichmentFields,
				}

				// update collection state
				collectionState.Upsert(logEntry.Timestamp, logEntry.InsertID, nil)
				collectionStateJSON, err := s.GetCollectionStateJSON()
				if err != nil {
					return fmt.Errorf("error serialising collectionState data: %w", err)
				}

				if err := s.OnRow(ctx, row, collectionStateJSON); err != nil {
					return fmt.Errorf("error processing row: %w", err)
				}
			}
		}
	}

	collectionState.EndCollection()

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

func (s *AuditLogAPISource) getLogNameFilter(projectId string, logTypes []string, startTime time.Time) string {
	activity := fmt.Sprintf(`"projects/%s/logs/cloudaudit.googleapis.com%sactivity"`, projectId, "%2F")
	dataAccess := fmt.Sprintf(`"projects/%s/logs/cloudaudit.googleapis.com%sdata_access"`, projectId, "%2F")
	systemEvent := fmt.Sprintf(`"projects/%s/logs/cloudaudit.googleapis.com%ssystem_event"`, projectId, "%2F")
	timePart := fmt.Sprintf(`AND (timestamp > "%s")`, startTime.Format(time.RFC3339Nano))

	// short-circuit default
	if len(logTypes) == 0 {
		return fmt.Sprintf("logName=(%s OR %s OR %s) %s", activity, dataAccess, systemEvent, timePart)
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
		}
	}

	switch len(selected) {
	case 0: // TODO: #error do we throw an error instead of returning default options here?
		return fmt.Sprintf("logName=(%s OR %s OR %s) %s", activity, dataAccess, systemEvent, timePart)
	case 1:
		return fmt.Sprintf("logName=%s %s", selected[0], timePart)
	default:
		return fmt.Sprintf("logName=(%s) %s", strings.Join(selected, " OR "), timePart)
	}
}
