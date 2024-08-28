package gcp

import (
	"github.com/turbot/tailpipe-plugin-gcp/gcp_partition"
	"github.com/turbot/tailpipe-plugin-gcp/gcp_source"
	"github.com/turbot/tailpipe-plugin-sdk/partition"
	"github.com/turbot/tailpipe-plugin-sdk/plugin"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
)

type Plugin struct {
	plugin.PluginBase
}

func NewPlugin() (plugin.TailpipePlugin, error) {
	p := &Plugin{}

	err := p.RegisterResources(
		&plugin.ResourceFunctions{
			Partitions: []func() partition.Partition{gcp_partition.NewAuditLogCollection},
			Sources:    []func() row_source.RowSource{gcp_source.NewAuditLogAPISource},
		})

	if err != nil {
		return nil, err
	}

	return p, nil
}

func (t *Plugin) Identifier() string {
	return "gcp"
}
