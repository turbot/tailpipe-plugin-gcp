package gcp

import (
	"github.com/turbot/tailpipe-plugin-gcp/gcp_collection"
	"github.com/turbot/tailpipe-plugin-sdk/plugin"
)

type Plugin struct {
	plugin.Base
}

func NewPlugin() (plugin.TailpipePlugin, error) {
	p := &Plugin{}

	err := p.RegisterCollections(gcp_collection.NewAuditLogCollection)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (t *Plugin) Identifier() string {
	return "gcp"
}
