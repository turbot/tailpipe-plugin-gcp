package mappers

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/logging"
	"google.golang.org/genproto/googleapis/cloud/audit"
	adminpb "google.golang.org/genproto/googleapis/iam/admin/v1"
	loggingpb "google.golang.org/genproto/googleapis/iam/v1/logging"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-sdk/table"
)

type AuditLogMapper struct {
}

func (m *AuditLogMapper) Identifier() string {
	return "gcp_audit_log_mapper"
}

func (m *AuditLogMapper) Map(_ context.Context, a any, _ ...table.MapOption[*rows.AuditLog]) (*rows.AuditLog, error) {
	var item logging.Entry

	switch v := a.(type) {
	case string:
		err := json.Unmarshal([]byte(v), &item)
		if err != nil {
			return nil, err
		}
	case logging.Entry:
		item = v
	case []byte:
		err := json.Unmarshal(v, &item)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("expected logging.Entry, string or []byte, got %T", a)
	}

	row := rows.NewAuditLog()
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
			row.Status = &rows.AuditLogStatus{
				Code:    payload.Status.Code,
				Message: payload.Status.Message,
			}
		}

		if payload.AuthenticationInfo != nil {
			row.AuthenticationInfo = &rows.AuditLogAuthenticationInfo{
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
			row.RequestMetadata = &rows.AuditLogRequestMetadata{
				CallerIp:                payload.RequestMetadata.CallerIp,
				CallerSuppliedUserAgent: payload.RequestMetadata.CallerSuppliedUserAgent,
				CallerNetwork:           payload.RequestMetadata.CallerNetwork,
				RequestAttributes:       payload.RequestMetadata.RequestAttributes,
			}

			if payload.RequestMetadata.DestinationAttributes != nil {
				row.RequestMetadata.DestinationAttributes = &rows.AuditLogRequestMetadataDestinationAttributes{
					Ip:         payload.RequestMetadata.DestinationAttributes.Ip,
					Port:       payload.RequestMetadata.DestinationAttributes.Port,
					Principal:  payload.RequestMetadata.DestinationAttributes.Principal,
					RegionCode: payload.RequestMetadata.DestinationAttributes.RegionCode,
					Labels:     payload.RequestMetadata.DestinationAttributes.Labels,
				}
			}
		}

		if payload.ResourceLocation != nil {
			row.ResourceLocation = &rows.AuditLogResourceLocation{
				CurrentLocations:  payload.ResourceLocation.CurrentLocations,
				OriginalLocations: payload.ResourceLocation.OriginalLocations,
			}
		}

		if payload.PolicyViolationInfo != nil {
			row.PolicyViolationInfo = payload.PolicyViolationInfo
		}

		if payload.AuthorizationInfo != nil {
			for _, v := range payload.AuthorizationInfo {
				row.AuthorizationInfo = append(row.AuthorizationInfo, v)
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
		row.Resource = &rows.AuditLogResource{
			Type:   item.Resource.Type,
			Labels: item.Resource.Labels,
		}
	}

	// operation
	if item.Operation != nil {
		row.Operation = &rows.AuditLogOperation{
			Id:       item.Operation.Id,
			Producer: item.Operation.Producer,
			First:    item.Operation.First,
			Last:     item.Operation.Last,
		}
	}

	// http request
	if item.HTTPRequest != nil {
		row.HttpRequest = &rows.AuditLogHttpRequest{
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
		row.SourceLocation = &rows.AuditLogSourceLocation{
			File:     item.SourceLocation.File,
			Line:     item.SourceLocation.Line,
			Function: item.SourceLocation.Function,
		}
	}

	return row, nil
}

//func decodeServiceData(tu string, v []byte) (string, error) {
//	switch tu {
//	case "type.googleapis.com/google.iam.v1.logging.AuditData":
//		var auditData loggingpb.AuditData
//		if err := proto.Unmarshal(v, &auditData); err != nil {
//			return "", fmt.Errorf("error decoding proto: %w", err)
//		}
//		return auditData.String(), nil
//	case "type.googleapis.com/google.iam.admin.v1.AuditData":
//		var auditData adminpb.AuditData
//		if err := proto.Unmarshal(v, &auditData); err != nil {
//			return "", fmt.Errorf("error decoding proto: %w", err)
//		}
//		return auditData.String(), nil
//	default:
//		return "", nil
//	}
//}

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
