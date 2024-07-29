package gcp_source

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/logging/logadmin"
	"github.com/turbot/tailpipe-plugin-gcp/gcp_types"
	"github.com/turbot/tailpipe-plugin-sdk/artifact"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
	"github.com/turbot/tailpipe-plugin-sdk/plugin"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// AuditLogAPISource source is responsible for collecting audit logs from GCP
type AuditLogAPISource struct {
	row_source.Base

	Config *gcp_types.AuditLogCollectionConfig
}

func NewAuditLogAPISource(config *gcp_types.AuditLogCollectionConfig) plugin.RowSource {
	return &AuditLogAPISource{
		Config: config,
	}
}

func (s *AuditLogAPISource) Identifier() string {
	return "gcp_audit_log_api_source"
}

func (s *AuditLogAPISource) Collect(ctx context.Context) error {
	var opts []option.ClientOption
	projectID := s.Config.Project
	logTypes := []string{"activity", "data_access", "system_event", "policy"}
	var startTime *time.Time

	if s.Config.Credentials != nil {
		opts = append(opts, option.WithCredentialsFile(*s.Config.Credentials))
	}

	client, err := logadmin.NewClient(ctx, projectID, opts...)
	if err != nil {
		return err
	}
	defer client.Close()

	for _, logType := range logTypes {
		logName := fmt.Sprintf("projects/%s/logs/cloudaudit.googleapis.com%s%s", projectID, "%2f", logType)
		sourceEnrichmentFields := &enrichment.CommonFields{
			TpConnection: logName,
		}
		// TODO: #paging fetch startTime/lastInsertId from paging state per bucket (logName)
		tempStartTime := time.Now().Add(-time.Hour * 24)
		startTime = &tempStartTime
		lastInsertId := "0"

		filter := fmt.Sprintf(`logName="%s"`, logName)
		if startTime != nil {
			st := startTime.Format(time.RFC3339)
			filter += fmt.Sprintf(` AND (timestamp > "%s" OR (timestamp = "%s" AND insertId>"%s"))`, st, st, lastInsertId)
		}

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
				if err := s.OnRow(ctx, row, nil); err != nil {
					return fmt.Errorf("error processing row: %w", err)
				}

				// TODO: #paging update startTime/lastInsertId in paging state for the bucket
			}

		}
	}
	return nil
}
