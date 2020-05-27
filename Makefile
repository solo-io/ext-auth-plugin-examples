.PHONY: format
format:
	gofmt -w -e plugins scripts
	goimports -w -e plugins scripts

#----------------------------------------------------------------------------------
# Set build variables
#----------------------------------------------------------------------------------

# Set this variable to the version of GlooE you want to target
GLOOE_VERSION ?= 1.3.4

# Set this variable to the name of your build plugin
PLUGIN_BUILD_NAME ?= RequiredHeader.so

# Set this variable to the image name and tag of your plugin
PLUGIN_IMAGE ?= gloo-ext-auth-plugins:$(GLOOE_VERSION)

# Set this variable to the name of your plugin
PLUGIN_NAME ?= required_header

# Set this variable to the base image name for the container that will have the compiled plugin
RUN_IMAGE ?= alpine:3.11

# Set this variable to the hostname of your custom (air gapped) storage server
STORAGE_HOSTNAME ?= storage.googleapis.com

GLOOE_DIR := _glooe
_ := $(shell mkdir -p $(GLOOE_DIR))

PLUGIN_MODULE_PATH := $(shell grep module go.mod | cut -d ' ' -f 2-)

#----------------------------------------------------------------------------------
# Build an docker image which contains the plugin framework and plugin implementation
#----------------------------------------------------------------------------------
.PHONY: build
build:
	docker build --no-cache \
		--build-arg GO_BUILD_IMAGE=$(call get_glooe_var,GO_BUILD_IMAGE) \
		--build-arg RUN_IMAGE=$(RUN_IMAGE) \
		--build-arg GLOOE_VERSION=$(GLOOE_VERSION) \
		--build-arg STORAGE_HOSTNAME=$(STORAGE_HOSTNAME) \
		--build-arg PLUGIN_MODULE_PATH=$(PLUGIN_MODULE_PATH) \
		-t $(PLUGIN_IMAGE) .

#----------------------------------------------------------------------------------
# Retrieve GlooE build information
#----------------------------------------------------------------------------------
.PHONY: get-glooe-info
get-glooe-info: $(GLOOE_DIR)/dependencies $(GLOOE_DIR)/verify-plugins-linux-amd64 $(GLOOE_DIR)/build_env

$(GLOOE_DIR)/dependencies:
	wget -O $@ http://$(STORAGE_HOSTNAME)/gloo-ee-dependencies/$(GLOOE_VERSION)/dependencies

$(GLOOE_DIR)/verify-plugins-linux-amd64:
	wget -O $@ http://$(STORAGE_HOSTNAME)/gloo-ee-dependencies/$(GLOOE_VERSION)/verify-plugins-linux-amd64

$(GLOOE_DIR)/build_env:
	wget -O $@ http://$(STORAGE_HOSTNAME)/gloo-ee-dependencies/$(GLOOE_VERSION)/build_env

#----------------------------------------------------------------------------------
# Compare dependencies against GlooE
#----------------------------------------------------------------------------------
.PHONY: get-plugin-dependencies
get-plugin-dependencies: go.mod go.sum
	go list -m all > plugin_dependencies

#----------------------------------------------------------------------------------
# Compare and merge mon matching dependencies against GlooE
#----------------------------------------------------------------------------------
.PHONY: resolve-deps
resolve-deps: go.mod $(GLOOE_DIR)/dependencies
	go run scripts/resolve_deps/main.go go.mod $(GLOOE_DIR)/dependencies

#----------------------------------------------------------------------------------
# Build plugins
#----------------------------------------------------------------------------------
EXAMPLES_DIR := plugins
SOURCES := $(shell find . -name "*.go" | grep -v test)

define get_glooe_var
$(shell grep $(1) $(GLOOE_DIR)/build_env | cut -d '=' -f 2-)
endef

.PHONY: build-plugin
build-plugin: compile-plugin verify-plugin

compile-plugin: $(GLOOE_DIR)/build_env
	CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -buildmode=plugin -gcflags=$(call get_glooe_var,GC_FLAGS) -o plugins/$(PLUGIN_BUILD_NAME) plugins/$(PLUGIN_NAME)/plugin.go

verify-plugin: $(GLOOE_DIR)/verify-plugins-linux-amd64
	chmod +x $(GLOOE_DIR)/verify-plugins-linux-amd64
	$(GLOOE_DIR)/verify-plugins-linux-amd64 -pluginDir plugins -manifest plugins/plugin_manifest.yaml

.PHONY: build-plugins-for-tests
build-plugins-for-tests: $(EXAMPLES_DIR)/$(PLUGIN_NAME)/$(PLUGIN_BUILD_NAME)

$(EXAMPLES_DIR)/required_header/RequiredHeader.so: $(SOURCES)
	go build -buildmode=plugin -o $(EXAMPLES_DIR)/$(PLUGIN_NAME)/$(PLUGIN_BUILD_NAME) $(EXAMPLES_DIR)/required_header/plugin.go

clean:
	rm -rf _glooe
