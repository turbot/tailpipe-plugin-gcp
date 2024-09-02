package gcp_table

import (
	"fmt"
	"time"

	"cloud.google.com/go/logging"
	"github.com/rs/xid"
	"github.com/turbot/tailpipe-plugin-gcp/gcp_types"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
	"github.com/turbot/tailpipe-plugin-sdk/helpers"
	"github.com/turbot/tailpipe-plugin-sdk/parse"
	"github.com/turbot/tailpipe-plugin-sdk/table"
	"google.golang.org/genproto/googleapis/cloud/audit"
)

type AuditLogTable struct {
	table.TableBase[*AuditLogTableConfig]
}

func NewAuditLogCollection() table.Table {
	return &AuditLogTable{}
}

func (c *AuditLogTable) Identifier() string {
	return "gcp_audit_log"
}

func (c *AuditLogTable) GetRowSchema() any {
	return gcp_types.AuditLogRow{}
}

func (c *AuditLogTable) GetConfigSchema() parse.Config {
	return &AuditLogTableConfig{}
}

func (c *AuditLogTable) EnrichRow(row any, sourceEnrichmentFields *enrichment.CommonFields) (any, error) {
	item, ok := row.(logging.Entry)
	if !ok {
		return nil, fmt.Errorf("invalid row type: %T, expected logging.Entry", row)
	}

	payload, ok := item.Payload.(*audit.AuditLog)
	if !ok {
		return nil, fmt.Errorf("invalid payload type: %T, expected *audit.AuditLog", item.Payload)
	}

	if sourceEnrichmentFields == nil || sourceEnrichmentFields.TpIndex == "" {
		return nil, fmt.Errorf("source must provide connection in sourceEnrichmentFields")
	}

	record := &gcp_types.AuditLogRow{CommonFields: *sourceEnrichmentFields}

	// Record Standardization
	record.TpID = xid.New().String()
	record.TpSourceType = "gcp_audit_log" // TODO: Verify (might be able to use specific log type)
	record.TpTimestamp = helpers.UnixMillis(item.Timestamp.UnixNano() / int64(time.Millisecond))
	record.TpIngestTimestamp = helpers.UnixMillis(time.Now().UnixNano() / int64(time.Millisecond))

	// Record Data
	record.Timestamp = item.Timestamp
	record.LogName = item.LogName
	record.InsertId = item.InsertID
	record.Severity = item.Severity.String()
	record.ServiceName = payload.ServiceName
	record.MethodName = payload.MethodName
	record.ResourceName = payload.ResourceName

	if payload.Status != nil {
		record.StatusCode = &payload.Status.Code
		record.StatusMessage = &payload.Status.Message
	}

	if item.Resource != nil {
		record.ResourceType = &item.Resource.Type
		//record.ResourceLabels = &item.Resource.Labels // TODO: #finish add back in once we have support for map
	}

	if item.Operation != nil {
		record.OperationId = &item.Operation.Id
		record.OperationProducer = &item.Operation.Producer
		record.OperationFirst = &item.Operation.First
		record.OperationLast = &item.Operation.Last
	}

	if payload.AuthenticationInfo != nil {
		record.TpUsernames = append(record.TpUsernames, payload.AuthenticationInfo.PrincipalEmail)
		record.AuthenticationPrincipal = &payload.AuthenticationInfo.PrincipalEmail
	}

	if payload.RequestMetadata != nil {
		record.TpSourceIP = &payload.RequestMetadata.CallerIp
		record.TpIps = append(record.TpIps, payload.RequestMetadata.CallerIp)
		record.RequestCallerIp = &payload.RequestMetadata.CallerIp
		record.RequestCallerSuppliedUserAgent = &payload.RequestMetadata.CallerSuppliedUserAgent
	}

	// TODO: #finish payload.Request is a struct which has `Fields` property of map[string]*Value - how to handle? (common keys: @type / name - but this can literally contain anything!)
	// TODO: #finish payload.AuthorizationInfo is an array of structs with Resource (string), Permission (string), and Granted (bool) properties, seems to mostly be a single item but could be more - best way to handle?

	// Hive Fields
	record.TpPartition = "gcp_audit_log"
	record.TpYear = int32(item.Timestamp.Year())
	record.TpMonth = int32(item.Timestamp.Month())
	record.TpDay = int32(item.Timestamp.Day())

	return record, nil
}
