# helm-dump

`helm-dump` is a Helm plugin that extracts Helm Charts from existing resources.

## Installing from sources

To compile and install `helm-dump` from sources, perform the following commands:

```text
# creates the environment variables Helm provides to plugins to
# properly install in the host system
$ eval $(helm env)
$ echo $HELM_PLUGINS
/home/isuttonl/.local/share/helm/plugins

# compiles and install the plugin in HELM_PLUGINS directory
$ make install
mkdir -p ./dist
go build -o ./dist/helm-dump main.go
mkdir -p ./dist/dump
cp ./dist/helm-dump ./dist/dump
cp ./plugin.yaml ./dist/dump
mkdir -p /home/isuttonl/.local/share/helm/plugins/dump
install ./dist/dump/* /home/isuttonl/.local/share/helm/plugins/dump/
```

Once this is finished, the plugin should be available:

```text
$ helm dump
A Helm plugin that remove failed releases revisions from the cluster

Usage:
  helm-dump [command]

Available Commands:
  clean       remove unused artifacts of previous failed releases
  completion  generate the autocompletion script for the specified shell
  help        Help about any command

Flags:
      --config string   config file (default is $HOME/.helm_dump.yaml)
  -h, --help            help for helm-dump

Use "helm-dump [command] --help" for more information about a command.
```

## Building `helm-dump`

To build only the plugin program, use the `build` target:

```text
make build
```

# License

MIT
