//nolint:staticcheck
package requests_log

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	// for debugging
	"os"

	loggingpb "cloud.google.com/go/logging/apiv2/loggingpb"

	"github.com/turbot/tailpipe-plugin-sdk/mappers"
)

func dumpRow(row *RequestsLog) {
	f, err := os.Create("/tmp/row_debug.json")
	if err == nil {
		defer f.Close()
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		enc.Encode(row)
	}
}

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
	case []byte:
		return mapFromBucketJson(v)
	default:
		return nil, fmt.Errorf("expected loggingpb.LogEntry, string or []byte, got %T", a)
	}
}

func mapFromSDKType(item *loggingpb.LogEntry) (*RequestsLog, error) {
	// === 1. Early exit for non-HTTP(S) logs or those missing a payload ===
	if item.GetHttpRequest() == nil || item.GetJsonPayload() == nil {
		return nil, nil
	}

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
	row.BackendTargetProjectNumber = jsonPayload["backendTargetProjectNumber"].(string)
	// Handle CacheDecision specifically:
	if rawCacheDecision, ok := jsonPayload["cacheDecision"].([]interface{}); ok {
		// Iterate over the []interface{} and append string elements to row.CacheDecision
		for _, v := range rawCacheDecision {
			if s, ok := v.(string); ok {
				row.CacheDecision = append(row.CacheDecision, s)
			}
		}
	}
	row.RemoteIp = jsonPayload["remoteIp"].(string)
	row.StatusDetails = jsonPayload["statusDetails"].(string)

	securityPolicyMap := jsonPayload["enforcedSecurityPolicy"].(map[string]interface{})

	row.EnforcedSecurityPolicy = &RequestLogSecurityPolicy{
		// Direct assignments for guaranteed scalar fields
		ConfiguredAction: securityPolicyMap["configuredAction"].(string),
		Name:             securityPolicyMap["name"].(string),
		Outcome:          securityPolicyMap["outcome"].(string),
		Priority:         int(securityPolicyMap["priority"].(float64)), // JSON numbers are float64
	}

	// Handle PreconfiguredExpressionIds only if it exists *and* has values.
	if rawIds, ok := securityPolicyMap["preconfiguredExpressionIds"].([]interface{}); ok && len(rawIds) > 0 {
		row.EnforcedSecurityPolicy.PreconfiguredExpressionIds = make([]string, 0, len(rawIds))
		for _, id := range rawIds {
			row.EnforcedSecurityPolicy.PreconfiguredExpressionIds = append(row.EnforcedSecurityPolicy.PreconfiguredExpressionIds, id.(string))
		}
	}

	if previewPolicyMap, ok := jsonPayload["previewSecurityPolicy"].(map[string]interface{}); ok {
		// If it exists, initialize the PreviewSecurityPolicy struct
		row.PreviewSecurityPolicy = &RequestLogSecurityPolicy{
			// Direct assignments for its guaranteed scalar fields within previewPolicyMap
			ConfiguredAction: previewPolicyMap["configuredAction"].(string),
			Name:             previewPolicyMap["name"].(string),
			Outcome:          previewPolicyMap["outcome"].(string),
			Priority:         int(previewPolicyMap["priority"].(float64)), // JSON numbers are float64
		}

		// Handle PreconfiguredExpressionIds within PreviewSecurityPolicy only if it exists and has values.
		if rawIds, ok := previewPolicyMap["preconfiguredExpressionIds"].([]interface{}); ok && len(rawIds) > 0 {
			row.PreviewSecurityPolicy.PreconfiguredExpressionIds = make([]string, 0, len(rawIds))
			for _, id := range rawIds {
				row.PreviewSecurityPolicy.PreconfiguredExpressionIds = append(row.PreviewSecurityPolicy.PreconfiguredExpressionIds, id.(string))
			}
		}
	}

	// === 5. Map HTTPRequest (guaranteed to be present due to early exit) ===
	// No 'if' check needed here for item.GetHttpRequest() because we already filtered.
	httpRequestPb := item.GetHttpRequest()
	row.HttpRequest = &RequestLogHttpRequest{
		RequestMethod:                  httpRequestPb.GetRequestMethod(),
		RequestUrl:                     httpRequestPb.GetRequestUrl(),
		RequestSize:                    strconv.FormatInt(httpRequestPb.GetRequestSize(), 10),
		Referer:                        httpRequestPb.GetReferer(),
		UserAgent:                      httpRequestPb.GetUserAgent(),
		Status:                         httpRequestPb.GetStatus(),
		ResponseSize:                   strconv.FormatInt(httpRequestPb.GetResponseSize(), 10),
		RemoteIp:                       httpRequestPb.GetRemoteIp(),
		Latency:                        httpRequestPb.GetLatency().String(),
		ServerIp:                       httpRequestPb.GetServerIp(),
		Protocol:                       httpRequestPb.GetProtocol(),
		CacheFillBytes:                 strconv.FormatInt(httpRequestPb.GetCacheFillBytes(), 10),
		CacheLookup:                    httpRequestPb.GetCacheLookup(),
		CacheHit:                       httpRequestPb.GetCacheHit(),
		CacheValidatedWithOriginServer: httpRequestPb.GetCacheValidatedWithOriginServer(),
	}

	return row, nil
}

func mapFromBucketJson(itemBytes []byte) (*RequestsLog, error) {
	var log requestsLog
	if err := json.Unmarshal(itemBytes, &log); err != nil {
		return nil, fmt.Errorf("failed to parse requests log JSON: %w", err)
	}

	// Filter out log entries that are not HTTP requests.
	if log.HttpRequest == nil || log.JsonPayload == nil {
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

	// FIX: Only create objects if they exist in the source log.
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
		if log.JsonPayload.EnforcedSecurityPolicy.PreconfiguredExpressionIds != nil {
			ids = log.JsonPayload.EnforcedSecurityPolicy.PreconfiguredExpressionIds
		}
		row.EnforcedSecurityPolicy = &RequestLogSecurityPolicy{
			ConfiguredAction:           log.JsonPayload.EnforcedSecurityPolicy.ConfiguredAction,
			Name:                       log.JsonPayload.EnforcedSecurityPolicy.Name,
			Outcome:                    log.JsonPayload.EnforcedSecurityPolicy.Outcome,
			Priority:                   log.JsonPayload.EnforcedSecurityPolicy.Priority,
			PreconfiguredExpressionIds: ids,
		}
	}

	if log.JsonPayload.PreviewSecurityPolicy != nil {
		ids := []string{}
		if log.JsonPayload.PreviewSecurityPolicy.PreconfiguredExpressionIds != nil {
			ids = log.JsonPayload.PreviewSecurityPolicy.PreconfiguredExpressionIds
		}
		row.PreviewSecurityPolicy = &RequestLogSecurityPolicy{
			ConfiguredAction:           log.JsonPayload.PreviewSecurityPolicy.ConfiguredAction,
			Name:                       log.JsonPayload.PreviewSecurityPolicy.Name,
			Outcome:                    log.JsonPayload.PreviewSecurityPolicy.Outcome,
			Priority:                   log.JsonPayload.PreviewSecurityPolicy.Priority,
			PreconfiguredExpressionIds: ids,
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

	// HttpRequest is guaranteed non-nil by the filter at the top of the function.
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

	// dumpRow(row)

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
	EnforcedSecurityPolicy     *requestLogEnforcedSecurityPolicy    `json:"enforcedSecurityPolicy"`
	PreviewSecurityPolicy      *requestLogPreviewSecurityPolicy     `json:"previewSecurityPolicy,omitempty"`
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

type requestLogEnforcedSecurityPolicy struct {
	ConfiguredAction           string   `json:"configuredAction"`
	Name                       string   `json:"name"`
	Outcome                    string   `json:"outcome"`
	Priority                   int      `json:"priority"`
	PreconfiguredExpressionIds []string `json:"preconfiguredExpressionIds,omitempty"`
}

type requestLogPreviewSecurityPolicy struct {
	ConfiguredAction           string   `json:"configuredAction"`
	Name                       string   `json:"name"`
	Outcome                    string   `json:"outcome"`
	Priority                   int      `json:"priority"`
	PreconfiguredExpressionIds []string `json:"preconfiguredExpressionIds,omitempty"`
}

type requestLogSecurityPolicyRequestData struct {
	RemoteIpInfo      *requestLogRemoteIpInfo `json:"remoteIpInfo"`
	TlsJa3Fingerprint string                  `json:"tlsJa3Fingerprint"`
	TlsJa4Fingerprint string                  `json:"tlsJa4Fingerprint"`
}

type requestLogRemoteIpInfo struct {
	Asn        int    `json:"asn"`
	RegionCode string `json:"regionCode"`
}
