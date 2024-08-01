package gcp

import (
	"time"

	"github.com/turbot/tailpipe-plugin-gcp/gcp_collection"
	"github.com/turbot/tailpipe-plugin-sdk/plugin"
)

type Plugin struct {
	plugin.Base
}

func NewPlugin() (plugin.TailpipePlugin, error) {
	p := &Plugin{}
	time.Sleep(10 * time.Second) // TODO: #debug remove this
	err := p.RegisterCollections(gcp_collection.NewAuditLogCollection)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (t *Plugin) Identifier() string {
	return "gcp"
}
