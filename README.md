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
