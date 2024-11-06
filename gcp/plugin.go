package gcp

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/tailpipe-plugin-gcp/config"
	"github.com/turbot/tailpipe-plugin-gcp/sources"
	"github.com/turbot/tailpipe-plugin-gcp/tables"
	"github.com/turbot/tailpipe-plugin-sdk/plugin"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/table"
)

type Plugin struct {
	plugin.PluginImpl
}

func NewPlugin() (_ plugin.TailpipePlugin, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	p := &Plugin{
		PluginImpl: plugin.NewPluginImpl("gcp", config.NewGcpConnection),
	}

	resources := &plugin.ResourceFunctions{
		Tables:  []func() table.Table{tables.NewAuditLogTable},
		Sources: []func() row_source.RowSource{sources.NewAuditLogAPISource},
	}

	if err := p.RegisterResources(resources); err != nil {
		return nil, err
	}
	return p, nil
}
