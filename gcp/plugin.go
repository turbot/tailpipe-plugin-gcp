package gcp

import (
	"github.com/turbot/tailpipe-plugin-gcp/gcp_source"
	"github.com/turbot/tailpipe-plugin-gcp/gcp_table"
	"github.com/turbot/tailpipe-plugin-sdk/plugin"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/table"
	"time"
)

type Plugin struct {
	plugin.PluginBase
}

func NewPlugin() (plugin.TailpipePlugin, error) {
	p := &Plugin{}

	time.Sleep(10 * time.Second)
	err := p.RegisterResources(
		&plugin.ResourceFunctions{
			Tables:  []func() table.Table{gcp_table.NewAuditLogCollection},
			Sources: []func() row_source.RowSource{gcp_source.NewAuditLogAPISource},
		})

	if err != nil {
		return nil, err
	}

	return p, nil
}

func (t *Plugin) Identifier() string {
	return "gcp"
}
