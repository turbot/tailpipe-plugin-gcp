package rows

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
)

// AuditDataAccessLog represents an enriched row ready for parquet writing
type AuditDataAccessLog struct {
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
	AuthenticationPrincipal *string `json:"authentication_principal,omitempty"`
	ResourceType            *string `json:"resource_type,omitempty"`
	//ResourceLabels                 *map[string]string `json:"resource_labels,omitempty"` // TODO: #finish add back in once we have support for map
	RequestCallerIp                *string `json:"request_caller_ip,omitempty"`
	RequestCallerSuppliedUserAgent *string `json:"request_caller_supplied_user_agent,omitempty"`
	StatusCode                     *int32  `json:"status_code,omitempty"`
	StatusMessage                  *string `json:"status_message,omitempty"`
	OperationId                    *string `json:"operation_id,omitempty"`
	OperationProducer              *string `json:"operation_producer,omitempty"`
	OperationFirst                 *bool   `json:"operation_first,omitempty"`
	OperationLast                  *bool   `json:"operation_last,omitempty"`
	RequestMethod                  string  `json:"request_url,omitempty"`
	RequestSize                    int64   `json:"request_size,omitempty"`
	RequestStatus                  int     `json:"request_status,omitempty"`
	RequestResponseSize            int64   `json:"request_response_size,omitempty"`
}

func NewAuditDataAccessLog() *AuditDataAccessLog {
	return &AuditDataAccessLog{}
}
