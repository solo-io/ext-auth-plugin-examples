# Prepare the build environment.
# Use this stage to add certificates and set proxies
# All ARGs need to be set via the docker `--build-arg` flags.
ARG GO_BUILD_IMAGE
ARG RUN_IMAGE
FROM $GO_BUILD_IMAGE AS build-env

# This stage is parametrized to replicate the same environment Gloo Enterprise was built in.
# It is important to use the same container to build the plugin that Gloo Enterprise was
# built in to ensure that the same go version and linker is used during compilation.
# All ARGs need to be set via the docker `--build-arg` flags.
FROM build-env as build
ARG GLOOE_VERSION
ARG STORAGE_HOSTNAME
ARG PLUGIN_MODULE_PATH

ENV GONOSUMDB=*
ENV GLOOE_VERSION=$GLOOE_VERSION

# We don't have the same check as on GC_FLAGS as empty values are allowed there
RUN if [ ! $GLOOE_VERSION ]; then echo "Required GLOOE_VERSION build argument not set" && exit 1; fi
RUN if [ ! $STORAGE_HOSTNAME ]; then echo "Required STORAGE_HOSTNAME build argument not set" && exit 1; fi

# Install packages needed for compilation
RUN apk add --no-cache gcc musl-dev git make

# Sets working dir to the correct directory
# /go/src to support older versions of Gloo that built plugins with go modules disabled (i.e., gopath builds)
WORKDIR /go/src/$PLUGIN_MODULE_PATH

# Resolve dependencies and ensure dependency version usage
COPY Makefile go.mod go.sum ./
COPY pkg ./pkg
COPY scripts ./scripts
COPY plugins ./plugins

RUN make resolve-deps
RUN echo "// Generated for GlooE $GLOOE_VERSION" | cat - go.mod > go.new && mv go.new go.mod
# Compile and verify the plugin can be loaded by Gloo
RUN make build-plugin || { echo "Used module:" | cat - go.mod; exit 1; }

# This stage builds the final image containing just the plugin .so files. It can really be any linux/amd64 image.
FROM $RUN_IMAGE
ARG PLUGIN_MODULE_PATH

# Copy compiled plugin file from previous stage
RUN mkdir /compiled-auth-plugins
COPY --from=build /go/src/$PLUGIN_MODULE_PATH/plugins/*.so /compiled-auth-plugins/
COPY --from=build /go/src/$PLUGIN_MODULE_PATH/go.mod /compiled-auth-plugins/

# This is the command that will be executed when the container is run.
# It has to copy the compiled plugin file(s) to a directory.
CMD cp /compiled-auth-plugins/*.so /auth-plugins/