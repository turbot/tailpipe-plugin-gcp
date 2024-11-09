package sources

import (
	"context"
	"errors"
	"fmt"
	"github.com/turbot/tailpipe-plugin-sdk/config_data"
	"time"

	"cloud.google.com/go/logging/logadmin"
	"github.com/turbot/tailpipe-plugin-sdk/collection_state"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
	"github.com/turbot/tailpipe-plugin-sdk/parse"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/types"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const AuditLogAPISourceIdentifier = "gcp_audit_log_api"

// AuditLogAPISource source is responsible for collecting audit logs from GCP
type AuditLogAPISource struct {
	row_source.RowSourceImpl[*AuditLogAPISourceConfig]
}

func NewAuditLogAPISource() row_source.RowSource {
	return &AuditLogAPISource{}
}

func (s *AuditLogAPISource) Init(ctx context.Context, configData config_data.ConfigData, opts ...row_source.RowSourceOption) error {
	// set the collection state ctor
	s.NewCollectionStateFunc = collection_state.NewTimeRangeCollectionState

	// call base init
	return s.RowSourceImpl.Init(ctx, configData, opts...)
}

func (s *AuditLogAPISource) Identifier() string {
	return AuditLogAPISourceIdentifier
}

func (s *AuditLogAPISource) GetConfigSchema() parse.Config {
	return &AuditLogAPISourceConfig{}
}

func (s *AuditLogAPISource) Collect(ctx context.Context) error {
	collectionState := s.CollectionState.(*collection_state.TimeRangeCollectionState[*AuditLogAPISourceConfig])
	// TODO: #config the below should be settable via a config option
	collectionState.IsChronological = true
	collectionState.HasContinuation = true
	startTime := collectionState.GetLatestEndTime()
	projectID := s.Config.Project
	logType := s.Config.LogType

	var opts []option.ClientOption
	if s.Config.Credentials != nil {
		opts = append(opts, option.WithCredentialsFile(*s.Config.Credentials))
	}

	client, err := logadmin.NewClient(ctx, projectID, opts...)
	if err != nil {
		return err
	}
	defer client.Close()

	if startTime == nil {
		// TODO: #config figure out how to determine an appropriate start time if collectionState doesn't provide one, defaulting to 30 days in meantime
		st := time.Now().Add(-720 * time.Hour)
		startTime = &st
	}

	sourceEnrichmentFields := &enrichment.CommonFields{
		TpIndex: projectID,
		// TODO: #finish determine if we can establish more source enrichment fields
	}

	filter := s.getLogNameFilter(projectID, logType, *startTime)
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

func (s *AuditLogAPISource) getLogNameFilter(projectId string, logType string, startTime time.Time) string {
	activity := fmt.Sprintf(`"projects/%s/logs/cloudaudit.googleapis.com%sactivity"`, projectId, "%2F")
	dataAccess := fmt.Sprintf(`"projects/%s/logs/cloudaudit.googleapis.com%sdata_access"`, projectId, "%2F")
	systemEvent := fmt.Sprintf(`"projects/%s/logs/cloudaudit.googleapis.com%ssystem_event"`, projectId, "%2F")
	timePart := fmt.Sprintf(`AND (timestamp > "%s")`, startTime.Format(time.RFC3339Nano))

		switch logType {
		case "activity":
			return fmt.Sprintf("logName=%s %s", activity, timePart)
		case "data_access":
			return fmt.Sprintf("logName=%s %s", dataAccess, timePart)
		case "system_event":
			return fmt.Sprintf("logName=%s %s", systemEvent, timePart)
		default:
			return fmt.Sprintf("logName=%s %s", logType, timePart)
		}
}
