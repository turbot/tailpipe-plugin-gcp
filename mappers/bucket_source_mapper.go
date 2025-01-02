package mappers

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-sdk/table"
)

type StorageBucketAuditLogMapper struct {
}

func (m *StorageBucketAuditLogMapper) Identifier() string {
	return "gcp_storage_bucket_audit_log_mapper"
}

func (m *StorageBucketAuditLogMapper) Map(_ context.Context, a any, _ ...table.MapOption[*rows.AuditLog]) (*rows.AuditLog, error) {
	var item rows.BucketSourceLogEntry

	switch v := a.(type) {
	case string:
		err := json.Unmarshal([]byte(v), &item)
		if err != nil {
			return nil, err
		}
	}

	row := rows.NewAuditLog()
	row.Timestamp = item.Timestamp
	row.LogName = item.LogName
	row.InsertId = item.InsertID
	row.Severity = item.Severity

	// TODO: Rescale those properties
	// row.Trace = item.Trace
	// row.TraceSampled = item.TraceSampled
	// row.SpanId = item.SpanID

	if item.ProtoPayload != nil {
		row.ServiceName = item.ProtoPayload.ServiceName
		row.MethodName = item.ProtoPayload.MethodName
		row.ResourceName = item.ProtoPayload.ResourceName
		// row.NumResponseItems = &item.ProtoPayload.NumResponseItems

		if item.ProtoPayload.Status != nil {
			row.Status = &rows.AuditLogStatus{
				Code:    int32(*item.ProtoPayload.Status.Code),
				Message: item.ProtoPayload.Status.Message,
			}
		}

		if item.ProtoPayload.AuthenticationInfo != nil {
			row.AuthenticationInfo = &rows.AuditLogAuthenticationInfo{
				PrincipalEmail: *item.ProtoPayload.AuthenticationInfo.PrincipalEmail,
				// PrincipalSubject:      item.ProtoPayload.AuthenticationInfo.principalSubject,
				// AuthoritySelector:     payload.AuthenticationInfo.AuthoritySelector,
				// ServiceAccountKeyName: payload.AuthenticationInfo.ServiceAccountKeyName,
			}

			// if payload.AuthenticationInfo.ThirdPartyPrincipal != nil {
			// 	tpp := payload.AuthenticationInfo.ThirdPartyPrincipal.GetFields()
			// 	row.AuthenticationInfo.ThirdPartyPrincipal = make(map[string]string, len(tpp))
			// 	for k, v := range tpp {
			// 		row.AuthenticationInfo.ThirdPartyPrincipal[k] = v.String()
			// 	}
			// }

			if payload.AuthenticationInfo.ServiceAccountDelegationInfo != nil {
				for _, v := range payload.AuthenticationInfo.ServiceAccountDelegationInfo {
					row.AuthenticationInfo.ServiceAccountDelegationInfo = append(row.AuthenticationInfo.ServiceAccountDelegationInfo, v.PrincipalSubject)
				}
			}
		}

		if item.ProtoPayload.RequestMetadata != nil {
			row.RequestMetadata = &rows.AuditLogRequestMetadata{
				CallerIp:                *item.ProtoPayload.RequestMetadata.CallerIP,
				CallerSuppliedUserAgent: *item.ProtoPayload.RequestMetadata.CallerSuppliedUserAgent,
				// CallerNetwork:           *item.ProtoPayload.RequestMetadata.CallerNetwork,
				// RequestAttributes:       &attribute_context.AttributeContext_Request{
				// *item.ProtoPayload.RequestMetadata.RequestAttributes.Auth[""],
			}

			if item.ProtoPayload.RequestMetadata.DestinationAttributes != nil {
				row.RequestMetadata.DestinationAttributes = &rows.AuditLogRequestMetadataDestinationAttributes{
					Ip:         &item.ProtoPayload.RequestMetadata.DestinationAttributes["ip"],
					Port:       payload.RequestMetadata.DestinationAttributes.Port,
					Principal:  payload.RequestMetadata.DestinationAttributes.Principal,
					RegionCode: payload.RequestMetadata.DestinationAttributes.RegionCode,
					Labels:     payload.RequestMetadata.DestinationAttributes.Labels,
				}
			}
		}
	}

	// resource
	if item.Resource != nil {
		row.Resource = &rows.AuditLogResource{
			Type:   item.Resource.Type,
			Labels: *item.Resource.Labels,
		}
	}

	// labels
	if item.Labels != nil {
		row.Labels = &item.Labels
	}

	slog.Debug("Payload of storage bucket source service name: ", row.ServiceName, "condition satisfied.")

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
