package mappers

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/logging"
	"google.golang.org/genproto/googleapis/cloud/audit"

	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-sdk/table"
)

type AuditLogMapper struct {
}

func NewAuditLogMapper() table.Mapper[*rows.AuditLog] {
	return &AuditLogMapper{}
}

func (m *AuditLogMapper) Identifier() string {
	return "gcp_audit_log_mapper"
}

func (m *AuditLogMapper) Map(_ context.Context, a any) (*rows.AuditLog, error) {
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

	if payload, ok := item.Payload.(*audit.AuditLog); ok {
		row.ServiceName = &payload.ServiceName
		row.MethodName = &payload.MethodName
		row.ResourceName = &payload.ResourceName

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
		}

		if payload.RequestMetadata != nil {
			row.RequestMetadata = &rows.AuditLogRequestMetadata{
				CallerIp:                payload.RequestMetadata.CallerIp,
				CallerSuppliedUserAgent: payload.RequestMetadata.CallerSuppliedUserAgent,
			}
		}
	}

	if item.Resource != nil {
		row.Resource = &rows.AuditLogResource{
			Type:   item.Resource.Type,
			Labels: item.Resource.Labels,
		}
	}

	if item.Operation != nil {
		row.Operation = &rows.AuditLogOperation{
			Id:       item.Operation.Id,
			Producer: item.Operation.Producer,
			First:    item.Operation.First,
			Last:     item.Operation.Last,
		}
	}

	if item.HTTPRequest != nil {
		row.HttpRequest = &rows.AuditLogHttpRequest{
			Method:       item.HTTPRequest.Request.Method,
			Url:          item.HTTPRequest.Request.URL.String(),
			Size:         item.HTTPRequest.RequestSize,
			Status:       item.HTTPRequest.Status,
			ResponseSize: item.HTTPRequest.ResponseSize,
			LocalIp:      item.HTTPRequest.LocalIP,
			RemoteIp:     item.HTTPRequest.RemoteIP,
		}
	}

	return row, nil
}
