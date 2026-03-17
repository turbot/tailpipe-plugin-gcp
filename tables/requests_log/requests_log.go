package requests_log

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/schema"
)

// RequestsLog represents an enriched row ready for parquet writing
type RequestsLog struct {
	// embed required enrichment fields
	schema.CommonFields

	// Mandatory fields
	Timestamp        time.Time `json:"timestamp"`
	ReceiveTimestamp time.Time `json:"receive_timestamp"`
	LogName          string    `json:"log_name"`
	InsertId         string    `json:"insert_id"`
	Severity         string    `json:"severity"`
	Trace            string    `json:"trace_id"`
	SpanId           string    `json:"span_id"`
	TraceSampled     bool      `json:"trace_sampled"`

	// the json payload fields from the requests log, moved to the top level
	BackendTargetProjectNumber string                               `json:"backend_target_project_number,omitempty"`
	CacheDecision              []string                             `json:"cache_decision,omitempty"`
	RemoteIp                   string                               `json:"remote_ip,omitempty"`
	StatusDetails              string                               `json:"status_details,omitempty"`
	EnforcedSecurityPolicy     *RequestLogSecurityPolicy            `json:"enforced_security_policy" parquet:"type=JSON"`
	PreviewSecurityPolicy      *RequestLogSecurityPolicy            `json:"preview_security_policy,omitempty" parquet:"type=JSON"`
	SecurityPolicyRequestData  *RequestLogSecurityPolicyRequestData `json:"security_policy_request_data,omitempty" parquet:"type=JSON"`
	EnforcedEdgeSecurityPolicy *RequestLogEdgeSecurityPolicy        `json:"enforced_edge_security_policy,omitempty" parquet:"type=JSON"`
	PreviewEdgeSecurityPolicy  *RequestLogEdgeSecurityPolicy        `json:"preview_edge_security_policy,omitempty" parquet:"type=JSON"`

	// other top level fields
	Resource    *RequestLogResource    `json:"resource,omitempty" parquet:"type=JSON"`
	HttpRequest *RequestLogHttpRequest `json:"http_request,omitempty" parquet:"type=JSON"`
}

func NewRequestsLog() *RequestsLog {
	return &RequestsLog{}
}

type RequestLogResource struct {
	Type   string            `json:"type,omitempty"`
	Labels map[string]string `json:"labels,omitempty" parquet:"type=JSON"`
}

type RequestLogHttpRequest struct {
	RequestMethod                  string `json:"request_method,omitempty"`
	RequestUrl                     string `json:"request_url,omitempty"`
	RequestSize                    int64  `json:"request_size,omitempty"`
	Referer                        string `json:"referer,omitempty"`
	Status                         int32  `json:"status,omitempty"`
	ResponseSize                   int64  `json:"response_size,omitempty"`
	RemoteIp                       string `json:"remote_ip,omitempty"`
	ServerIp                       string `json:"server_ip,omitempty"`
	Latency                        string `json:"latency,omitempty"`
	Protocol                       string `json:"protocol,omitempty"`
	CacheHit                       bool   `json:"cache_hit,omitempty"`
	CacheLookup                    bool   `json:"cache_lookup,omitempty"`
	CacheValidatedWithOriginServer bool   `json:"cache_validated_with_origin_server,omitempty"`
	CacheFillBytes                 int64  `json:"cache_fill_bytes,omitempty"`
	UserAgent                      string `json:"user_agent,omitempty"`
}

type RequestLogSecurityPolicy struct {
	ConfiguredAction    string                        `json:"configured_action,omitempty"`
	RateLimitAction     *RequestLogRateLimitAction    `json:"rate_limit_action,omitempty"`
	Name                string                        `json:"name,omitempty"`
	Outcome             string                        `json:"outcome,omitempty"`
	Priority            int                           `json:"priority,omitempty"`
	PreconfiguredExprId string                        `json:"preconfigured_expr_id,omitempty"`
	ThreatIntelligence  *RequestLogThreatIntelligence `json:"threat_intelligence,omitempty"`
	AddressGroup        *RequestLogAddressGroup       `json:"address_group,omitempty"`
	MatchedFieldType    string                        `json:"matched_field_type,omitempty"`
	MatchedFieldValue   string                        `json:"matched_field_value,omitempty"`
	MatchedFieldName    string                        `json:"matched_field_name,omitempty"`
	MatchedOffset       int                           `json:"matched_offset,omitempty"`
	MatchedLength       int                           `json:"matched_length,omitempty"`
}

type RequestLogRateLimitAction struct {
	Key     string `json:"key,omitempty"`
	Outcome string `json:"outcome,omitempty"`
}

type RequestLogThreatIntelligence struct {
	Categories []string `json:"categories,omitempty"`
}

type RequestLogAddressGroup struct {
	Names []string `json:"names,omitempty"`
}

type RequestLogSecurityPolicyRequestData struct {
	RemoteIpInfo          *RequestLogRemoteIpInfo   `json:"remote_ip_info,omitempty"`
	RecaptchaActionToken  *RequestLogRecaptchaToken `json:"recaptcha_action_token,omitempty"`
	RecaptchaSessionToken *RequestLogRecaptchaToken `json:"recaptcha_session_token,omitempty"`
	TlsJa3Fingerprint     string                    `json:"tls_ja3_fingerprint,omitempty"`
	TlsJa4Fingerprint     string                    `json:"tls_ja4_fingerprint,omitempty"`
}

type RequestLogRemoteIpInfo struct {
	Asn        int    `json:"asn,omitempty"`
	RegionCode string `json:"region_code,omitempty"`
}

type RequestLogEdgeSecurityPolicy struct {
	Name             string `json:"name,omitempty"`
	Priority         int    `json:"priority,omitempty"`
	ConfiguredAction string `json:"configured_action,omitempty"`
	Outcome          string `json:"outcome,omitempty"`
}

type RequestLogRecaptchaToken struct {
	Score float64 `json:"score,omitempty"`
}

func (a *RequestsLog) GetColumnDescriptions() map[string]string {
	return map[string]string{
		"timestamp":                     "The date and time when the request was received, in ISO 8601 format.",
		"receive_timestamp":             "The time when the log entry was received by Cloud Logging.",
		"log_name":                      "The name of the log that recorded the request, e.g., 'projects/[PROJECT_ID]/logs/requests'.",
		"insert_id":                     "A unique identifier for the log entry, used to prevent duplicate log entries.",
		"severity":                      "The severity level of the log entry (e.g., 'INFO', 'WARNING', 'ERROR', 'CRITICAL').",
		"trace_id":                      "The unique trace ID associated with the request, used for distributed tracing.",
		"trace_sampled":                 "Indicates whether the request trace was sampled for analysis (true or false).",
		"span_id":                       "The span ID for the request, used in distributed tracing to identify specific operations.",
		"resource":                      "The monitored resource associated with the log entry, including type and labels.",
		"http_request":                  "Details about the HTTP request associated with the log entry, if available (present in application load balancer logs).",
		"backend_target_project_number": "The project number of the backend target.",
		"cache_decision":                "A list of cache decisions made for the request.",
		"enforced_security_policy":      "Details about the enforced security policy for the request.",
		"preview_security_policy":       "Details about the preview security policy for the request, if any.",
		"security_policy_request_data":  "Additional data about the security policy request.",
		"enforced_edge_security_policy": "Details about the enforced edge security policy for the request.",
		"preview_edge_security_policy":  "Details about the preview edge security policy for the request, if any.",
		"remote_ip":                     "The remote IP address from which the request originated.",
		"status_details":                "Additional status details for the request.",

		// Override table specific tp_* column descriptions
		"tp_index": "The GCP project.",
	}
}
