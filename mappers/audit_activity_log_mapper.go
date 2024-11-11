package mappers

import (
	"context"
	"encoding/json"
	"fmt"
	// "log"

	"cloud.google.com/go/logging"
	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-sdk/helpers"
	"github.com/turbot/tailpipe-plugin-sdk/table"
	// "google.golang.org/genproto/googleapis/cloud/audit"
)

type AuditActivityLogMapper struct {
}

func NewAuditActivityLogMapper() table.Mapper[*rows.AuditActivityLog] {
	return &AuditActivityLogMapper{}
}

func (m *AuditActivityLogMapper) Identifier() string {
	return "gcp_audit_activity_log_mapper"
}

func (m *AuditActivityLogMapper) Map(_ context.Context, a any) ([]*rows.AuditActivityLog, error) {
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

	row := &rows.AuditActivityLog{}
	row.Timestamp = item.Timestamp
	row.LogName = item.LogName
	row.InsertId = item.InsertID
	row.Severity = item.Severity.String()

	payloadBytes, err := json.Marshal(item.Payload)
	if err != nil {
		return nil, err
	}
	tempPayload := helpers.JSONString(payloadBytes)
	row.ProtoPayload = &tempPayload
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

	if item.HTTPRequest != nil {
		row.RequestMethod = item.HTTPRequest.Request.Method
		row.RequestSize = item.HTTPRequest.RequestSize
		row.RequestStatus = item.HTTPRequest.Status
		row.RequestResponseSize = item.HTTPRequest.ResponseSize
	}

	return []*rows.AuditActivityLog{row}, nil
}