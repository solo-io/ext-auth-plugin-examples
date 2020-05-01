.PHONY: format
format:
	gofmt -w -e plugins scripts
	goimports -w -e plugins scripts

#----------------------------------------------------------------------------------
# Retrieve GlooE build information
#----------------------------------------------------------------------------------
GLOOE_DIR := _glooe
_ := $(shell mkdir -p $(GLOOE_DIR))

# Set this variable to the version of GlooE you want to target
GLOOE_VERSION ?= 1.3.4

# Set this variable to the hostname of your custom (air gapped) storage server
STORAGE_HOSTNAME ?= storage.googleapis.com

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

.PHONY: compare-deps
compare-deps: get-plugin-dependencies $(GLOOE_DIR)/dependencies
	go run scripts/compare_deps/main.go plugin_dependencies $(GLOOE_DIR)/dependencies

#----------------------------------------------------------------------------------
# Compare and merged dependencies against GlooE
#----------------------------------------------------------------------------------
.PHONY: merge-deps
MERGE_ATTEMPTS ?= 10
merge-deps: get-plugin-dependencies $(GLOOE_DIR)/dependencies
	go run scripts/merge_deps/main.go plugin_dependencies $(GLOOE_DIR)/dependencies $(MERGE_ATTEMPTS)

#----------------------------------------------------------------------------------
# Build plugins
#----------------------------------------------------------------------------------
EXAMPLES_DIR := plugins
SOURCES := $(shell find . -name "*.go" | grep -v test)

define get_glooe_var
$(shell grep $(1) $(GLOOE_DIR)/build_env | cut -d '=' -f 2-)
endef

.PHONY: build-plugins
build-plugins: $(GLOOE_DIR)/build_env $(GLOOE_DIR)/verify-plugins-linux-amd64
	docker build --no-cache \
		--build-arg GO_BUILD_IMAGE=$(call get_glooe_var,GO_BUILD_IMAGE) \
		--build-arg GC_FLAGS=$(call get_glooe_var,GC_FLAGS) \
		--build-arg VERIFY_SCRIPT=$(GLOOE_DIR)/verify-plugins-linux-amd64 \
		.

.PHONY: build-plugins-for-tests
build-plugins-for-tests: $(EXAMPLES_DIR)/required_header/RequiredHeader.so

$(EXAMPLES_DIR)/required_header/RequiredHeader.so: $(SOURCES)
	go build -buildmode=plugin -o $(EXAMPLES_DIR)/required_header/RequiredHeader.so $(EXAMPLES_DIR)/required_header/plugin.go

clean:
	rm -rf _glooe
	rm suggestions mismatched_dependencies.json plugin_dependencies
