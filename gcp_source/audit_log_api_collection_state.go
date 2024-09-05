package gcp_source

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/collection_state"
)

type AuditLogAPICollectionState struct {
	collection_state.CollectionStateBase

	StartTime time.Time `json:"start_time,omitempty"` // oldest record timestamp
	EndTime   time.Time `json:"end_time,omitempty"`   // newest record timestamp
}

func NewAuditLogApiPaging() collection_state.CollectionState[*AuditLogAPISourceConfig] {
	return &AuditLogAPICollectionState{}
}

func (s *AuditLogAPICollectionState) Init(*AuditLogAPISourceConfig) error {
	return nil
}

func (s *AuditLogAPICollectionState) IsEmpty() bool {
	return s.StartTime.IsZero() // && s.EndTime.IsZero()
}

func (s *AuditLogAPICollectionState) Upsert(ts time.Time) {
	if s.StartTime.IsZero() || ts.Before(s.StartTime) {
		s.StartTime = ts
	}

	if s.EndTime.IsZero() || ts.After(s.EndTime) {
		s.EndTime = ts
	}
}
