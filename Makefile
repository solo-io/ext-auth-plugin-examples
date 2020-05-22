.PHONY: format
format:
	gofmt -w -e plugins scripts
	goimports -w -e plugins scripts

#----------------------------------------------------------------------------------
# Set build variables
#----------------------------------------------------------------------------------
# Set this variable to the name of your plugin
PLUGIN_NAME ?= sample

# Set this variable to the version of your plugin
PLUGIN_VERSION ?= 0.0.1

# Set this variable to the version of GlooE you want to target
GLOOE_VERSION ?= 1.3.1

# Set this variable to the base image name for the container that will have the compiled plugin
RUN_IMAGE ?= alpine:3.11

# Set this variable to the hostname of your custom (air gapped) storage server
STORAGE_HOSTNAME ?= storage.googleapis.com

GLOOE_DIR := _glooe
_ := $(shell mkdir -p $(GLOOE_DIR))

PLUGIN_PATH := $(shell grep module go.mod | cut -d ' ' -f 2-)
PLUGIN_IMAGE := gloo-ext-auth-plugin-$(PLUGIN_NAME):$(PLUGIN_VERSION)


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
build-plugin: $(GLOOE_DIR)/build_env $(GLOOE_DIR)/verify-plugins-linux-amd64
	docker build --no-cache \
		--build-arg GO_BUILD_IMAGE=$(call get_glooe_var,GO_BUILD_IMAGE) \
		--build-arg RUN_IMAGE=$(RUN_IMAGE) \
		--build-arg GC_FLAGS=$(call get_glooe_var,GC_FLAGS) \
		--build-arg VERIFY_SCRIPT=$(GLOOE_DIR)/verify-plugins-linux-amd64 \
		--build-arg STORAGE_HOSTNAME=$(STORAGE_HOSTNAME) \
		--build-arg PLUGIN_PATH=$(PLUGIN_PATH) \
		-t $(PLUGIN_IMAGE) .

.PHONY: compile-plugin
compile-plugin: $(GLOOE_DIR)/build_env
	CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -buildmode=plugin -gcflags="$(call get_glooe_var,GC_FLAGS)" -o plugins/RequiredHeader.so plugins/required_header/plugin.go

.PHONY: build-plugins-for-tests
build-plugins-for-tests: $(EXAMPLES_DIR)/required_header/RequiredHeader.so

$(EXAMPLES_DIR)/required_header/RequiredHeader.so: $(SOURCES)
	go build -buildmode=plugin -o $(EXAMPLES_DIR)/required_header/RequiredHeader.so $(EXAMPLES_DIR)/required_header/plugin.go

clean:
	rm -rf _glooe
