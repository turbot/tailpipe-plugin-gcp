package main

import (
	"github.com/turbot/tailpipe-plugin-gcp/gcp"
	"github.com/turbot/tailpipe-plugin-sdk/plugin"
	"log/slog"
	"os"
)

func main() {
	// if the `metadata` arg was passed, we are running in metadata mode - return our metadata
	if len(os.Args) > 1 && os.Args[1] == "metadata" {
		// print the metadata and exit
		os.Exit(plugin.PrintMetadata(gcp.NewPlugin))
	}

	err := plugin.Serve(&plugin.ServeOpts{
		PluginFunc: gcp.NewPlugin,
	})

	if err != nil {
		slog.Error("Error starting plugin", "error", err)
	}
}
