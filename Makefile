TAILPIPE_INSTALL_DIR ?= ~/.tailpipe
BUILD_TAGS = netgo
install:
	go build -o $(TAILPIPE_INSTALL_DIR)/plugins/hub.tailpipe.io/plugins/turbot/gcp@latest/tailpipe-plugin-gcp.plugin -tags "${BUILD_TAGS}" *.go

## Paths
#PLUGIN_NAME=tailpipe-plugin-gcp.plugin
#PLUGIN_DIR=~/.tailpipe/plugins/hub.tailpipe.io/plugins/turbot/gcp@latest/
#
## Build in development mode by default
#.PHONY: default
#default: install
#
## Production build, optimized
#.PHONY: build
#build:
#	go build -o $(PLUGIN_NAME) .
#
## Install the development build
#.PHONY: install
#install: build
#	mv $(PLUGIN_NAME) $(PLUGIN_DIR)
#
## Run tests
#.PHONY: test
#test:
#	go test ./... -v
#
## Clean up generated files
#.PHONY: clean
#clean:
#	rm -f $(PLUGIN_NAME)