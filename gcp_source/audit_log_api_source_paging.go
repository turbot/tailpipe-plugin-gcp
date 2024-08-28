package gcp_source

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/collection_state"
)

type AuditLogApiCollectionState struct {
	collection_state.CollectionStateBase
	Timestamp *time.Time `json:"timestamp"`
}

func NewAuditLogApiPaging() collection_state.CollectionState {
	return &AuditLogApiCollectionState{}
}
