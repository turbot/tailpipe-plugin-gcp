package extractors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
)

// ActivityLogExtractor is an extractor that receives JSON serialised ActivityLogBatch objects
// and extracts ActivityLog records from them
type ActivityLogExtractor struct {
}

// NewActivityLogExtractor creates a new ActivityLogExtractor
func NewActivityLogExtractor() artifact_source.Extractor {
	return &ActivityLogExtractor{}
}

func (c *ActivityLogExtractor) Identifier() string {
	return "activity_log_extractor"
}

// Extract unmarshalls the artifact data as an ActivityLogBatch and returns the ActivityLog records
func (c *ActivityLogExtractor) Extract(_ context.Context, a any) ([]any, error) {
	// the expected input type is a JSON byte[] deserializable to ActivityLogBatch
	jsonBytes, ok := a.([]byte)
	if !ok {
		return nil, fmt.Errorf("expected byte[], got %T", a)
	}

	// decode json ito ActivityLogBatch
	var log []rows.AuditLog
	slog.Error("Result ===>>>>> ", string(jsonBytes), "Found")
	err := json.Unmarshal(jsonBytes, &log)
	if err != nil {
		return nil, fmt.Errorf("error decoding json: %w", err)
	}

	slog.Debug("ActivityLogExtractor", "record count", len(log))
	var res = make([]any, len(log))
	for i, record := range log {
		res[i] = &record
	}
	return res, nil
}
