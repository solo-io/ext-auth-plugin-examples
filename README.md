<h1 align="center">
    <img src="https://github.com/solo-io/ext-auth-plugin-examples/raw/master/img/gloo-plugin.png" alt="Gloo Plugins" width="440" height="309">
  <br>
  External auth plugin examples
</h1>

This repository contains example implementations of the 
[ExtAuthPlugin interface](https://github.com/solo-io/ext-auth-plugins/blob/master/api/interface.go) and a set 
of utilities that you can (and should!) use when building your own plugins.

Please refer to the [Auth Plugin Developer Guide](https://gloo.solo.io/dev/writing_auth_plugins/) for an in-depth 
explanation on how you can use this repository as a template to write your own Gloo Auth plugins.

## Makefile overview
Following is an overview of the most relevant `make` targets.

### get-glooe-info
When you are writing your own Ext Auth plugins, you must target a specific GlooE version. This is because of the 
nature of Go plugins (you can find more info in [this section](https://gloo.solo.io/dev/writing_auth_plugins/#build-helper-tools) 
of the [Auth Plugin Developer Guide](https://gloo.solo.io/dev/writing_auth_plugins/)). With each release GlooE publishes 
the information that you will require to replicate its build environment. You can get them by running

```bash
GLOOE_VERSION=<target-glooe-version> make get-glooe-info
```

where `GLOOE_VERSION` is the desired GlooE version, e.g. `0.20.6`.

This will download the following files:
- `_glooe/build_env`: values to parameterize the plugin build with
- `_glooe/Gopkg.lock`: the [dep .lock file](https://golang.github.io/dep/docs/Gopkg.lock.html) containing all GlooE 
dependency version
- `_glooe/verify-plugins-linux-amd64`: a script to verify whether your plugin can be loaded by GlooE

### compare-deps
The `compare-deps` target compares the local `Gopkg.lock` with the one describing the GlooE dependencies. It will succeed 
if the shared dependencies match _exactly_ (this is another constraint imposed by Go plugins, more info 
[here](https://gloo.solo.io/dev/writing_auth_plugins/#build-helper-tools)) and fail otherwise, outputting information 
about mismatches to stdout and a file.

### build-plugins
The `build-plugins` target compiles the plugin inside a docker container using the `Dockerfile` at the root of this 
repository (this is done for reproducibility). It uses the information published by GlooE to mirror its build 
environment and verify compatibility.

## Get example images
You can get the images for the example plugin(s) whose source code is contained in this repository by running:

```bash
docker pull quay.io/solo-io/ext-auth-plugins:<glooe_version>
```

where the tag `glooe_version` is the version of GlooE you want to run the plugins with, e.g. `0.20.6`.

## Publishing your own plugins
To publish your own images you can just tag the image built in the `build-plugins` target (by adding add a `-t` option) 
and publish it to a docker registry that is reachable from the cluster you are running GlooE in.

# Example workflow

Note: these instructions work on gloo-e <= 1.0.0-rc2
In gloo-e 1.0.0-rc3, we transitioned to using go modules for dependency managment. We will update 
these instructions soon for versions of gloo-e >= 1.0.0-rc3.

First, set your gloo-e version to an environment variable:
```
export GLOOE_VERSION=1.0.0-rc2
```

Run `compare-deps` to check if gloo and your plugin have different dependencies:

```
make GLOOE_VERSION=$GLOOE_VERSION compare-deps
```

If there are any mismatched dependencies, A file named `overrides.toml` will be created. Please 
reconcile the contents of this file with your `Gopkg.toml`. This means replacing existing entries,
and adding missing entries.

Once the `Gopkg.toml` is up to date, run `dep ensure` to update your vendor folder:

```
dep ensure
```

Build the plugins using the `build-plugins` make target.
It is important to build the plugins using the Makefile as this will build them in a docker container
in such a way that they are compatible with gloo. This step will also verify that they can be loaded.

```
make GLOOE_VERSION=$GLOOE_VERSION build-plugins
```

This will build an anonymous container with the plugin for you. The output will look like so:
```
...
 ---> 4846ae5e0a4d
Step 15/16 : COPY --from=build-env /go/src/github.com/solo-io/ext-auth-plugin-examples/plugins/RequiredHeader.so /compiled-auth-plugins/
 ---> 1932dbdca716
Step 16/16 : CMD cp /compiled-auth-plugins/* /auth-plugins/
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

And push it:

```
docker push username/dockerrepo:v1
```

You can now use this plugin as an init container for gloo's extauth module.

## Common Errors

If you see this error in the json log:
```
{"level":"error","ts":"2019-12-17T20:59:17.301Z","logger":"verify-plugins","caller":"scripts/verify_plugins.go:54","msg":"Plugin(s) cannot be loaded by Gloo","error":"failed to load plugin: failed to open plugin file: plugin.Open(\"plugins/RequiredHeader\"): plugin was built with a different version of package github.com/golang/protobuf/proto","errorVerbose":"failed to load plugin:\n    github.com/solo-io/go-utils/errors.Wrapf\n        /go/src/github.com/solo-io/go-utils/errors/utils.go:12\n  - failed to open plugin file:\n    github.com/solo-io/go-utils/errors.Wrapf\n        /go/src/github.com/solo-io/go-utils/errors/utils.go:12\n  - plugin.Open(\"plugins/RequiredHeader\"): plugin was built with a different version of package github.com/golang/protobuf/proto","stacktrace":"main.main\n\t/go/src/github.com/solo-io/solo-projects/projects/extauth/scripts/verify_plugins.go:54\nruntime.main\n\t/usr/local/go/src/runtime/proc.go:200"}
```

Make sure that the Gopkg.toml has the right deps from the `overrides.toml` file and that you have run `dep ensure`
