package audit_log_api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/logging/logadmin"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/googleapi"
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

	filter := s.getLogNameFilter(project, logTypes, s.FromTime)

	// Create retry configuration using gax
	backoff := gax.Backoff{
		Initial:    s.Config.GetInitialBackoff(),
		Max:        s.Config.GetMaxBackoff(),
		Multiplier: 2.0, // Standard exponential backoff multiplier
	}

	// Create a retryer that handles both HTTP and gRPC errors
	retryer := gax.OnErrorFunc(backoff, s.isRetryableError)

	// Implement retry mechanism for the entire collection process using gax.Invoke
	return gax.Invoke(ctx, func(ctx context.Context, settings gax.CallSettings) error {
		// Create a new iterator for each retry attempt
		it := client.Entries(ctx, logadmin.Filter(filter), logadmin.PageSize(250))

		var logEntry *logging.Entry
		for {
			logEntry, err = it.Next()
			if err != nil && errors.Is(err, iterator.Done) {
				break
			}
			if err != nil {
				return fmt.Errorf("error fetching log entries: %w", err)
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
	}, gax.WithRetry(func() gax.Retryer {
		return retryer
	}), gax.WithTimeout(time.Duration(s.Config.GetMaxRetries())*s.Config.GetMaxBackoff()))
}

// isRetryableError determines if an error should trigger a retry
func (s *AuditLogAPISource) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for Google API errors that are retryable
	var googleErr *googleapi.Error
	if errors.As(err, &googleErr) {
		switch googleErr.Code {
		case 429: // Too Many Requests - rate limiting
			return true
		case 500: // Internal Server Error
			return true
		case 502: // Bad Gateway
			return true
		case 503: // Service Unavailable
			return true
		case 504: // Gateway Timeout
			return true
		default:
			return false
		}
	}

	// Check for network-related errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() || netErr.Temporary() {
			return true
		}
	}

	// Check for DNS errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return dnsErr.Temporary()
	}

	// Check for connection-related syscall errors
	var syscallErr *net.OpError
	if errors.As(err, &syscallErr) {
		if syscallErr.Err == syscall.ECONNRESET ||
			syscallErr.Err == syscall.ECONNREFUSED ||
			syscallErr.Err == syscall.ETIMEDOUT {
			return true
		}
	}

	// Check for context deadline exceeded (but not context canceled)
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Check for specific error messages that indicate transient issues
	errStr := strings.ToLower(err.Error())
	transientMessages := []string{
		"connection reset",
		"connection refused",
		"timeout",
		"temporary failure",
		"service unavailable",
		"try again",
		"rate limit",
		"quota exceeded",
	}

	for _, msg := range transientMessages {
		if strings.Contains(errStr, msg) {
			return true
		}
	}

	return false
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
	policyDenied := fmt.Sprintf(`"projects/%s/logs/cloudaudit.googleapis.com%spolicy"`, projectId, "%2F")
	timePart := fmt.Sprintf(`AND (timestamp > "%s")`, startTime.Format(time.RFC3339Nano))

	// short-circuit default
	if len(logTypes) == 0 {
		return fmt.Sprintf("logName=(%s OR %s OR %s OR %s) %s", activity, dataAccess, systemEvent, policyDenied, timePart)
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
		}
	}

	switch len(selected) {
	case 0:
		return fmt.Sprintf("logName=(%s OR %s OR %s or %s) %s", activity, dataAccess, systemEvent, policyDenied, timePart)
	case 1:
		return fmt.Sprintf("logName=%s %s", selected[0], timePart)
	default:
		return fmt.Sprintf("logName=(%s) %s", strings.Join(selected, " OR "), timePart)
	}
}
