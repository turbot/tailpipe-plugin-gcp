package mappers

import (
	"cloud.google.com/go/logging"
	"context"
	"fmt"
	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-sdk/table"
	"google.golang.org/genproto/googleapis/cloud/audit"
)

type AuditLogMapper struct {
}

func NewAuditLogMapper() table.Mapper[*rows.AuditLog] {
	return &AuditLogMapper{}
}

func (m *AuditLogMapper) Identifier() string {
	return "gcp_audit_log_mapper"
}

func (m *AuditLogMapper) Map(_ context.Context, a any) ([]*rows.AuditLog, error) {
	item, ok := a.(logging.Entry)
	if !ok {
		return nil, fmt.Errorf("expected logging.Entry, got %T", a)
	}

	payload, ok := item.Payload.(*audit.AuditLog)
	if !ok {
		return nil, fmt.Errorf("invalid payload type: %T, expected *audit.AuditLog", item.Payload)
	}

	row := rows.NewAuditLog()

	row.Timestamp = item.Timestamp
	row.LogName = item.LogName
	row.InsertId = item.InsertID
	row.Severity = item.Severity.String()
	row.ServiceName = payload.ServiceName
	row.MethodName = payload.MethodName
	row.ResourceName = payload.ResourceName

	if payload.Status != nil {
		row.StatusCode = &payload.Status.Code
		row.StatusMessage = &payload.Status.Message
	}

	if item.Resource != nil {
		row.ResourceType = &item.Resource.Type
		//row.ResourceLabels = &item.Resource.Labels // TODO: #finish add back in once we have support for map
	}

	if item.Operation != nil {
		row.OperationId = &item.Operation.Id
		row.OperationProducer = &item.Operation.Producer
		row.OperationFirst = &item.Operation.First
		row.OperationLast = &item.Operation.Last
	}

	if payload.AuthenticationInfo != nil {
		row.AuthenticationPrincipal = &payload.AuthenticationInfo.PrincipalEmail
	}

	if payload.RequestMetadata != nil {
		row.RequestCallerIp = &payload.RequestMetadata.CallerIp
		row.RequestCallerSuppliedUserAgent = &payload.RequestMetadata.CallerSuppliedUserAgent
	}

	return []*rows.AuditLog{row}, nil
}
