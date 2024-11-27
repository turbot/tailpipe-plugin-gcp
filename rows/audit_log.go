package rows

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
)

// AuditLog represents an enriched row ready for parquet writing
type AuditLog struct {
	// embed required enrichment fields
	enrichment.CommonFields

	// Mandatory fields
	Timestamp time.Time `json:"timestamp"`
	LogName   string    `json:"log_name"`
	InsertId  string    `json:"insert_id"`
	Severity  string    `json:"severity"`

	// Optional fields
	ServiceName        *string                     `json:"service_name,omitempty"`
	MethodName         *string                     `json:"method_name,omitempty"`
	ResourceName       *string                     `json:"resource_name,omitempty"`
	AuthenticationInfo *AuditLogAuthenticationInfo `json:"authentication_info,omitempty"`
	Status             *AuditLogStatus             `json:"status,omitempty"`
	Resource           *AuditLogResource           `json:"resource,omitempty"`
	Operation          *AuditLogOperation          `json:"operation,omitempty"`
	RequestMetadata    *AuditLogRequestMetadata    `json:"request_metadata,omitempty"`
	HttpRequest        *AuditLogHttpRequest        `json:"http_request,omitempty"`
}

func NewAuditLog() *AuditLog {
	return &AuditLog{}
}

type AuditLogAuthenticationInfo struct {
	PrincipalEmail        string `json:"principal_email"`
	PrincipalSubject      string `json:"principal_subject"`
	AuthoritySelector     string `json:"authority_selector"`
	ServiceAccountKeyName string `json:"service_account_key_name"`
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
	Method       string `json:"method"`
	Url          string `json:"url"`
	Size         int64  `json:"size"`
	Status       int    `json:"status"`
	ResponseSize int64  `json:"response_size"`
	LocalIp      string `json:"local_ip"`
	RemoteIp     string `json:"remote_ip"`
}

type AuditLogRequestMetadata struct {
	CallerIp                string `json:"caller_ip"`
	CallerSuppliedUserAgent string `json:"caller_supplied_user_agent"`
}
