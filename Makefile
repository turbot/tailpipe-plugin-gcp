TAILPIPE_INSTALL_DIR ?= ~/.tailpipe
BUILD_TAGS = netgo

PLUGIN_DIR = $(TAILPIPE_INSTALL_DIR)/plugins/hub.tailpipe.io/plugins/turbot/gcp@latest
PLUGIN_BINARY = $(PLUGIN_DIR)/tailpipe-plugin-gcp.plugin
VERSION_JSON = $(PLUGIN_DIR)/version.json
VERSIONS_JSON = $(TAILPIPE_INSTALL_DIR)/plugins/versions.json

.PHONY: install

install:
	go build -o $(PLUGIN_BINARY) -tags "${BUILD_TAGS}" *.go
	$(PLUGIN_BINARY) metadata > $(VERSION_JSON)
	rm -f $(VERSIONS_JSON)
