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
GLOOE_VERSION ?= 0.18.21

.PHONY: get-glooe-info
get-glooe-info: $(GLOOE_DIR)/Gopkg.lock $(GLOOE_DIR)/verify-plugins-linux-amd64 $(GLOOE_DIR)/build_env

$(GLOOE_DIR)/Gopkg.lock:
	curl -o $@ http://storage.googleapis.com/gloo-ee-dependencies/$(GLOOE_VERSION)/Gopkg.lock

$(GLOOE_DIR)/verify-plugins-linux-amd64:
	curl -o $@ http://storage.googleapis.com/gloo-ee-dependencies/$(GLOOE_VERSION)/verify-plugins-linux-amd64

$(GLOOE_DIR)/build_env:
	curl -o $@ http://storage.googleapis.com/gloo-ee-dependencies/$(GLOOE_VERSION)/build_env


#----------------------------------------------------------------------------------
# Compare dependencies against GlooE
#----------------------------------------------------------------------------------

.PHONY: compare-deps
compare-deps: Gopkg.lock $(GLOOE_DIR)/Gopkg.lock
	go run scripts/compare_dependencies.go Gopkg.lock $(GLOOE_DIR)/Gopkg.lock


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