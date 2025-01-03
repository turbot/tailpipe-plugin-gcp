package rows

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/schema"
	"google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/genproto/googleapis/rpc/context/attribute_context"
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
	ServiceName           *string                     `json:"service_name,omitempty"`
	MethodName            *string                     `json:"method_name,omitempty"`
	ResourceName          *string                     `json:"resource_name,omitempty"`
	ResourceLocation      *AuditLogResourceLocation   `json:"resource_location,omitempty"`
	AuthenticationInfo    *AuditLogAuthenticationInfo `json:"authentication_info,omitempty"`
	Status                *AuditLogStatus             `json:"status,omitempty"`
	Resource              *AuditLogResource           `json:"resource,omitempty"`
	Operation             *AuditLogOperation          `json:"operation,omitempty"`
	RequestMetadata       *AuditLogRequestMetadata    `json:"request_metadata,omitempty"`
	HttpRequest           *AuditLogHttpRequest        `json:"http_request,omitempty"`
	SourceLocation        *AuditLogSourceLocation     `json:"source_location,omitempty"`
	Labels                *map[string]string          `json:"labels,omitempty" parquet:"type=JSON"`
	NumResponseItems      *int64                      `json:"num_response_items,omitempty"`
	AuthorizationInfo     []*audit.AuthorizationInfo  `json:"authorization_info,omitempty" parquet:"type=JSON"`
	PolicyViolationInfo   *audit.PolicyViolationInfo  `json:"policy_violation_info,omitempty" parquet:"type=JSON"` // nested map/[]struct
	ResourceOriginalState interface{}                 `json:"resource_original_state,omitempty" parquet:"type=JSON"`
	Request               map[string]interface{}      `json:"request,omitempty" parquet:"type=JSON"`
	Response              map[string]interface{}      `json:"response,omitempty" parquet:"type=JSON"`
	Metadata              map[string]interface{}      `json:"metadata,omitempty" parquet:"type=JSON"`
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
}

type AuditLogRequestMetadata struct {
	CallerIp                string                                        `json:"caller_ip"`
	CallerSuppliedUserAgent string                                        `json:"caller_supplied_user_agent"`
	CallerNetwork           string                                        `json:"caller_network"`
	RequestAttributes       *attribute_context.AttributeContext_Request   `json:"request_attributes,omitempty" parquet:"type=JSON"`
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

// Bucket source log

type BucketSourceLogEntry struct {
	InsertID         *string       `json:"insertId"`
	LogName          string        `json:"logName"`
	ProtoPayload     *ProtoPayload `json:"protoPayload"`
	ReceiveTimestamp string        `json:"receiveTimestamp"`
	Resource         *Resource     `json:"resource"`
	Severity         string        `json:"severity"`
	Timestamp        time.Time     `json:"timestamp"`
}

type ProtoPayload struct {
	Type               *string              `json:"@type"`
	AuthenticationInfo *AuthenticationInfo  `json:"authenticationInfo"`
	AuthorizationInfo  *[]AuthorizationInfo `json:"authorizationInfo"`
	MethodName         *string              `json:"methodName"`
	RequestMetadata    *RequestMetadata     `json:"requestMetadata"`
	ResourceLocation   *ResourceLocation    `json:"resourceLocation"`
	ResourceName       *string              `json:"resourceName"`
	ServiceName        *string              `json:"serviceName"`
	Status             *Status              `json:"status"`
}

type AuthenticationInfo struct {
	PrincipalEmail               *string           `json:"principalEmail"`
	PrincipalSubject             string            `json:"principal_subject"`
	AuthoritySelector            string            `json:"authority_selector"`
	ServiceAccountKeyName        string            `json:"service_account_key_name"`
	ThirdPartyPrincipal          map[string]string `json:"third_party_principal,omitempty" parquet:"type=JSON"`
	ServiceAccountDelegationInfo []string          `json:"service_account_delegation_info,omitempty" parquet:"type=JSON"`
}

type AuthorizationInfo struct {
	Granted            *bool              `json:"granted"`
	Permission         *string            `json:"permission"`
	Resource           *string            `json:"resource"`
	ResourceAttributes *map[string]string `json:"resourceAttributes"`
}

type RequestMetadata struct {
	CallerIP                *string            `json:"callerIp"`
	CallerSuppliedUserAgent *string            `json:"callerSuppliedUserAgent"`
	DestinationAttributes   *map[string]string `json:"destinationAttributes"`
	RequestAttributes       *RequestAttributes `json:"requestAttributes"`
}

type RequestAttributes struct {
	Auth *map[string]interface{} `json:"auth"`
	Time *string                 `json:"time"`
}

type ResourceLocation struct {
	CurrentLocations *[]string `json:"currentLocations"`
}

type Status struct {
	Code    *int    `json:"code"`
	Message *string `json:"message"`
}

type Resource struct {
	Labels *map[string]string `json:"labels"`
	Type   *string         `json:"type"`
}

type ResourceLabels struct {
	BucketName *string `json:"bucket_name"`
	Location   *string `json:"location"`
	ProjectID  *string `json:"project_id"`
}
