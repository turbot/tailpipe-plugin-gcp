package gcp_types

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
)

// AuditLogRow represents an enriched row ready for parquet writing
type AuditLogRow struct {
	// embed required enrichment fields
	enrichment.CommonFields

	// Mandatory fields
	Timestamp    time.Time `json:"timestamp"`
	LogName      string    `json:"log_name"`
	InsertId     string    `json:"insert_id"`
	Severity     string    `json:"severity"`
	ServiceName  string    `json:"service_name"`
	MethodName   string    `json:"method_name"`
	ResourceName string    `json:"resource_name"`

	// Optional fields
	AuthenticationPrincipal        *string            `json:"authentication_principal,omitempty"`
	ResourceType                   *string            `json:"resource_type,omitempty"`
	ResourceLabels                 *map[string]string `json:"resource_labels,omitempty"`
	RequestCallerIp                *string            `json:"request_caller_ip,omitempty"`
	RequestCallerSuppliedUserAgent *string            `json:"request_caller_supplied_user_agent,omitempty"`
	StatusCode                     *int32             `json:"status_code,omitempty"`
	StatusMessage                  *string            `json:"status_message,omitempty"`
	// TODO: #finish add the rest of the fields
}
