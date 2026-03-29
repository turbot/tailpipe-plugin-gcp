//nolint:staticcheck
package requests_log

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	loggingpb "cloud.google.com/go/logging/apiv2/loggingpb"

	"github.com/turbot/tailpipe-plugin-sdk/mappers"
)

type RequestsLogMapper struct {
}

func (m *RequestsLogMapper) Identifier() string {
	return "gcp_requests_log_mapper"
}

// flexibleInt64 can unmarshal from both string and int64 JSON values
type flexibleInt64 int64

func (f *flexibleInt64) UnmarshalJSON(data []byte) error {
	// Handle null or empty string
	if len(data) == 0 || string(data) == "null" || string(data) == `""` {
		*f = 0
		return nil
	}

	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		// If it's an empty string, set to 0
		if str == "" {
			*f = 0
			return nil
		}
		// If it's a string, parse it
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse string as int64: %w", err)
		}
		*f = flexibleInt64(val)
		return nil
	}

	// If not a string, try as int64
	var val int64
	if err := json.Unmarshal(data, &val); err != nil {
		return fmt.Errorf("failed to unmarshal as int64 or string: %w", err)
	}
	*f = flexibleInt64(val)
	return nil
}

// sanitizeURL converts Unicode escape sequences (e.g., %u002e) to their actual characters
// to prevent URL parsing errors. This handles malformed URLs that contain Unicode escapes
// instead of proper URL percent-encoding.
func sanitizeURL(url string) string {
	// Match Unicode escape sequences like %u002e, %u0041, etc.
	re := regexp.MustCompile(`%u([0-9a-fA-F]{4})`)
	return re.ReplaceAllStringFunc(url, func(match string) string {
		// Extract the hex code (e.g., "002e" from "%u002e")
		hexCode := match[2:6] // Skip "%u" prefix
		code, err := strconv.ParseInt(hexCode, 16, 32)
		if err != nil {
			// If parsing fails, return the original match
			return match
		}
		// Convert to Unicode character
		return string(rune(code))
	})
}

func (m *RequestsLogMapper) Map(_ context.Context, a any, _ ...mappers.MapOption[*RequestsLog]) (*RequestsLog, error) {
	switch v := a.(type) {
	case string:
		return mapFromBucketJson([]byte(v))
	case *loggingpb.LogEntry:
		return mapFromSDKType(v)
	case []byte:
		return mapFromBucketJson(v)
	default:
		return nil, fmt.Errorf("expected loggingpb.LogEntry, string or []byte, got %T", a)
	}
}

func mapFromSDKType(item *loggingpb.LogEntry) (*RequestsLog, error) {
	row := NewRequestsLog()

	// === 2. Map common LogEntry fields ===
	row.Timestamp = item.GetTimestamp().AsTime()
	row.LogName = item.GetLogName()
	row.InsertId = item.GetInsertId()
	row.Severity = item.GetSeverity().String()
	row.ReceiveTimestamp = item.GetReceiveTimestamp().AsTime()
	row.Trace = item.GetTrace()   // Access Trace directly using GetTrace()
	row.SpanId = item.GetSpanId() // Access SpanID directly using GetSpanId()
	// TraceSampled is often inferred from SpanID or Trace, or may be a specific field
	row.TraceSampled = item.GetTraceSampled() // Assuming GetTraceSampled() exists for bool, else infer as before

	// === 3. Map Resource ===
	// Access Resource fields using GetResource() and its Get* methods
	if item.GetResource() != nil { // Check if resource object exists
		row.Resource = &RequestLogResource{
			Type:   item.GetResource().GetType(),
			Labels: item.GetResource().GetLabels(),
		}
	}

	// === 4. Map JsonPayload (for LB/Cloud Armor specific data) ===
	jsonPayload := item.GetJsonPayload().AsMap()
	if jsonPayload == nil {
		jsonPayload = make(map[string]interface{})
	}

	// Safely extract string fields with type checking
	if v, ok := jsonPayload["backendTargetProjectNumber"].(string); ok {
		row.BackendTargetProjectNumber = v
	}
	// Handle CacheDecision specifically:
	if rawCacheDecision, ok := jsonPayload["cacheDecision"].([]interface{}); ok {
		// Iterate over the []interface{} and append string elements to row.CacheDecision
		for _, v := range rawCacheDecision {
			if s, ok := v.(string); ok {
				row.CacheDecision = append(row.CacheDecision, s)
			}
		}
	}

	if v, ok := jsonPayload["remoteIp"].(string); ok {
		row.RemoteIp = v
	}

	if v, ok := jsonPayload["statusDetails"].(string); ok {
		row.StatusDetails = v
	}

	// Safely extract security policy map
	if securityPolicyMap, ok := jsonPayload["enforcedSecurityPolicy"].(map[string]interface{}); ok && securityPolicyMap != nil {
		policy := &RequestLogSecurityPolicy{}
		if val, ok := securityPolicyMap["configuredAction"].(string); ok {
			policy.ConfiguredAction = val
		}
		if val, ok := securityPolicyMap["name"].(string); ok {
			policy.Name = val
		}
		if val, ok := securityPolicyMap["outcome"].(string); ok {
			policy.Outcome = val
		}
		if val, ok := securityPolicyMap["priority"].(float64); ok {
			policy.Priority = int(val)
		}

		if rawIds, ok := securityPolicyMap["preconfiguredExpressionIds"].([]interface{}); ok && len(rawIds) > 0 {
			if ruleId, ok := rawIds[0].(string); ok {
				policy.PreconfiguredExprId = ruleId
			}
		}
		row.EnforcedSecurityPolicy = policy
	}

	if previewPolicyMap, ok := jsonPayload["previewSecurityPolicy"].(map[string]interface{}); ok {
		// If it exists, initialize the PreviewSecurityPolicy struct
		row.PreviewSecurityPolicy = &RequestLogSecurityPolicy{}

		// Safely extract fields with type checking
		if v, ok := previewPolicyMap["configuredAction"].(string); ok {
			row.PreviewSecurityPolicy.ConfiguredAction = v
		}
		if v, ok := previewPolicyMap["name"].(string); ok {
			row.PreviewSecurityPolicy.Name = v
		}
		if v, ok := previewPolicyMap["outcome"].(string); ok {
			row.PreviewSecurityPolicy.Outcome = v
		}
		if v, ok := previewPolicyMap["priority"].(float64); ok {
			row.PreviewSecurityPolicy.Priority = int(v)
		}

		// Handle PreconfiguredExpressionIds within PreviewSecurityPolicy only if it exists and has values.
		if rawIds, ok := previewPolicyMap["preconfiguredExpressionIds"].([]interface{}); ok && len(rawIds) > 0 {
			if ruleId, ok := rawIds[0].(string); ok {
				row.PreviewSecurityPolicy.PreconfiguredExprId = ruleId
			}
		}
	}

	// Map SecurityPolicyRequestData if present
	if secPolicyData, ok := jsonPayload["securityPolicyRequestData"].(map[string]interface{}); ok {
		row.SecurityPolicyRequestData = &RequestLogSecurityPolicyRequestData{}

		if v, ok := secPolicyData["tlsJa3Fingerprint"].(string); ok {
			row.SecurityPolicyRequestData.TlsJa3Fingerprint = v
		}
		if v, ok := secPolicyData["tlsJa4Fingerprint"].(string); ok {
			row.SecurityPolicyRequestData.TlsJa4Fingerprint = v
		}

		if remoteIpInfo, ok := secPolicyData["remoteIpInfo"].(map[string]interface{}); ok {
			row.SecurityPolicyRequestData.RemoteIpInfo = &RequestLogRemoteIpInfo{}
			if v, ok := remoteIpInfo["asn"].(float64); ok {
				row.SecurityPolicyRequestData.RemoteIpInfo.Asn = int(v)
			}
			if v, ok := remoteIpInfo["regionCode"].(string); ok {
				row.SecurityPolicyRequestData.RemoteIpInfo.RegionCode = v
			}
		}

		// Handle recaptcha tokens
		if recaptchaActionToken, ok := secPolicyData["recaptchaActionToken"].(map[string]interface{}); ok {
			row.SecurityPolicyRequestData.RecaptchaActionToken = &RequestLogRecaptchaToken{}
			if v, ok := recaptchaActionToken["score"].(float64); ok {
				row.SecurityPolicyRequestData.RecaptchaActionToken.Score = v
			}
		}
		if recaptchaSessionToken, ok := secPolicyData["recaptchaSessionToken"].(map[string]interface{}); ok {
			row.SecurityPolicyRequestData.RecaptchaSessionToken = &RequestLogRecaptchaToken{}
			if v, ok := recaptchaSessionToken["score"].(float64); ok {
				row.SecurityPolicyRequestData.RecaptchaSessionToken.Score = v
			}
		}
	}

	// === 6. Map HTTPRequest ===
	if item.GetHttpRequest() == nil {
		row.HttpRequest = &RequestLogHttpRequest{}
	} else {
		httpRequestPb := item.GetHttpRequest()
		// Sanitize URLs to handle Unicode escape sequences that cause parsing errors
		requestUrl := sanitizeURL(httpRequestPb.GetRequestUrl())
		referer := sanitizeURL(httpRequestPb.GetReferer())
		row.HttpRequest = &RequestLogHttpRequest{
			RequestMethod:                  httpRequestPb.GetRequestMethod(),
			RequestUrl:                     requestUrl,
			RequestSize:                    httpRequestPb.GetRequestSize(),
			Referer:                        referer,
			UserAgent:                      httpRequestPb.GetUserAgent(),
			Status:                         httpRequestPb.GetStatus(),
			ResponseSize:                   httpRequestPb.GetResponseSize(),
			RemoteIp:                       httpRequestPb.GetRemoteIp(),
			Latency:                        httpRequestPb.GetLatency().String(), // Latency is a duration type in Protobuf
			ServerIp:                       httpRequestPb.GetServerIp(),
			Protocol:                       httpRequestPb.GetProtocol(),
			CacheFillBytes:                 httpRequestPb.GetCacheFillBytes(),
			CacheLookup:                    httpRequestPb.GetCacheLookup(),
			CacheHit:                       httpRequestPb.GetCacheHit(),
			CacheValidatedWithOriginServer: httpRequestPb.GetCacheValidatedWithOriginServer(),
		}
	}

	return row, nil
}

func mapFromBucketJson(itemBytes []byte) (*RequestsLog, error) {
	var log requestsLog
	if err := json.Unmarshal(itemBytes, &log); err != nil {
		return nil, fmt.Errorf("failed to parse requests log JSON: %w", err)
	}

	// Filter by log name - only process requests logs
	// Requests logs have log names like "projects/{project}/logs/requests"
	if !strings.Contains(log.LogName, "/logs/requests") {
		// Not a requests log, skip it
		return nil, nil
	}

	// Early exit if missing required fields
	if log.JsonPayload == nil {
		return nil, nil
	}

	row := NewRequestsLog()

	// Map top-level fields
	row.Timestamp = log.Timestamp
	row.ReceiveTimestamp = log.ReceiveTimestamp
	row.LogName = log.LogName
	row.InsertId = log.InsertId
	row.Severity = log.Severity
	row.Trace = log.Trace
	row.SpanId = log.SpanId
	row.TraceSampled = log.TraceSampled

	// Only create objects if they exist in the source log.
	// This avoids creating empty-but-non-nil objects that the downstream
	// validator might reject.

	if log.Resource != nil {
		row.Resource = &RequestLogResource{
			Type: log.Resource.Type,
			Labels: func() map[string]string {
				if log.Resource.Labels != nil {
					return log.Resource.Labels
				}
				return make(map[string]string)
			}(),
		}
	}

	// Map JSON Payload fields (JsonPayload is guaranteed to be non-nil here)
	row.BackendTargetProjectNumber = log.JsonPayload.BackendTargetProjectNumber
	row.CacheDecision = log.JsonPayload.CacheDecision
	row.RemoteIp = log.JsonPayload.RemoteIp
	row.StatusDetails = log.JsonPayload.StatusDetails

	if log.JsonPayload.EnforcedSecurityPolicy != nil {
		ids := []string{}
		if log.JsonPayload.EnforcedSecurityPolicy.PreconfiguredExprIds != nil {
			ids = log.JsonPayload.EnforcedSecurityPolicy.PreconfiguredExprIds
		}
		var exprId string
		if len(ids) > 0 {
			exprId = ids[0]
		}
		row.EnforcedSecurityPolicy = &RequestLogSecurityPolicy{
			ConfiguredAction:    log.JsonPayload.EnforcedSecurityPolicy.ConfiguredAction,
			Name:                log.JsonPayload.EnforcedSecurityPolicy.Name,
			Outcome:             log.JsonPayload.EnforcedSecurityPolicy.Outcome,
			Priority:            log.JsonPayload.EnforcedSecurityPolicy.Priority,
			MatchedFieldType:    log.JsonPayload.EnforcedSecurityPolicy.MatchedFieldType,
			MatchedFieldValue:   log.JsonPayload.EnforcedSecurityPolicy.MatchedFieldValue,
			MatchedFieldName:    log.JsonPayload.EnforcedSecurityPolicy.MatchedFieldName,
			MatchedOffset:       log.JsonPayload.EnforcedSecurityPolicy.MatchedOffset,
			MatchedLength:       log.JsonPayload.EnforcedSecurityPolicy.MatchedLength,
			PreconfiguredExprId: exprId,
		}
	}

	if log.JsonPayload.PreviewSecurityPolicy != nil {
		ids := []string{}
		if log.JsonPayload.PreviewSecurityPolicy.PreconfiguredExprIds != nil {
			ids = log.JsonPayload.PreviewSecurityPolicy.PreconfiguredExprIds
		}
		// preconfiguredExprIds is always an array of one string, grab the index 0 slice and case this to a string in the row struct
		var exprId string
		if len(ids) > 0 {
			exprId = ids[0]
		}
		row.PreviewSecurityPolicy = &RequestLogSecurityPolicy{
			ConfiguredAction:    log.JsonPayload.PreviewSecurityPolicy.ConfiguredAction,
			Name:                log.JsonPayload.PreviewSecurityPolicy.Name,
			Outcome:             log.JsonPayload.PreviewSecurityPolicy.Outcome,
			Priority:            log.JsonPayload.PreviewSecurityPolicy.Priority,
			MatchedFieldType:    log.JsonPayload.PreviewSecurityPolicy.MatchedFieldType,
			MatchedFieldValue:   log.JsonPayload.PreviewSecurityPolicy.MatchedFieldValue,
			MatchedFieldName:    log.JsonPayload.PreviewSecurityPolicy.MatchedFieldName,
			MatchedOffset:       log.JsonPayload.PreviewSecurityPolicy.MatchedOffset,
			MatchedLength:       log.JsonPayload.PreviewSecurityPolicy.MatchedLength,
			PreconfiguredExprId: exprId,
		}
	}

	if log.JsonPayload.SecurityPolicyRequestData != nil {
		row.SecurityPolicyRequestData = &RequestLogSecurityPolicyRequestData{
			TlsJa3Fingerprint: log.JsonPayload.SecurityPolicyRequestData.TlsJa3Fingerprint,
			TlsJa4Fingerprint: log.JsonPayload.SecurityPolicyRequestData.TlsJa4Fingerprint,
			// Always initialize the nested object to be safe.
			RemoteIpInfo: &RequestLogRemoteIpInfo{},
		}

		if log.JsonPayload.SecurityPolicyRequestData.RemoteIpInfo != nil {
			row.SecurityPolicyRequestData.RemoteIpInfo = &RequestLogRemoteIpInfo{
				Asn:        log.JsonPayload.SecurityPolicyRequestData.RemoteIpInfo.Asn,
				RegionCode: log.JsonPayload.SecurityPolicyRequestData.RemoteIpInfo.RegionCode,
			}
		}
	}

	// Sanitize URLs to handle Unicode escape sequences that cause parsing errors
	if log.HttpRequest == nil {
		row.HttpRequest = &RequestLogHttpRequest{}
	} else {
		requestUrl := sanitizeURL(log.HttpRequest.RequestURL)
		referer := sanitizeURL(log.HttpRequest.Referer)
		row.HttpRequest = &RequestLogHttpRequest{
			RequestMethod:                  log.HttpRequest.RequestMethod,
			RequestUrl:                     requestUrl,
			RequestSize:                    int64(log.HttpRequest.RequestSize),
			Status:                         log.HttpRequest.Status,
			ResponseSize:                   int64(log.HttpRequest.ResponseSize),
			UserAgent:                      log.HttpRequest.UserAgent,
			RemoteIp:                       log.HttpRequest.RemoteIP,
			ServerIp:                       log.HttpRequest.ServerIP,
			Referer:                        referer,
			Latency:                        log.HttpRequest.Latency,
			CacheLookup:                    log.HttpRequest.CacheLookup,
			CacheHit:                       log.HttpRequest.CacheHit,
			CacheValidatedWithOriginServer: log.HttpRequest.CacheValidatedWithOriginServer,
			CacheFillBytes:                 int64(log.HttpRequest.CacheFillBytes),
		}
	}

	return row, nil
}

type requestsLog struct {
	InsertId         string       `json:"insertId"`
	LogName          string       `json:"logName"`
	Resource         *resource    `json:"resource,omitempty"`
	Timestamp        time.Time    `json:"timestamp"`
	Severity         string       `json:"severity"`
	JsonPayload      *jsonPayload `json:"jsonPayload,omitempty"`
	ReceiveTimestamp time.Time    `json:"receiveTimestamp"`
	Trace            string       `json:"trace,omitempty"`
	SpanId           string       `json:"spanId,omitempty"`
	HttpRequest      *httpRequest `json:"httpRequest,omitempty"`
	TraceSampled     bool         `json:"traceSampled,omitempty"`
}

type resource struct {
	Type   string            `json:"type"`
	Labels map[string]string `json:"labels"`
}

type jsonPayload struct {
	TypeName                   string                               `json:"@type"`
	BackendTargetProjectNumber string                               `json:"backendTargetProjectNumber"`
	CacheDecision              []string                             `json:"cacheDecision"`
	CacheId                    string                               `json:"cacheId,omitempty"`
	CompressionStatus          string                               `json:"compressionStatus,omitempty"`
	EnforcedSecurityPolicy     *requestLogSecurityPolicy            `json:"enforcedSecurityPolicy"`
	PreviewSecurityPolicy      *requestLogSecurityPolicy            `json:"previewSecurityPolicy,omitempty"`
	SecurityPolicyRequestData  *requestLogSecurityPolicyRequestData `json:"securityPolicyRequestData"`
	RemoteIp                   string                               `json:"remoteIp"`
	StatusDetails              string                               `json:"statusDetails"`
}

type httpRequest struct {
	RequestMethod                  string        `json:"requestMethod"`
	RequestURL                     string        `json:"requestUrl"`
	RequestSize                    flexibleInt64 `json:"requestSize,omitempty"`
	Status                         int32         `json:"status"`
	ResponseSize                   flexibleInt64 `json:"responseSize,omitempty"`
	UserAgent                      string        `json:"userAgent"`
	RemoteIP                       string        `json:"remoteIp"`
	ServerIP                       string        `json:"serverIp,omitempty"`
	Referer                        string        `json:"referer,omitempty"`
	Latency                        string        `json:"latency,omitempty"`
	CacheLookup                    bool          `json:"cacheLookup,omitempty"`
	CacheHit                       bool          `json:"cacheHit,omitempty"`
	CacheValidatedWithOriginServer bool          `json:"cacheValidatedWithOriginServer,omitempty"`
	CacheFillBytes                 flexibleInt64 `json:"cacheFillBytes,omitempty"`
}

type requestLogSecurityPolicy struct {
	ConfiguredAction     string                        `json:"configuredAction"`
	Name                 string                        `json:"name"`
	Outcome              string                        `json:"outcome"`
	Priority             int                           `json:"priority"`
	PreconfiguredExprIds []string                      `json:"preconfiguredExprIds,omitempty"`
	RateLimitAction      *requestLogRateLimitAction    `json:"rateLimitAction,omitempty"`
	ThreatIntelligence   *requestLogThreatIntelligence `json:"threatIntelligence,omitempty"`
	AddressGroup         *requestLogAddressGroup       `json:"addressGroup,omitempty"`
	MatchedFieldType     string                        `json:"matchedFieldType,omitempty"`
	MatchedFieldValue    string                        `json:"matchedFieldValue,omitempty"`
	MatchedFieldName     string                        `json:"matchedFieldName,omitempty"`
	MatchedOffset        int                           `json:"matchedOffset,omitempty"`
	MatchedLength        int                           `json:"matchedLength,omitempty"`
}

type requestLogSecurityPolicyRequestData struct {
	RemoteIpInfo          *requestLogRemoteIpInfo   `json:"remoteIpInfo"`
	TlsJa3Fingerprint     string                    `json:"tlsJa3Fingerprint"`
	TlsJa4Fingerprint     string                    `json:"tlsJa4Fingerprint"`
	RecaptchaActionToken  *requestLogRecaptchaToken `json:"recaptchaActionToken,omitempty"`
	RecaptchaSessionToken *requestLogRecaptchaToken `json:"recaptchaSessionToken,omitempty"`
}

type requestLogRemoteIpInfo struct {
	Asn        int    `json:"asn"`
	RegionCode string `json:"regionCode"`
}

type requestLogRecaptchaToken struct {
	Score float64 `json:"score"`
}

type requestLogRateLimitAction struct {
	Key     string `json:"key"`
	Outcome string `json:"outcome"`
}

type requestLogThreatIntelligence struct {
	Categories []string `json:"categories"`
}

type requestLogAddressGroup struct {
	Names []string `json:"names"`
}
