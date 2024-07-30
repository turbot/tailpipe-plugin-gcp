package gcp_source

import (
	"context"
	"errors"
	"fmt"
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
	pg, hasPaging := context_values.PagingDataFromContext[*AuditLogApiPaging](ctx)
	if hasPaging {
		s.paging = *pg
	}
	var opts []option.ClientOption
	projectID := s.Config.Project
	logTypes := []string{"activity", "data_access", "system_event", "policy"}

	if s.Config.Credentials != nil {
		opts = append(opts, option.WithCredentialsFile(*s.Config.Credentials))
	}

	client, err := logadmin.NewClient(ctx, projectID, opts...)
	if err != nil {
		return err
	}
	defer client.Close()

	for _, logType := range logTypes {
		var startTime *time.Time
		lastInsertId := "0"

		if pagingEntry, ok := s.paging.AuditLogTypes[logType]; ok {
			startTime = &pagingEntry.Timestamp
			lastInsertId = pagingEntry.LastInsertedId
		}

		logName := fmt.Sprintf("projects/%s/logs/cloudaudit.googleapis.com%s%s", projectID, "%2f", logType)
		sourceEnrichmentFields := &enrichment.CommonFields{
			TpConnection: projectID,
		}

		// TODO: #finish figure out how to determine an appropriate start time if paging doesn't provide one
		if startTime == nil {
			tempTime := time.Now().Add(-time.Hour * 12)
			startTime = &tempTime
		}

		filter := fmt.Sprintf(`logName="%s"`, logName)
		if startTime != nil {
			st := startTime.Format(time.RFC3339)
			filter += fmt.Sprintf(` AND (timestamp > "%s" OR (timestamp = "%s" AND insertId>"%s"))`, st, st, lastInsertId)
		}

		// TODO: #ratelimit implement rate limiting - see https://pkg.go.dev/google.golang.org/api/option#RateLimiter
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

				s.paging.Add(logType, *NewAuditLogApiPagingEntry(logEntry.Timestamp, logEntry.InsertID))

				if err := s.OnRow(ctx, row, &s.paging); err != nil {
					return fmt.Errorf("error processing row: %w", err)
				}
			}

		}
	}

	return nil
}

func (s *AuditLogAPISource) GetPagingData() paging.Data {
	return &s.paging
}
