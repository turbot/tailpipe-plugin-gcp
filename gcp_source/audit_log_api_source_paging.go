package gcp_source

import (
	"fmt"
	"maps"
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/paging"
)

type AuditLogApiPagingEntry struct {
	Timestamp      time.Time `json:"timestamp"`
	LastInsertedId string    `json:"last_inserted_id"`
}

func NewAuditLogApiPagingEntry(timestamp time.Time, lastInsertedId string) *AuditLogApiPagingEntry {
	return &AuditLogApiPagingEntry{
		Timestamp:      timestamp,
		LastInsertedId: lastInsertedId,
	}
}

type AuditLogApiPaging struct {
	AuditLogTypes map[string]AuditLogApiPagingEntry `json:"audit_log_types"`
}

func NewAuditLogApiPaging() *AuditLogApiPaging {
	return &AuditLogApiPaging{
		AuditLogTypes: make(map[string]AuditLogApiPagingEntry),
	}
}

func (a *AuditLogApiPaging) Update(data paging.Data) error {
	other, ok := data.(*AuditLogApiPaging)
	if !ok {
		return fmt.Errorf("cannot update AuditLogApi paging data with %T", data)
	}
	// merge the timestamps, preferring the latest
	maps.Copy(a.AuditLogTypes, other.AuditLogTypes)
	return nil
}

func (a *AuditLogApiPaging) Add(name string, entry AuditLogApiPagingEntry) {
	if a.AuditLogTypes == nil {
		a.AuditLogTypes = make(map[string]AuditLogApiPagingEntry)
	}
	a.AuditLogTypes[name] = entry
}
