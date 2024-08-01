package gcp_source

import (
	"fmt"
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/paging"
)

type AuditLogApiPaging struct {
	Timestamp *time.Time `json:"timestamp"`
}

func NewAuditLogApiPaging() *AuditLogApiPaging {
	return &AuditLogApiPaging{}
}

func (a *AuditLogApiPaging) Update(data paging.Data) error {
	other, ok := data.(*AuditLogApiPaging)
	if !ok {
		return fmt.Errorf("cannot update AuditLogApi paging data with %T", data)
	}
	if other.Timestamp != nil {
		a.Timestamp = other.Timestamp
	}
	return nil
}
