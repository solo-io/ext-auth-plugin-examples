<h1 align="center">
    <img src="https://github.com/solo-io/ext-auth-plugin-examples/raw/master/img/gloo-plugin.png" alt="Gloo Plugins" width="440" height="309">
  <br>
  External auth plugin examples
</h1>

This repository contains example implementations of the 
[ExtAuthPlugin interface](https://github.com/solo-io/ext-auth-plugins/blob/master/api/interface.go) and a set 
of utilities that you can (and should!) use when building your own plugins.

Please refer to the [Auth Plugin Developer Guide](https://docs.solo.io/gloo/latest/guides/dev/writing_auth_plugins/) for an in-depth 
explanation on how you can use this repository as a template to write your own Gloo Auth plugins.

---
**NOTE**

The following instructions are assuming you are targeting Gloo Enterprise `v1.x` releases. If you are using a `v0.x` 
version of Gloo Enterprise, please refer to [this revision](https://github.com/solo-io/ext-auth-plugin-examples/tree/v0.1.1) 
of this repository.

---

## Get example images
You can get the images for the example plugin(s) contained in this repository by running:

```bash
docker pull quay.io/solo-io/ext-auth-plugins:<glooe_version>
```

where the tag `glooe_version` is the version of Gloo Enterprise you want to run the plugins with, e.g. `1.3.8`.

## Publishing your own plugins
The images you created can be [published](#push-image) to a docker registry that is reachable from the cluster you are running Gloo Enterprise in.

## Example workflow
Assuming that you create a plugin called ExamplePlugin.

### Setup environment
* Create a copy of this template repo
* Rename the directory `required_header` to `example_plugin` in the `plugins` directory. 
* Change the code in `plugins/example_plugin/pkg/impl.go` with your custom code.
* Change the module name in [go.mod](go.mod)

### building the plugin
First, store the version of Gloo Enterprise you want to target in an environment variable:
```
export GLOOE_VERSION=1.3.8
export PLUGIN_NAME=example_plugin
export PLUGIN_BUILD_NAME=ExamplePlugin.so
export PLUGIN_IMAGE=gloo-ext-auth-plugin-${PLUGIN_NAME}-${GLOOE_VERSION}:0.0.1
```

#### Containerized (docker) build
Run [`build`](#build) to build the plugin in a container.
```
make build
```

#### Local build
Run [`get-glooe-info`](#get-glooe-info) to fetch the build information for the targeted Gloo Enterprise version:
```
make get-glooe-info
```

Run [`resolve-deps`](#resolve-deps) to check and resolve any dependency mismatches between gloo and your plugin:

```
make resolve-deps
```

Run [`build-plugin`](#build-plugin) to build the plugin.
It is important to build the plugins on a linux operating system, because this is compatible with Gloo Enterprise. 
This step will also verify that they can be loaded.

Note: It is recommended to build in docker to ensure that other plugin/C dependencies for plugin compilation match the
same environment Gloo Enterprise was built in (e.g., linker)

```
make build-plugin
```

### Tag image
In both previous build cases, an anonymous container with the plugin is build for you. The output will look like so:
```
...
 ---> 4846ae5e0a4d
Step 15/16 : COPY --from=build-env /go/src/github.com/solo-io/ext-auth-plugin-examples/plugins/RequiredHeader.so /compiled-auth-plugins/
 ---> 1932dbdca716
Step 16/16 : CMD cp /compiled-auth-plugins/*.so /auth-plugins/
 ---> Running in c33580aff7e1
Removing intermediate container c33580aff7e1
 ---> c6d2a92c47e4
Successfully built c6d2a92c47e4
```

Specifically note the `Successfully built c6d2a92c47e4` line. 
In our case `c6d2a92c47e4` is the ID of the container just built.
To push this container, first re-tag it with your docker registry:

```
docker tag c6d2a92c47e4 username/dockerrepo:v1
```

### Push image

```
docker push username/dockerrepo:v1
```

You can now use this image to load your plugin into the Gloo Enterprise external auth server.
See [this section](https://docs.solo.io/gloo/latest/guides/security/auth/plugin_auth/#installation) of our docs 
for an example of how to do this.

#### Configurable options
The following options can be used to create a framework and/or plugin images
These options can be set by changing its value in the `Makefile`, exporting them as a environment variable (`export GLOOE_VERSION=1.3.8`)
or as command argument (`GLOOE_VERSION=1.3.8 make <target>` )

| Option | Default | Description |
| ------ | ------- | ----------- |
| GLOOE_VERSION | 1.3.8 | Set this variable to the version of GlooE you want to target |
| PLUGIN_BUILD_NAME | RequiredHeader.so | Set this variable to the name of your build plugin |
| PLUGIN_IMAGE | gloo-ext-auth-plugins:$(GLOOE_VERSION) | Set this variable to the image name and tag of your plugin |
| PLUGIN_NAME | required_header | Set this variable to the name of your plugin |
| RUN_IMAGE | alpine:3.11 | Set this variable to the image name and version used for running the plugin |
| STORAGE_HOSTNAME | storage.googleapis.com | Set this variable to the hostname of your custom (air gapped) storage server |


## Common Errors

### plugin was built with a different version of package
You might see an error similar to this one in the logs for the [`build-plugin`](#build-plugin) target:
```
{"level":"error","ts":"2019-12-17T20:59:17.301Z","logger":"verify-plugins","caller":"scripts/verify_plugins.go:54","msg":"Plugin(s) cannot be loaded by Gloo","error":"failed to load plugin: failed to open plugin file: plugin.Open(\"plugins/RequiredHeader\"): plugin was built with a different version of package github.com/golang/protobuf/proto","errorVerbose":"failed to load plugin:\n    github.com/solo-io/go-utils/errors.Wrapf\n        /go/src/github.com/solo-io/go-utils/errors/utils.go:12\n  - failed to open plugin file:\n    github.com/solo-io/go-utils/errors.Wrapf\n        /go/src/github.com/solo-io/go-utils/errors/utils.go:12\n  - plugin.Open(\"plugins/RequiredHeader\"): plugin was built with a different version of package github.com/golang/protobuf/proto","stacktrace":"main.main\n\t/go/src/github.com/solo-io/solo-projects/projects/extauth/scripts/verify_plugins.go:54\nruntime.main\n\t/usr/local/go/src/runtime/proc.go:200"}
```

This is caused by a dependency mismatch. Please run the [`resolve-deps`](#resolve-deps) target to update your `go.mod` file.
Note that the mismatch can also be caused by building in a different mode than Gloo Enterprise (go modules vs go path,
as the path to the dependency is included in the version mismatch check). The make target should try to determine how
to build your plugin for you (go mod vendor -> go path build or just building go modules) using the version of Gloo
Enterprise you are targeting.

## Makefile overview
Following is an overview of the most relevant `make` targets.

### get-glooe-info
When you are writing your own Ext Auth plugins, you must target a specific Gloo Enterprise version. This is because of the 
nature of Go plugins (you can find more info in [this section](https://docs.solo.io/gloo/latest/guides/dev/writing_auth_plugins/#build-helper-tools) 
of the [Auth Plugin Developer Guide](https://docs.solo.io/gloo/latest/guides/dev/writing_auth_plugins/)). 
With each release Gloo Enterprise publishes the information that you will require to replicate its build environment. 

You can get them by running the following command, where `GLOOE_VERSION` is the desired Gloo Enterprise version, e.g. `1.3.8`.

```bash
GLOOE_VERSION=<target-glooe-version> make get-glooe-info
```

This will download the following files:
- `_glooe/build_env`: values to parameterize the plugin build with;
- `_glooe/dependencies`: the full list of the dependencies used by the `GLOOE_VERSION` version of Gloo Enterprise (
generated by running `go list -m all`);
- `_glooe/verify-plugins-linux-amd64`: a script to verify whether your plugin can be loaded by Gloo Enterprise.

### resolve-deps
The `resolve-deps` target compares and merge the dependencies of your plugin module with the dependencies of the Gloo Enterprise one. 
It will succeed if the shared dependencies match _exactly_ (this is another constraint imposed by Go plugins, more info 
[here](https://docs.solo.io/gloo/latest/guides/dev/writing_auth_plugins/#build-helper-tools)) and fail otherwise, outputting information 
about mismatches to stdout.

### build-plugin
The `build-plugin` target uses the information published by Gloo Enterprise to mirror its build 
environment to compile the plugin and verify compatibility.

#### compile-plugin
The `compile-plugin` target compiles the plugin for the targeted Gloo Enterprise version.

#### verify-plugin
The `verify-plugin` target verifies if the plugin can be loaded by the targeted Gloo Enterprise version.

### build
The `build-plugin` target compiles the plugin inside a docker container using the `Dockerfile` at the root of this 
repository (this is done for reproducibility). It uses the information published by Gloo Enterprise to mirror its build 
environment and verify compatibility.
The `Dockerfile` executes the following targets in the container:
* [`get-glooe-info`](#get-glooe-info) to fetch the build information for the targeted Gloo Enterprise version
* [`resolve-deps`](#resolve-deps) to check if gloo and your plugin have different dependencies
* [`compile-plugin`](#build-plugin) to build the plugin for the targeted Gloo Enterprise version 
* [`verify-plugin`](#build-plugin) to verify if the plugin can be loaded by the targeted Gloo Enterprise version 
