# helm-dump: a Helm plugin to create a Helm chart from existing resources


![](https://img.shields.io/github/v/release/redhat-developer/helm-dump)
![](https://img.shields.io/github/workflow/status/redhat-developer/helm-dump/release)


The **helm-dump** [Helm](https://helm.sh) plugin allows you to create a Helm chart 
using as starting point existing resources from an available Kubernetes cluster. 

The project at this point is empty, while it is being configured.

## Install

Binary downloads of the plugin can be found [on the Releases page](https://github.com/redhat-developer/helm-dump/releases/latest)

Download either `helm-dump_<VERSION>.tar.gz` or `helm-dump_<VERSION>.zip` and unpack its contents in the `$HELM_PLUGINS` directory:

```shell
# for HELM_PLUGINS environment variable
eval $(helm env)

# unpack the tarball
tar xvfz ~/Downloads/helm-dump_0.2.1.tar.gz -C "$HELM_PLUGINS"
# or the zip file
unzip -d "$HELM_PLUGINS" ~/Downloads/helm-dump_0.2.1.zip
```

Once the bundle file is unpacked, the plugin should be available to use:

```text
$ helm dump 
A Helm plugin that creates a chart from a cluster's existing resources

Usage:
  helm-dump [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  version     print the helm-dump plugin version information

Flags:
      --config string   config file (default is $HOME/.helm_dump.yaml)
  -h, --help            help for helm-dump

Use "helm-dump [command] --help" for more information about a command.
```

## Building from sources

[GoReleaser](https://github.com/goreleaser/goreleaser/) is used to manage the project's build process, so it is required 
to compile this software and is considered a pre-requisite.

### Building the `helm-dump` binary

The `build` target is used to build the `helm-dump` binary:

```text
$ make build
   • building...      
   • loading config file       file=.goreleaser.yaml
   • loading environment variables
   • getting and validating git state
      • building...               commit=eb2f496ae94abcf135f0df3d6d05c258c1df8bde latest tag=v0.2.1
      • pipe skipped              error=validation is disabled
...
   • building binaries
      • building                  binary=dist/helm-dump_windows_arm64/helm-dump.exe
      • building                  binary=dist/helm-dump_linux_arm64/helm-dump
      • building                  binary=dist/helm-dump_linux_s390x/helm-dump
      • building                  binary=dist/helm-dump_linux_arm_6/helm-dump
      • building                  binary=dist/helm-dump_linux_386/helm-dump
      • building                  binary=dist/helm-dump_darwin_amd64/helm-dump
      • building                  binary=dist/helm-dump_darwin_arm64/helm-dump
      • building                  binary=dist/helm-dump_linux_ppc64le/helm-dump
      • building                  binary=dist/helm-dump_linux_amd64/helm-dump
      • building                  binary=dist/helm-dump_windows_amd64/helm-dump.exe
   • storing release metadata
      • writing                   file=dist/artifacts.json
      • writing                   file=dist/metadata.json
   • build succeeded after 1.09s
```

### Packaging the `helm-dump` plugin

The `plugin` target is used to bundle the `helm-dump` plugin:

```text
$ make plugin
Building plugin in /home/isuttonl/Documents/src/helm-dump/dist/plugin/dump... Done!
Creating helm-dump_0.2.1.tar.gz... Done!
Creating helm-dump_0.2.1.zip... Done!
Calculating checksum for plugin bundles... Done!
```

### Installing the `helm-dump` plugin

The `install` target installs the plugin bundle built in `./dist/plugin/dump` after `make plugin`:

```text
$ make install
Installed plugin: dump
$ helm plugin list
NAME    VERSION DESCRIPTION                                                           
dump    0.2.1   A Helm plugin that creates a chart from a cluster's existing resources
```
## License

Apache License Version 2.0
