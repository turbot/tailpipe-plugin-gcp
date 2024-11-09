package mappers

import (
    "context"
    "encoding/json"
    "fmt"

    "cloud.google.com/go/logging"
    "github.com/turbot/tailpipe-plugin-gcp/rows"
    "github.com/turbot/tailpipe-plugin-sdk/table"
    "google.golang.org/genproto/googleapis/cloud/audit"
)

type StorageLogMapper struct {
}

func NewStorageLogMapper() table.Mapper[*rows.StorageLog] {
    return &StorageLogMapper{}
}

func (m *StorageLogMapper) Identifier() string {
    return "gcp_storage_log_mapper"
}

func (m *StorageLogMapper) Map(_ context.Context, a any) ([]*rows.StorageLog, error) {
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

    row := &rows.StorageLog{}
    row.Timestamp = item.Timestamp
    row.InsertID = item.InsertID
    row.Severity = item.Severity.String()
    row.LogName = item.LogName

    if item.Payload != nil {
        var protoPayload audit.AuditLog
        payloadBytes, err := json.Marshal(item.Payload)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal payload: %w", err)
        }
        if err := json.Unmarshal(payloadBytes, &protoPayload); err != nil {
            return nil, fmt.Errorf("failed to unmarshal payload to AuditLog: %w", err)
        }

        row.StatusCode = protoPayload.Status.GetCode()
        row.StatusMessage = protoPayload.Status.GetMessage()
        // row.AuthenticationPrincipal = protoPayload.AuthenticationInfo.GetPrincipalEmail()
        // row.CallerIp = protoPayload.RequestMetadata.GetCallerIp()
        // row.CallerSuppliedUserAgent = protoPayload.RequestMetadata.GetCallerSuppliedUserAgent()
        // row.RequestTime = protoPayload.RequestMetadata.RequestAttributes.GetTime()
        row.ServiceName = protoPayload.ServiceName
        row.MethodName = protoPayload.MethodName
        row.ResourceName = protoPayload.ResourceName
        // row.ResourceCurrentLocations = protoPayload.ResourceLocation.GetCurrentLocations()

        // if len(protoPayload.AuthorizationInfo) > 0 {
        //     firstAuthInfo := protoPayload.AuthorizationInfo[0]
        //     row.Resource = firstAuthInfo.Resource
        //     row.Permission = firstAuthInfo.Permission
        //     row.Granted = firstAuthInfo.Granted
        // }
    }

    // if item.Resource != nil {
    //     row.ResourceType = item.Resource.Type
    //     row.Location = item.Resource.Labels["location"]
    //     row.BucketName = item.Resource.Labels["bucket_name"]
    //     row.ProjectId = item.Resource.Labels["project_id"]
    // }

    return []*rows.StorageLog{row}, nil
}
