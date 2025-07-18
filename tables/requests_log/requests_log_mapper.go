//nolint:staticcheck
package requests_log

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	// for debugging

	"cloud.google.com/go/logging"
	loggingpb "cloud.google.com/go/logging/apiv2/loggingpb"

	"github.com/turbot/tailpipe-plugin-sdk/mappers"
)

type RequestsLogMapper struct {
}

func (m *RequestsLogMapper) Identifier() string {
	return "gcp_requests_log_mapper"
}

func (m *RequestsLogMapper) Map(_ context.Context, a any, _ ...mappers.MapOption[*RequestsLog]) (*RequestsLog, error) {
	switch v := a.(type) {
	case string:
		return mapFromBucketJson([]byte(v))
	case *loggingpb.LogEntry:
		return mapFromSDKType(v)
	case *logging.Entry:
		return nil, fmt.Errorf("logging.Entry did not convert to *loggingpb.LogEntry: %T", a)
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
	row.Trace = item.GetTrace()
	row.SpanId = item.GetSpanId()
	row.TraceSampled = item.GetTraceSampled()

	// === 3. Map Resource ===
	if item.GetResource() != nil {
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
	}

	if item.GetHttpRequest() == nil {
		row.HttpRequest = &RequestLogHttpRequest{}
	} else {
		httpRequestPb := item.GetHttpRequest()
		row.HttpRequest = &RequestLogHttpRequest{
			RequestMethod: httpRequestPb.GetRequestMethod(),
			RequestUrl:    httpRequestPb.GetRequestUrl(),
			RequestSize:   strconv.FormatInt(httpRequestPb.GetRequestSize(), 10),
			Referer:       httpRequestPb.GetReferer(),
			UserAgent:     httpRequestPb.GetUserAgent(),
			Status:        httpRequestPb.GetStatus(),
			ResponseSize:  strconv.FormatInt(httpRequestPb.GetResponseSize(), 10),
			RemoteIp:      httpRequestPb.GetRemoteIp(),
			Latency: func() string {
				if lat := httpRequestPb.GetLatency(); lat != nil {
					return lat.String()
				}
				return ""
			}(),
			ServerIp:                       httpRequestPb.GetServerIp(),
			Protocol:                       httpRequestPb.GetProtocol(),
			CacheFillBytes:                 strconv.FormatInt(httpRequestPb.GetCacheFillBytes(), 10),
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

	// Map JSON Payload fields
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

	if log.HttpRequest == nil {
		row.HttpRequest = &RequestLogHttpRequest{}
	} else {
		row.HttpRequest = &RequestLogHttpRequest{
			RequestMethod:                  log.HttpRequest.RequestMethod,
			RequestUrl:                     log.HttpRequest.RequestURL,
			RequestSize:                    log.HttpRequest.RequestSize,
			Status:                         log.HttpRequest.Status,
			ResponseSize:                   log.HttpRequest.ResponseSize,
			UserAgent:                      log.HttpRequest.UserAgent,
			RemoteIp:                       log.HttpRequest.RemoteIP,
			ServerIp:                       log.HttpRequest.ServerIP,
			Referer:                        log.HttpRequest.Referer,
			Latency:                        log.HttpRequest.Latency,
			CacheLookup:                    log.HttpRequest.CacheLookup,
			CacheHit:                       log.HttpRequest.CacheHit,
			CacheValidatedWithOriginServer: log.HttpRequest.CacheValidatedWithOriginServer,
			CacheFillBytes:                 log.HttpRequest.CacheFillBytes,
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
	RequestMethod                  string `json:"requestMethod"`
	RequestURL                     string `json:"requestUrl"`
	RequestSize                    string `json:"requestSize,omitempty"`
	Status                         int32  `json:"status"`
	ResponseSize                   string `json:"responseSize,omitempty"`
	UserAgent                      string `json:"userAgent"`
	RemoteIP                       string `json:"remoteIp"`
	ServerIP                       string `json:"serverIp,omitempty"`
	Referer                        string `json:"referer,omitempty"`
	Latency                        string `json:"latency,omitempty"`
	CacheLookup                    bool   `json:"cacheLookup,omitempty"`
	CacheHit                       bool   `json:"cacheHit,omitempty"`
	CacheValidatedWithOriginServer bool   `json:"cacheValidatedWithOriginServer,omitempty"`
	CacheFillBytes                 string `json:"cacheFillBytes,omitempty"`
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

// (No replacement lines; the block is removed entirely.)

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
