package mappers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-sdk/table"
	"google.golang.org/genproto/googleapis/cloud/audit"
)

type StorageBucketAuditLogMapper struct{}

func (m *StorageBucketAuditLogMapper) Identifier() string {
	return "gcp_storage_bucket_audit_log_mapper"
}

func (m *StorageBucketAuditLogMapper) Map(_ context.Context, a any, _ ...table.MapOption[*rows.AuditLog]) (*rows.AuditLog, error) {
	// var item rows.BucketSourceLogEntry

	// Unmarshal input into `BucketSourceLogEntry`
	switch v := a.(type) {
	case string:
		row, err := unmarshalAuditLog([]byte(v))
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal input string: %w", err)
		}

		return row, nil
	default:
		return nil, fmt.Errorf("unsupported input type: %T", v)
	}
}

func unmarshalAuditLog(jsonData []byte) (*rows.AuditLog, error) {
	// Create a new AuditLog instance
	auditLog := rows.NewAuditLog()

	// Create a generic map to handle protoPayload and other properties
	var rawData map[string]interface{}
	if err := json.Unmarshal(jsonData, &rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Map mandatory fields
	if timestamp, ok := rawData["timestamp"].(string); ok {
		parsedTime, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp format: %w", err)
		}
		auditLog.Timestamp = parsedTime
	}
	auditLog.LogName = getString(rawData["logName"])
	auditLog.InsertId = getString(rawData["insertId"])
	
	auditLog.Severity = getString(rawData["severity"])

	// Map optional protoPayload
	if protoPayload, ok := rawData["protoPayload"].(map[string]interface{}); ok {
		auditLog.ServiceName = getStringPointer(protoPayload["serviceName"])
		auditLog.MethodName = getStringPointer(protoPayload["methodName"])
		auditLog.ResourceName = getStringPointer(protoPayload["resourceName"])

		// AuthenticationInfo
		if authInfo, ok := protoPayload["authenticationInfo"].(map[string]interface{}); ok {
			auditLog.AuthenticationInfo = &rows.AuditLogAuthenticationInfo{
				PrincipalEmail: getString(authInfo["principalEmail"]),
			}
		}

		// Status
		if status, ok := protoPayload["status"].(map[string]interface{}); ok {
			auditLog.Status = &rows.AuditLogStatus{
				Code:    int32(getInt(status["code"])),
				Message: getString(status["message"]),
			}
		}

		// ResourceLocation
		if resourceLocation, ok := protoPayload["resourceLocation"].(map[string]interface{}); ok {
			auditLog.ResourceLocation = &rows.AuditLogResourceLocation{
				CurrentLocations: getStringSlice(resourceLocation["currentLocations"]),
			}
		}

		// AuthorizationInfo
		if authInfoSlice, ok := protoPayload["authorizationInfo"].([]interface{}); ok {
			for _, authInfo := range authInfoSlice {
				if infoMap, ok := authInfo.(map[string]interface{}); ok {
					auditLog.AuthorizationInfo = append(auditLog.AuthorizationInfo, &audit.AuthorizationInfo{
						Permission: getString(infoMap["permission"]),
					})
				}
			}
		}

		// Request and Response
		auditLog.Request = getMap(protoPayload["request"])
		auditLog.Response = getMap(protoPayload["response"])
	}

	// Map resource field
	if resource, ok := rawData["resource"].(map[string]interface{}); ok {
		auditLog.Resource = &rows.AuditLogResource{
			Type:   getString(resource["type"]),
			Labels: getStringMap(resource["labels"]),
		}
	}

	// Map labels
	if labels, ok := rawData["labels"].(map[string]interface{}); ok {
		auditLog.Labels = &map[string]string{}
		for k, v := range labels {
			(*auditLog.Labels)[k] = getString(v)
		}
	}

	if auditLog.InsertId == "-k8zmxtd5dic" {
		slog.Debug("Insert ID found Storage source ==>>", auditLog.InsertId, auditLog)
	}

	return auditLog, nil
}

// Helper functions

func getString(value interface{}) string {
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

func getStringPointer(value interface{}) *string {
	if str, ok := value.(string); ok {
		return &str
	}
	return nil
}

func getInt(value interface{}) int {
	if num, ok := value.(float64); ok { // JSON numbers are float64
		return int(num)
	}
	return 0
}

func getStringSlice(value interface{}) []string {
	if slice, ok := value.([]interface{}); ok {
		var result []string
		for _, item := range slice {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return nil
}

func getStringMap(value interface{}) map[string]string {
	result := make(map[string]string)
	if rawMap, ok := value.(map[string]interface{}); ok {
		for k, v := range rawMap {
			result[k] = getString(v)
		}
	}
	return result
}

func getMap(value interface{}) map[string]interface{} {
	if m, ok := value.(map[string]interface{}); ok {
		return m
	}
	return nil
}
