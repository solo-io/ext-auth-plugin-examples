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
ARG PLUGIN_PATH

ENV GONOSUMDB=*
ENV GO111MODULE=on
ENV CGO_ENABLED=1

# This must contain the path to the plugin verification script
ARG VERIFY_SCRIPT

# Fail if VERIFY_SCRIPT not set
# We don't have the same check as on GC_FLAGS as empty values are allowed there
RUN if [[ ! $VERIFY_SCRIPT ]]; then echo "Required VERIFY_SCRIPT build argument not set" && exit 1; fi

# Install packages needed for compilation
RUN apk add --no-cache gcc musl-dev git make

WORKDIR $PLUGIN_PATH
# Resolve dependencies and ensure dependency version usage
COPY Makefile go.mod go.sum ./
COPY pkg ./pkg
COPY scripts ./scripts
COPY plugins ./plugins

RUN make get-glooe-info resolve-deps
RUN echo "// Generated for GlooE $GLOOE_VERSION" | cat - go.mod > go.new && mv go.new go.mod
RUN make compile-plugin || { echo "Used module:" | cat - go.mod; exit 1; }

# Run the script to verify that the plugin(s) can be loaded by Gloo
RUN chmod +x $VERIFY_SCRIPT
RUN $VERIFY_SCRIPT -pluginDir plugins -manifest plugins/plugin_manifest.yaml || { echo "Used module:" | cat - go.mod; exit 1; }

# This stage builds the final image containing just the plugin .so files. It can really be any linux/amd64 image.
FROM $RUN_IMAGE
ARG PLUGIN_PATH

# Copy compiled plugin file from previous stage
RUN mkdir /compiled-auth-plugins
COPY --from=build /go/$PLUGIN_PATH/plugins/RequiredHeader.so /compiled-auth-plugins/
COPY --from=build /go/$PLUGIN_PATH/go.mod /compiled-auth-plugins/

# This is the command that will be executed when the container is run.
# It has to copy the compiled plugin file(s) to a directory.
CMD cp /compiled-auth-plugins/*.so /auth-plugins/