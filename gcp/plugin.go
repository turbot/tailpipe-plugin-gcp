package gcp

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/tailpipe-plugin-gcp/config"
	"github.com/turbot/tailpipe-plugin-sdk/plugin"
	// reference the table package to ensure that the tables are registered by the init functions
	_ "github.com/turbot/tailpipe-plugin-gcp/tables"
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
		PluginImpl: plugin.NewPluginImpl(config.PluginName),
	}

	return p, nil
}
