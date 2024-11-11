package rows

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
	"github.com/turbot/tailpipe-plugin-sdk/helpers"
)

// AuditActivityLog represents an enriched row ready for parquet writing
type AuditActivityLog struct {
	// embed required enrichment fields
	enrichment.CommonFields

	// Mandatory fields
	Timestamp    time.Time           `json:"timestamp"`
	LogName      string              `json:"log_name"`
	InsertId     string              `json:"insert_id"`
	Severity     string              `json:"severity"`
	ProtoPayload *helpers.JSONString `json:"proto_payload"`
	// Optional fields

	ResourceType *string `json:"resource_type,omitempty"`
	//ResourceLabels                 *map[string]string `json:"resource_labels,omitempty"` // TODO: #finish add back in once we have support for map
	OperationId             *string `json:"operation_id,omitempty"`
	OperationProducer       *string `json:"operation_producer,omitempty"`
	OperationFirst          *bool   `json:"operation_first,omitempty"`
	OperationLast           *bool   `json:"operation_last,omitempty"`
	RequestMethod           string  `json:"request_url,omitempty"`
	RequestSize             int64   `json:"request_size,omitempty"`
	RequestStatus           int     `json:"request_status,omitempty"`
	RequestCallerIp         *string `json:"request_caller_ip,omitempty"`
	RequestResponseSize     int64   `json:"request_response_size,omitempty"`
	AuthenticationPrincipal *string `json:"authentication_principal,omitempty"`
}
