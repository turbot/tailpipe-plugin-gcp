package gcp

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/tailpipe-plugin-gcp/config"
	logging_log_entry "github.com/turbot/tailpipe-plugin-gcp/sources/logging_log_entry"
	"github.com/turbot/tailpipe-plugin-gcp/sources/storage_bucket"
	"github.com/turbot/tailpipe-plugin-gcp/tables/audit_log"
	"github.com/turbot/tailpipe-plugin-gcp/tables/billing_report"
	"github.com/turbot/tailpipe-plugin-sdk/plugin"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/table"
)

type Plugin struct {
	plugin.PluginImpl
}

func init() {
	// Register tables, with type parameters:
	// 1. row struct
	// 2. table implementation
	table.RegisterTable[*audit_log.AuditLog, *audit_log.AuditLogTable]()
	table.RegisterCustomTable[*billing_report.BillingReportTable]()

	// register sources
	row_source.RegisterRowSource[*logging_log_entry.LoggingLogEntrySource]()
	// Register deprecated source for backward compatibility
	// This will be removed in a future version
	row_source.RegisterRowSource[*logging_log_entry.DeprecatedLoggingLogEntrySource]()
	row_source.RegisterRowSource[*storage_bucket.GcpStorageBucketSource]()
}

func NewPlugin() (_ plugin.TailpipePlugin, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	p := &Plugin{
		PluginImpl: plugin.NewPluginImpl(config.PluginName),
	}

	return p, nil
}
