# This stage is parametrized to replicate the same environment GlooE was built in.
# All ARGs need to be set via the docker `--build-arg` flags.
ARG GO_BUILD_IMAGE
FROM $GO_BUILD_IMAGE AS build-env

ARG GC_FLAGS
ARG VERIFY_SCRIPT

# Fail if VERIFY_SCRIPT not set
RUN if [[ ! $VERIFY_SCRIPT ]]; then echo "Required VERIFY_SCRIPT build argument not set" && exit 1; fi

RUN apk add --no-cache gcc musl-dev

ADD . /go/src/github.com/solo-io/ext-auth-plugin-examples/
WORKDIR /go/src/github.com/solo-io/ext-auth-plugin-examples

# De-vendor all the dependencies and move them to the GOPATH.
# We need this so that the import paths for any library shared between the plugins and Gloo are the same.
RUN cp -a vendor/. /go/src/ && rm -rf vendor

# Build plugin(s) with CGO enabled
RUN CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -buildmode=plugin -gcflags="$GC_FLAGS" -o plugins/RequiredHeader.so plugins/required_header/plugin.go

# Verify that plugin(s) can be loaded by GlooE
RUN chmod +x $VERIFY_SCRIPT
RUN $VERIFY_SCRIPT -pluginDir plugins -manifest plugins/plugin_manifest.yaml

# This stage just copies over the plugin .so files from the previous stage
FROM alpine:3.10.1
RUN mkdir /compiled-auth-plugins
COPY --from=build-env /go/src/github.com/solo-io/ext-auth-plugin-examples/plugins/RequiredHeader.so /compiled-auth-plugins/
# This is the command that will be executed when the container is run.
# It has to copy the compiled plugin file(s) to a directory.
CMD cp /compiled-auth-plugins/* /auth-plugins/