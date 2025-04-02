//nolint:staticcheck
package audit_log

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/logging"
	"google.golang.org/genproto/googleapis/cloud/audit"
	adminpb "google.golang.org/genproto/googleapis/iam/admin/v1"
	loggingpb "google.golang.org/genproto/googleapis/iam/v1/logging"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/tailpipe-plugin-sdk/mappers"
)

type AuditLogMapper struct {
}

func (m *AuditLogMapper) Identifier() string {
	return "gcp_audit_log_mapper"
}

func (m *AuditLogMapper) Map(_ context.Context, a any, _ ...mappers.MapOption[*AuditLog]) (*AuditLog, error) {
	switch v := a.(type) {
	case string:
		return mapFromBucketJson([]byte(v))
	case logging.Entry:
		return mapFromSDKType(v)
	case []byte:
		return mapFromBucketJson(v)
	default:
		return nil, fmt.Errorf("expected logging.Entry, string or []byte, got %T", a)
	}

}

func decodeServiceData(tu string, v []byte) (*map[string]interface{}, error) {
	var protoMessage proto.Message

	switch tu {
	case "type.googleapis.com/google.iam.v1.logging.AuditData":
		protoMessage = &loggingpb.AuditData{}
	case "type.googleapis.com/google.iam.admin.v1.AuditData":
		protoMessage = &adminpb.AuditData{}
	default:
		return nil, fmt.Errorf("unsupported type: %s", tu)
	}

	// Unmarshal the protobuf payload into the appropriate struct
	if err := proto.Unmarshal(v, protoMessage); err != nil {
		return nil, fmt.Errorf("error decoding proto: %w", err)
	}

	// Marshal the protobuf message into JSON
	jsonBytes, err := protojson.Marshal(protoMessage)
	if err != nil {
		return nil, fmt.Errorf("error marshaling proto to JSON: %w", err)
	}

	// Unmarshal the JSON into a map[string]interface{}
	var result map[string]interface{}
	if err = json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON to map: %w", err)
	}

	return &result, nil
}

func mapFromSDKType(item logging.Entry) (*AuditLog, error) {
	row := NewAuditLog()
	row.Timestamp = item.Timestamp
	row.LogName = item.LogName
	row.InsertId = item.InsertID
	row.Severity = item.Severity.String()
	row.Trace = item.Trace
	row.TraceSampled = item.TraceSampled
	row.SpanId = item.SpanID

	// payload is special in this case as it's the core of the actual audit log, so it's properties are moved to top-level columns
	if payload, ok := item.Payload.(*audit.AuditLog); ok {
		row.ServiceName = &payload.ServiceName
		row.MethodName = &payload.MethodName
		row.ResourceName = &payload.ResourceName
		row.NumResponseItems = &payload.NumResponseItems

		if payload.Status != nil {
			row.Status = &AuditLogStatus{
				Code:    payload.Status.Code,
				Message: payload.Status.Message,
			}
		}

		if payload.AuthenticationInfo != nil {
			row.AuthenticationInfo = &AuditLogAuthenticationInfo{
				PrincipalEmail:        payload.AuthenticationInfo.PrincipalEmail,
				PrincipalSubject:      payload.AuthenticationInfo.PrincipalSubject,
				AuthoritySelector:     payload.AuthenticationInfo.AuthoritySelector,
				ServiceAccountKeyName: payload.AuthenticationInfo.ServiceAccountKeyName,
			}

			if payload.AuthenticationInfo.ThirdPartyPrincipal != nil {
				tpp := payload.AuthenticationInfo.ThirdPartyPrincipal.GetFields()
				row.AuthenticationInfo.ThirdPartyPrincipal = make(map[string]string, len(tpp))
				for k, v := range tpp {
					row.AuthenticationInfo.ThirdPartyPrincipal[k] = v.String()
				}
			}

			if payload.AuthenticationInfo.ServiceAccountDelegationInfo != nil {
				for _, v := range payload.AuthenticationInfo.ServiceAccountDelegationInfo {
					row.AuthenticationInfo.ServiceAccountDelegationInfo = append(row.AuthenticationInfo.ServiceAccountDelegationInfo, v.PrincipalSubject)
				}
			}
		}

		if payload.RequestMetadata != nil {
			var requestAttributes map[string]interface{}
			if payload.RequestMetadata.RequestAttributes != nil {
				jsonBytes, err := json.Marshal(payload.RequestMetadata.RequestAttributes)
				if err != nil {
					return nil, fmt.Errorf("error marshaling request attributes: %w", err)
				}
				err = json.Unmarshal(jsonBytes, &requestAttributes)
				if err != nil {
					return nil, fmt.Errorf("error unmarshaling request attributes: %w", err)
				}
			}

			row.RequestMetadata = &AuditLogRequestMetadata{
				CallerIp:                payload.RequestMetadata.CallerIp,
				CallerSuppliedUserAgent: payload.RequestMetadata.CallerSuppliedUserAgent,
				CallerNetwork:           payload.RequestMetadata.CallerNetwork,
				RequestAttributes:       &requestAttributes,
			}

			if payload.RequestMetadata.DestinationAttributes != nil {
				row.RequestMetadata.DestinationAttributes = &AuditLogRequestMetadataDestinationAttributes{
					Ip:         payload.RequestMetadata.DestinationAttributes.Ip,
					Port:       payload.RequestMetadata.DestinationAttributes.Port,
					Principal:  payload.RequestMetadata.DestinationAttributes.Principal,
					RegionCode: payload.RequestMetadata.DestinationAttributes.RegionCode,
					Labels:     payload.RequestMetadata.DestinationAttributes.Labels,
				}
			}
		}

		if payload.ResourceLocation != nil {
			row.ResourceLocation = &AuditLogResourceLocation{
				CurrentLocations:  payload.ResourceLocation.CurrentLocations,
				OriginalLocations: payload.ResourceLocation.OriginalLocations,
			}
		}

		if payload.PolicyViolationInfo != nil {
			row.PolicyViolationInfo = payload.PolicyViolationInfo
		}

		if payload.AuthorizationInfo != nil {
			for _, v := range payload.AuthorizationInfo {
				row.AuthorizationInfo = append(row.AuthorizationInfo, &AuditLogAuthorizationInfo{
					Resource:   v.Resource,
					Permission: v.Permission,
					Granted:    v.Granted,
				})
			}
		}

		if payload.ResourceOriginalState != nil {
			row.ResourceOriginalState = payload.ResourceOriginalState
		}

		if payload.Request != nil {
			row.Request = payload.Request.AsMap()
		}

		if payload.Response != nil {
			row.Response = payload.Response.AsMap()
		}

		if payload.Metadata != nil {
			row.Metadata = payload.Metadata.AsMap()
		}

		if payload.ServiceData != nil && payload.ServiceData.Value != nil {
			serviceData, err := decodeServiceData(payload.ServiceData.TypeUrl, payload.ServiceData.Value)
			if err != nil {
				return nil, fmt.Errorf("error decoding service data: %w", err)
			}
			row.ServiceData = serviceData
		}
	}

	// resource
	if item.Resource != nil {
		row.Resource = &AuditLogResource{
			Type:   item.Resource.Type,
			Labels: item.Resource.Labels,
		}
	}

	// operation
	if item.Operation != nil {
		row.Operation = &AuditLogOperation{
			Id:       item.Operation.Id,
			Producer: item.Operation.Producer,
			First:    item.Operation.First,
			Last:     item.Operation.Last,
		}
	}

	// http request
	if item.HTTPRequest != nil {
		row.HttpRequest = &AuditLogHttpRequest{
			Method:                         item.HTTPRequest.Request.Method,
			Url:                            item.HTTPRequest.Request.URL.String(),
			RequestHeaders:                 item.HTTPRequest.Request.Header,
			RequestSize:                    item.HTTPRequest.RequestSize,
			Status:                         item.HTTPRequest.Status,
			ResponseSize:                   item.HTTPRequest.ResponseSize,
			LocalIp:                        item.HTTPRequest.LocalIP,
			RemoteIp:                       item.HTTPRequest.RemoteIP,
			Latency:                        utils.HumanizeDuration(item.HTTPRequest.Latency),
			CacheHit:                       item.HTTPRequest.CacheHit,
			CacheLookup:                    item.HTTPRequest.CacheLookup,
			CacheValidatedWithOriginServer: item.HTTPRequest.CacheValidatedWithOriginServer,
			CacheFillBytes:                 item.HTTPRequest.CacheFillBytes,
		}
	}

	// labels
	if item.Labels != nil {
		row.Labels = &item.Labels
	}

	// source location
	if item.SourceLocation != nil {
		row.SourceLocation = &AuditLogSourceLocation{
			File:     item.SourceLocation.File,
			Line:     item.SourceLocation.Line,
			Function: item.SourceLocation.Function,
		}
	}

	return row, nil
}

func mapFromBucketJson(itemBytes []byte) (*AuditLog, error) {

	// make a struct for the json data
	var log auditLog
	err := json.Unmarshal(itemBytes, &log)
	if err != nil {
		return nil, fmt.Errorf("failed to parse audit log: %w", err)
	}

	// create a new row
	row := NewAuditLog()

	// set the fields
	row.InsertId = log.InsertID
	row.LogName = log.LogName
	row.Timestamp = log.Timestamp
	row.Severity = log.Severity
	row.Trace = log.Trace
	row.SpanId = log.SpanID
	row.TraceSampled = log.TraceSampled

	// proto payload
	if log.ProtoPayload != nil {
		row.ServiceName = &log.ProtoPayload.ServiceName
		row.MethodName = &log.ProtoPayload.MethodName
		row.ResourceName = &log.ProtoPayload.ResourceName
		if log.ProtoPayload.NumResponseItems != nil {
			switch v := log.ProtoPayload.NumResponseItems.(type) {
			case int64:
				row.NumResponseItems = &v
			case *int64:
				row.NumResponseItems = v
			case int:
				i := int64(v)
				row.NumResponseItems = &i
			case int16:
				i := int64(v)
				row.NumResponseItems = &i
			case int32:
				i := int64(v)
				row.NumResponseItems = &i
			case string:
				i, conErr := strconv.ParseInt(v, 10, 64)
				if conErr == nil {
					row.NumResponseItems = &i
				}
			}
		}

		if log.ProtoPayload.Status != nil {
			row.Status = &AuditLogStatus{
				Code:    int32(log.ProtoPayload.Status.Code),
				Message: log.ProtoPayload.Status.Message,
			}
		}

		if log.ProtoPayload.AuthenticationInfo != nil {
			row.AuthenticationInfo = &AuditLogAuthenticationInfo{
				PrincipalEmail:        log.ProtoPayload.AuthenticationInfo.PrincipalEmail,
				AuthoritySelector:     log.ProtoPayload.AuthenticationInfo.AuthoritySelector,
				ServiceAccountKeyName: log.ProtoPayload.AuthenticationInfo.ServiceAccountKeyName,
				PrincipalSubject:      log.ProtoPayload.AuthenticationInfo.PrincipalSubject,
			}

			if log.ProtoPayload.AuthenticationInfo.ThirdPartyPrincipal != nil {
				row.AuthenticationInfo.ThirdPartyPrincipal = make(map[string]string, len(log.ProtoPayload.AuthenticationInfo.ThirdPartyPrincipal))
				for k, v := range log.ProtoPayload.AuthenticationInfo.ThirdPartyPrincipal {
					row.AuthenticationInfo.ThirdPartyPrincipal[k] = v.(string)
				}
			}

			if log.ProtoPayload.AuthenticationInfo.ServiceAccountDelegationInfo != nil {
				for _, v := range log.ProtoPayload.AuthenticationInfo.ServiceAccountDelegationInfo {
					row.AuthenticationInfo.ServiceAccountDelegationInfo = append(row.AuthenticationInfo.ServiceAccountDelegationInfo, v.FirstPartyPrincipal.PrincipalEmail)
				}
			}
		}

		if log.ProtoPayload.RequestMetadata != nil {
			row.RequestMetadata = &AuditLogRequestMetadata{
				CallerIp:                log.ProtoPayload.RequestMetadata.CallerIP,
				CallerSuppliedUserAgent: log.ProtoPayload.RequestMetadata.CallerSuppliedUserAgent,
				CallerNetwork:           log.ProtoPayload.RequestMetadata.CallerNetwork,
				RequestAttributes:       &log.ProtoPayload.RequestMetadata.RequestAttributes,
			}

			if log.ProtoPayload.RequestMetadata.DestinationAttributes != nil {
				row.RequestMetadata.DestinationAttributes = &AuditLogRequestMetadataDestinationAttributes{
					Ip:         log.ProtoPayload.RequestMetadata.DestinationAttributes.Ip,
					Port:       int64(log.ProtoPayload.RequestMetadata.DestinationAttributes.Port),
					Principal:  log.ProtoPayload.RequestMetadata.DestinationAttributes.Principal,
					RegionCode: log.ProtoPayload.RequestMetadata.DestinationAttributes.RegionCode,
					Labels:     log.ProtoPayload.RequestMetadata.DestinationAttributes.Labels,
				}
			}
		}

		if log.ProtoPayload.ResourceLocation != nil {
			row.ResourceLocation = &AuditLogResourceLocation{
				CurrentLocations:  log.ProtoPayload.ResourceLocation.CurrentLocations,
				OriginalLocations: log.ProtoPayload.ResourceLocation.OriginalLocations,
			}
		}

		if log.ProtoPayload.AuthorizationInfo != nil {
			for _, v := range log.ProtoPayload.AuthorizationInfo {
				row.AuthorizationInfo = append(row.AuthorizationInfo, &AuditLogAuthorizationInfo{
					Resource:   v.Resource,
					Permission: v.Permission,
					Granted:    v.Granted,
				})
			}
		}

		if log.ProtoPayload.ResourceOriginalState != nil {
			row.ResourceOriginalState = log.ProtoPayload.ResourceOriginalState
		}

		if log.ProtoPayload.Request != nil {
			row.Request = log.ProtoPayload.Request
		}

		if log.ProtoPayload.Response != nil {
			row.Response = log.ProtoPayload.Response
		}

		if log.ProtoPayload.Metadata != nil {
			row.Metadata = log.ProtoPayload.Metadata
		}

		if log.ProtoPayload.ServiceData != nil {
			row.ServiceData = &log.ProtoPayload.ServiceData
		}
	}

	// resource
	if log.Resource != nil {
		row.Resource = &AuditLogResource{
			Type:   log.Resource.Type,
			Labels: log.Resource.Labels,
		}
	}

	// operation
	if log.Operation != nil {
		row.Operation = &AuditLogOperation{
			Id:       log.Operation.ID,
			Producer: log.Operation.Producer,
			First:    log.Operation.First,
			Last:     log.Operation.Last,
		}
	}

	// http request
	if log.HTTPRequest != nil {
		row.HttpRequest = &AuditLogHttpRequest{
			Method:                         log.HTTPRequest.RequestMethod,
			Url:                            log.HTTPRequest.RequestURL,
			Status:                         log.HTTPRequest.Status,
			UserAgent:                      &log.HTTPRequest.UserAgent,
			RemoteIp:                       log.HTTPRequest.RemoteIP,
			Latency:                        log.HTTPRequest.Latency,
			CacheHit:                       log.HTTPRequest.CacheHit,
			CacheLookup:                    log.HTTPRequest.CacheLookup,
			CacheValidatedWithOriginServer: log.HTTPRequest.CacheValidatedWithOriginServer,
			CacheFillBytes:                 log.HTTPRequest.CacheFillBytes,
		}
	}

	// labels
	if log.Labels != nil {
		row.Labels = &log.Labels
	}

	// source location
	if log.SourceLocation != nil {
		row.SourceLocation = &AuditLogSourceLocation{
			File:     log.SourceLocation.File,
			Line:     int64(log.SourceLocation.Line),
			Function: log.SourceLocation.Function,
		}
	}

	return row, nil
}

type auditLog struct {
	InsertID         string            `json:"insertId"`
	LogName          string            `json:"logName"`
	Resource         *resource         `json:"resource,omitempty"`
	Timestamp        time.Time         `json:"timestamp"`
	Severity         string            `json:"severity"`
	ProtoPayload     *protoPayload     `json:"protoPayload,omitempty"`
	ReceiveTimestamp time.Time         `json:"receiveTimestamp"`
	Operation        *operation        `json:"operation,omitempty"`
	Trace            string            `json:"trace,omitempty"`
	SpanID           string            `json:"spanId,omitempty"`
	TraceSampled     bool              `json:"traceSampled,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	SourceLocation   *sourceLocation   `json:"sourceLocation,omitempty"`
	HTTPRequest      *httpRequest      `json:"httpRequest,omitempty"`
}

type resource struct {
	Type   string            `json:"type"`
	Labels map[string]string `json:"labels"`
}

type operation struct {
	ID       string `json:"id"`
	Producer string `json:"producer"`
	First    bool   `json:"first"`
	Last     bool   `json:"last"`
}

type sourceLocation struct {
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Function string `json:"function,omitempty"`
}

type protoPayload struct {
	TypeName              string              `json:"@type"`
	MethodName            string              `json:"methodName"`
	AuthenticationInfo    *authenticationInfo `json:"authenticationInfo,omitempty"`
	RequestMetadata       *requestMetadata    `json:"requestMetadata,omitempty"`
	ServiceName           string              `json:"serviceName"`
	ResourceName          string              `json:"resourceName"`
	AuthorizationInfo     []authorizationInfo `json:"authorizationInfo,omitempty"`
	HTTPRequest           *httpRequest        `json:"httpRequest,omitempty"`
	Status                *status             `json:"status,omitempty"`
	Response              map[string]any      `json:"response,omitempty"`
	Request               map[string]any      `json:"request,omitempty"`
	Metadata              map[string]any      `json:"metadata,omitempty"`
	ServiceData           map[string]any      `json:"serviceData,omitempty"`
	ResourceLocation      *resourceLocation   `json:"resourceLocation,omitempty"`
	NumResponseItems      any                 `json:"numResponseItems,omitempty"`
	ResourceOriginalState map[string]any      `json:"resourceOriginalState,omitempty"`
}

type authenticationInfo struct {
	PrincipalEmail               string           `json:"principalEmail"`
	PrincipalSubject             string           `json:"principalSubject"`
	AuthoritySelector            string           `json:"authoritySelector,omitempty"`
	ThirdPartyPrincipal          map[string]any   `json:"thirdPartyPrincipal,omitempty"`
	ServiceAccountKeyName        string           `json:"serviceAccountKeyName,omitempty"`
	ServiceAccountDelegationInfo []delegationInfo `json:"serviceAccountDelegationInfo,omitempty"`
}

type authorizationInfo struct {
	Resource           string         `json:"resource"`
	Permission         string         `json:"permission"`
	Granted            bool           `json:"granted"`
	ResourceAttributes map[string]any `json:"resourceAttributes,omitempty"`
}

type requestMetadata struct {
	CallerIP                string                 `json:"callerIp"`
	CallerSuppliedUserAgent string                 `json:"callerSuppliedUserAgent"`
	CallerNetwork           string                 `json:"callerNetwork"`
	DestinationAttributes   *destinationAttributes `json:"destinationAttributes,omitempty"`
	RequestAttributes       map[string]any         `json:"requestAttributes,omitempty"`
}

type httpRequest struct {
	RequestMethod                  string `json:"requestMethod"`
	RequestURL                     string `json:"requestUrl"`
	RequestSize                    string `json:"requestSize,omitempty"`
	Status                         int    `json:"status"`
	ResponseSize                   string `json:"responseSize,omitempty"`
	UserAgent                      string `json:"userAgent"`
	RemoteIP                       string `json:"remoteIp"`
	ServerIP                       string `json:"serverIp,omitempty"`
	Referer                        string `json:"referer,omitempty"`
	Latency                        string `json:"latency,omitempty"`
	CacheLookup                    bool   `json:"cacheLookup,omitempty"`
	CacheHit                       bool   `json:"cacheHit,omitempty"`
	CacheValidatedWithOriginServer bool   `json:"cacheValidatedWithOriginServer,omitempty"`
	CacheFillBytes                 int64  `json:"cacheFillBytes,omitempty"`
	Protocol                       string `json:"protocol,omitempty"`
}

type status struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details []any  `json:"details,omitempty"`
}

type resourceLocation struct {
	CurrentLocations  []string `json:"currentLocations,omitempty"`
	OriginalLocations []string `json:"originalLocations,omitempty"`
}

type delegationInfo struct {
	FirstPartyPrincipal firstPartyPrincipal `json:"firstPartyPrincipal,omitempty"`
	ThirdPartyPrincipal map[string]any      `json:"thirdPartyPrincipal,omitempty"`
}

type firstPartyPrincipal struct {
	PrincipalEmail  string         `json:"principalEmail,omitempty"`
	ServiceMetadata map[string]any `json:"serviceMetadata,omitempty"`
}

type destinationAttributes struct {
	Ip         string            `json:"ip,omitempty"`
	Port       int               `json:"port,omitempty"`
	Principal  string            `json:"principal,omitempty"`
	RegionCode string            `json:"regionCode,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}
