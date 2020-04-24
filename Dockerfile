# This stage is parametrized to replicate the same environment Gloo Enterprise was built in.
# All ARGs need to be set via the docker `--build-arg` flags.
ARG GO_BUILD_IMAGE
FROM $GO_BUILD_IMAGE AS build-env

# This must contain the value of the `gcflag` build flag that Gloo was built with
ARG GC_FLAGS
# This must contain the path to the plugin verification script
ARG VERIFY_SCRIPT

# Fail if VERIFY_SCRIPT not set
# We don't have the same check as on GC_FLAGS as empty values are allowed there
RUN if [[ ! $VERIFY_SCRIPT ]]; then echo "Required VERIFY_SCRIPT build argument not set" && exit 1; fi

# Install packages needed for compilation
RUN apk add --no-cache gcc musl-dev git

# Copy the repository to the image and set it as the working directory. The GOPATH here is `/go`.
# The directory chosen here is arbitrary and does not influence the plugins compatibility with Gloo.
ADD . /go/src/github.com/solo-io/ext-auth-plugin-examples/
WORKDIR /go/src/github.com/solo-io/ext-auth-plugin-examples

# Build plugins with CGO enabled, passing the GC_FLAGS flags
ENV CGO_ENABLED=1 GOARCH=amd64 GOOS=linux
RUN go build -buildmode=plugin -gcflags="$GC_FLAGS" -o plugins/RequiredHeader.so plugins/required_header/plugin.go

# Run the script to verify that the plugin(s) can be loaded by Gloo
RUN chmod +x $VERIFY_SCRIPT
RUN $VERIFY_SCRIPT -pluginDir plugins -manifest plugins/plugin_manifest.yaml

# This stage builds the final image containing just the plugin .so files. It can really be any linux/amd64 image.
FROM alpine:3.10.1

# Copy compiled plugin file from previous stage
RUN mkdir /compiled-auth-plugins
COPY --from=build-env /go/src/github.com/solo-io/ext-auth-plugin-examples/plugins/RequiredHeader.so /compiled-auth-plugins/

# This is the command that will be executed when the container is run.
# It has to copy the compiled plugin file(s) to a directory.
CMD cp /compiled-auth-plugins/* /auth-plugins/