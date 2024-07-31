package gcp_source

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/logging/logadmin"
	"github.com/turbot/tailpipe-plugin-gcp/gcp_types"
	"github.com/turbot/tailpipe-plugin-sdk/artifact"
	"github.com/turbot/tailpipe-plugin-sdk/context_values"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
	"github.com/turbot/tailpipe-plugin-sdk/paging"
	"github.com/turbot/tailpipe-plugin-sdk/plugin"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// AuditLogAPISource source is responsible for collecting audit logs from GCP
type AuditLogAPISource struct {
	row_source.Base
	paging AuditLogApiPaging
	Config *gcp_types.AuditLogCollectionConfig
}

func NewAuditLogAPISource(config *gcp_types.AuditLogCollectionConfig) plugin.RowSource {
	return &AuditLogAPISource{
		Config: config,
		paging: *NewAuditLogApiPaging(),
	}
}

func (s *AuditLogAPISource) Identifier() string {
	return "gcp_audit_log_api_source"
}

func (s *AuditLogAPISource) Collect(ctx context.Context) error {
	// TODO: #validation validate the configuration

	pg, hasPaging := context_values.PagingDataFromContext[*AuditLogApiPaging](ctx)
	if hasPaging {
		s.paging = *pg
	}

	startTime := s.paging.Timestamp
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
		filter += fmt.Sprintf(` AND (timestamp > "%s")`, startTime.Format(time.RFC3339))
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
			row := &artifact.ArtifactData{
				Data:     *logEntry,
				Metadata: sourceEnrichmentFields,
			}
			s.paging.Timestamp = &logEntry.Timestamp

			if err := s.OnRow(ctx, row, &s.paging); err != nil {
				return fmt.Errorf("error processing row: %w", err)
			}
		}
	}

	return nil
}

func (s *AuditLogAPISource) GetPagingData() paging.Data {
	return &s.paging
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
