package gcp

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/tailpipe-plugin-gcp/config"
	"github.com/turbot/tailpipe-plugin-gcp/sources/audit_log_api"
	"github.com/turbot/tailpipe-plugin-gcp/sources/storage_bucket"
	"github.com/turbot/tailpipe-plugin-gcp/tables/audit_log"
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

	// register sources
	row_source.RegisterRowSource[*audit_log_api.AuditLogAPISource]()
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
