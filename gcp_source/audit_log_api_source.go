package gcp_source

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/logging/logadmin"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/types"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const AuditLogAPISourceIdentifier = "gcp_audit_log_api"

// AuditLogAPISource source is responsible for collecting audit logs from GCP
type AuditLogAPISource struct {
	row_source.RowSourceBase[AuditLogAPISourceConfig]
}

func NewAuditLogAPISource() row_source.RowSource {
	return &AuditLogAPISource{}
}

func (s *AuditLogAPISource) Identifier() string {
	return AuditLogAPISourceIdentifier
}

func (s *AuditLogAPISource) Collect(ctx context.Context) error {
	// TODO: #validation validate the configuration

	paging := s.PagingData.(*AuditLogApiPaging)

	startTime := paging.Timestamp
	projectID := s.Config.Project
	logTypes := s.Config.LogTypes

	var opts []option.ClientOption
	if s.Config.Credentials != nil {
		opts = append(opts, option.WithCredentialsFile(*s.Config.Credentials))
	}

	client, err := logadmin.NewClient(ctx, projectID, opts...)
	if err != nil {
		return err
	}
	defer client.Close()

	// TODO: #finish figure out how to determine an appropriate start time if paging doesn't provide one
	if startTime == nil {
		tempTime := time.Now().Add(-time.Hour * 24)
		startTime = &tempTime
	}

	sourceEnrichmentFields := &enrichment.CommonFields{
		TpConnection: projectID,
		// TODO: #finish determine if we can establish more source enrichment fields
	}

	filter := s.getLogNameFilter(projectID, logTypes)
	if startTime != nil {
		filter += fmt.Sprintf(` AND (timestamp > "%s")`, startTime.Format(time.RFC3339Nano))
	}

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
			row := &types.RowData{
				Data:     *logEntry,
				Metadata: sourceEnrichmentFields,
			}
			paging.Timestamp = &logEntry.Timestamp

			if err := s.OnRow(ctx, row, paging); err != nil {
				return fmt.Errorf("error processing row: %w", err)
			}
		}
	}

	return nil
}

func (s *AuditLogAPISource) getLogNameFilter(projectId string, logTypes []string) string {
	activity := fmt.Sprintf(`"projects/%s/logs/cloudaudit.googleapis.com%sactivity"`, projectId, "%2F")
	dataAccess := fmt.Sprintf(`"projects/%s/logs/cloudaudit.googleapis.com%sdata_access"`, projectId, "%2F")
	systemEvent := fmt.Sprintf(`"projects/%s/logs/cloudaudit.googleapis.com%ssystem_event"`, projectId, "%2F")

	// short-circuit default
	if len(logTypes) == 0 {
		return fmt.Sprintf("logName=(%s OR %s OR %s)", activity, dataAccess, systemEvent)
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
		return fmt.Sprintf("logName=(%s OR %s OR %s)", activity, dataAccess, systemEvent)
	case 1:
		return fmt.Sprintf("logName=%s", selected[0])
	default:
		return fmt.Sprintf("logName=(%s)", strings.Join(selected, " OR "))
	}
}
