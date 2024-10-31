package gcp

import (
	"github.com/turbot/tailpipe-plugin-gcp/sources"
	"github.com/turbot/tailpipe-plugin-gcp/tables"
	"github.com/turbot/tailpipe-plugin-sdk/plugin"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/table"
)

func NewPlugin() (plugin.TailpipePlugin, error) {
	p := plugin.NewPlugin("gcp")

	err := p.RegisterResources(
		&plugin.ResourceFunctions{
			Tables:  []func() table.Table{tables.NewAuditLogCollection},
			Sources: []func() row_source.RowSource{sources.NewAuditLogAPISource},
		})

	if err != nil {
		return nil, err
	}

	return p, nil
}
