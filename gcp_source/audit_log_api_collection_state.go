package gcp_source

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/collection_state"
)

type AuditLogAPICollectionState struct {
	collection_state.CollectionStateBase

	StartTime time.Time `json:"start_time,omitempty"` // oldest record timestamp
	EndTime   time.Time `json:"end_time,omitempty"`   // newest record timestamp

	prevTime time.Time `json:"-"`
}

func NewAuditLogApiPaging() collection_state.CollectionState {
	return &AuditLogAPICollectionState{}
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

// StartCollection stores the current state as previous state
func (s *AuditLogAPICollectionState) StartCollection() {
	s.prevTime = s.EndTime
}

func (s *AuditLogAPICollectionState) ShouldCollectRow(ts time.Time) bool {
	if !s.prevTime.IsZero() && ts.Equal(s.prevTime) {
		return false
	}

	return true
}
