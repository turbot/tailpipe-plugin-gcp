package audit_log

import (
	"time"

	"google.golang.org/genproto/googleapis/cloud/audit"

	"github.com/turbot/tailpipe-plugin-sdk/schema"
)

// AuditLog represents an enriched row ready for parquet writing
type AuditLog struct {
	// embed required enrichment fields
	schema.CommonFields

	// Mandatory fields
	Timestamp    time.Time `json:"timestamp"`
	LogName      string    `json:"log_name"`
	InsertId     string    `json:"insert_id"`
	Severity     string    `json:"severity"`
	Trace        string    `json:"trace"`
	TraceSampled bool      `json:"trace_sampled"`
	SpanId       string    `json:"span_id"`

	// Optional fields
	ServiceName           *string                      `json:"service_name,omitempty"`
	MethodName            *string                      `json:"method_name,omitempty"`
	ResourceName          *string                      `json:"resource_name,omitempty"`
	ResourceLocation      *AuditLogResourceLocation    `json:"resource_location,omitempty"`
	AuthenticationInfo    *AuditLogAuthenticationInfo  `json:"authentication_info,omitempty"`
	Status                *AuditLogStatus              `json:"status,omitempty"`
	Resource              *AuditLogResource            `json:"resource,omitempty"`
	Operation             *AuditLogOperation           `json:"operation,omitempty"`
	RequestMetadata       *AuditLogRequestMetadata     `json:"request_metadata,omitempty"`
	HttpRequest           *AuditLogHttpRequest         `json:"http_request,omitempty"`
	SourceLocation        *AuditLogSourceLocation      `json:"source_location,omitempty"`
	Labels                *map[string]string           `json:"labels,omitempty" parquet:"type=JSON"`
	NumResponseItems      *int64                       `json:"num_response_items,omitempty"`
	AuthorizationInfo     []*AuditLogAuthorizationInfo `json:"authorization_info,omitempty" parquet:"type=JSON"`
	PolicyViolationInfo   *audit.PolicyViolationInfo   `json:"policy_violation_info,omitempty" parquet:"type=JSON"` // nested map/[]struct
	ResourceOriginalState interface{}                  `json:"resource_original_state,omitempty" parquet:"type=JSON"`
	Request               map[string]interface{}       `json:"request,omitempty" parquet:"type=JSON"`
	Response              map[string]interface{}       `json:"response,omitempty" parquet:"type=JSON"`
	Metadata              map[string]interface{}       `json:"metadata,omitempty" parquet:"type=JSON"`
	ServiceData           *map[string]interface{}      `json:"service_data,omitempty" parquet:"type=JSON"`
}

func NewAuditLog() *AuditLog {
	return &AuditLog{}
}

type AuditLogAuthenticationInfo struct {
	PrincipalEmail               string            `json:"principal_email"`
	PrincipalSubject             string            `json:"principal_subject"`
	AuthoritySelector            string            `json:"authority_selector"`
	ServiceAccountKeyName        string            `json:"service_account_key_name"`
	ThirdPartyPrincipal          map[string]string `json:"third_party_principal,omitempty" parquet:"type=JSON"`
	ServiceAccountDelegationInfo []string          `json:"service_account_delegation_info,omitempty" parquet:"type=JSON"`
}

type AuditLogStatus struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type AuditLogResource struct {
	Type   string            `json:"type"`
	Labels map[string]string `json:"labels" parquet:"type=JSON"`
}

type AuditLogOperation struct {
	Id       string `json:"id"`
	Producer string `json:"producer"`
	First    bool   `json:"first"`
	Last     bool   `json:"last"`
}

type AuditLogHttpRequest struct {
	Method                         string              `json:"method"`
	Url                            string              `json:"url"`
	RequestSize                    int64               `json:"request_size"`
	RequestHeaders                 map[string][]string `json:"request_headers" parquet:"type=JSON"`
	Status                         int                 `json:"status"`
	ResponseSize                   int64               `json:"response_size"`
	LocalIp                        string              `json:"local_ip"`
	RemoteIp                       string              `json:"remote_ip"`
	Latency                        string              `json:"latency"`
	CacheHit                       bool                `json:"cache_hit"`
	CacheLookup                    bool                `json:"cache_lookup"`
	CacheValidatedWithOriginServer bool                `json:"cache_validated_with_origin_server"`
	CacheFillBytes                 int64               `json:"cache_fill_bytes"`
	UserAgent                      *string             `json:"user_agent,omitempty"`
}

type AuditLogRequestMetadata struct {
	CallerIp                string                                        `json:"caller_ip"`
	CallerSuppliedUserAgent string                                        `json:"caller_supplied_user_agent"`
	CallerNetwork           string                                        `json:"caller_network"`
	RequestAttributes       *map[string]interface{}                       `json:"request_attributes,omitempty" parquet:"type=JSON"`
	DestinationAttributes   *AuditLogRequestMetadataDestinationAttributes `json:"destination_attributes,omitempty"`
}

type AuditLogSourceLocation struct {
	File     string `json:"file"`
	Line     int64  `json:"line"`
	Function string `json:"function"`
}

type AuditLogResourceLocation struct {
	CurrentLocations  []string `json:"current_locations" parquet:"type=JSON"`
	OriginalLocations []string `json:"original_locations" parquet:"type=JSON"`
}

type AuditLogRequestMetadataDestinationAttributes struct {
	Ip         string            `json:"ip"`
	Port       int64             `json:"port"`
	Principal  string            `json:"principal"`
	RegionCode string            `json:"region_code"`
	Labels     map[string]string `json:"labels" parquet:"type=JSON"`
}

type AuditLogAuthorizationInfo struct {
	Resource   string `json:"resource"`
	Permission string `json:"permission"`
	Granted    bool   `json:"granted"`
}

func (a *AuditLog) GetColumnDescriptions() map[string]string {
	return map[string]string{
		"timestamp":               "The date and time when the event occurred, in ISO 8601 format.",
		"log_name":                "The name of the log that recorded the event, indicating the type of log (e.g., 'cloudaudit.googleapis.com/activity').",
		"insert_id":               "A unique identifier for the log entry, used to prevent duplicate log entries.",
		"severity":                "The severity level of the log entry (e.g., 'INFO', 'WARNING', 'ERROR', 'CRITICAL').",
		"trace":                   "The unique trace ID associated with the request, used for distributed tracing.",
		"trace_sampled":           "Indicates whether the request trace was sampled for analysis (true or false).",
		"span_id":                 "The span ID for the request, used in distributed tracing to identify specific operations.",
		"service_name":            "The Google Cloud service that handled the request, such as 'compute.googleapis.com'.",
		"method_name":             "The API method or operation that was invoked, such as 'google.iam.v1.GetPolicy'.",
		"resource_name":           "The full resource name of the affected Google Cloud resource, in the format 'projects/123456789012/buckets/my-bucket'.",
		"resource_location":       "The geographic location of the affected resource, if applicable.",
		"authentication_info":     "Details about the authenticated user or service account that made the request.",
		"status":                  "The status of the request, including error codes if the request failed.",
		"resource":                "Detailed metadata about the affected resource, such as type and labels.",
		"operation":               "Information about the larger operation that this event is a part of, if applicable.",
		"request_metadata":        "Metadata about the request, including caller IP and user agent.",
		"http_request":            "Details about the HTTP request associated with the log entry, if applicable.",
		"source_location":         "The location in the source code where the request originated, if available.",
		"labels":                  "Key-value labels associated with the log entry for filtering and analysis.",
		"num_response_items":      "The number of items returned in the response, if applicable.",
		"authorization_info":      "Details about the authorization checks performed for the request, including granted and denied permissions.",
		"policy_violation_info":   "Information about any policy violations detected in the request.",
		"resource_original_state": "The original state of the resource before the request was processed, if available.",
		"request":                 "The request parameters sent with the API call, in JSON format.",
		"response":                "The response data returned by the service, in JSON format.",
		"metadata":                "Additional metadata related to the log entry, in JSON format.",
		"service_data":            "Additional service-specific data related to the event, in JSON format.",

		// Override table specific tp_* column descriptions
		"tp_index": "The GCP project.",
	}

}
