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
RUN apk add --no-cache gcc musl-dev

# Copy the repository to the image and set it as the working directory. The GOPATH here is `/go`.
#
# You have to update the path here to the one corresponding to your repository. This is usually in the form:
# /go/src/github.com/YOUR_ORGANIZATION/PLUGIN_REPO_NAME
ADD . /go/src/github.com/solo-io/ext-auth-plugin-examples/
WORKDIR /go/src/github.com/solo-io/ext-auth-plugin-examples

# De-vendor all the dependencies and move them to the GOPATH, so they will be loaded from there.
# We need this so that the import paths for any library shared between the plugins and Gloo are the same.
#
# For example, if we were to vendor the ext-auth-plugin dependency, the ext-auth-server would load the plugin interface
# as `GLOOE_REPO/vendor/github.com/solo-io/ext-auth-plugins/api.ExtAuthPlugin`, while the plugin
# would instead implement `THIS_REPO/vendor/github.com/solo-io/ext-auth-plugins/api.ExtAuthPlugin`. These would be seen
# by the go runtime as two different types, causing Gloo to fail.
# Also, some packages cause problems if loaded more than once. For example, loading `golang.org/x/net/trace` twice
# causes a panic (see here: https://github.com/golang/go/issues/24137). By flattening the dependencies this way we
# prevent these sorts of problems.
RUN cp -a vendor/. /go/src/ && rm -rf vendor

# Build plugins with CGO enabled, passing the GC_FLAGS flags
RUN CGO_ENABLED=1 GOARCH=amd64 GOOS=linux GO111MODULE=off go build -buildmode=plugin -gcflags="$GC_FLAGS" -o plugins/RequiredHeader.so plugins/required_header/plugin.go

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