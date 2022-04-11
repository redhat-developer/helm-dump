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

# Quickstart

This quickstart showcases the `helm-dump` Helm plugin, which one you to create a Helm chart using as
starting point existing resources from an available Kubernetes cluster.

It contains four parts:

1.  Installing the example deployment
2.  Extracting a Helm Chart
3.  Installing the Helm Chart

# Installing the template deployment

Let's start installing a simple deployment resource in the cluster, for example the following resource:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: default
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
	app: nginx
    spec:
      containers:
	- name: nginx
	  image: nginx:1.14.2
	  ports:
	    - containerPort: 80
```

Now that we have inspected the contents of `nginx-deployment.yaml`, we can install it in the cluster using `kubectl`:

```shell
kubectl apply -f nginx-deployment.yaml
```

```text
deployment.apps/nginx-deployment created
```

Now that we're happy with the deployment, it is fine to label the `Deployment` resource to be picked up by `helm dump init` on the next step:

```shell
kubectl label deployment nginx-deployment helm-dump=please
```

```text
deployment.apps/nginx-deployment labeled
```


# Extracting a Helm Chart

Now that we have a template deployment running, `helm dump init` can be used to extract it to a Helm chart:

```shell
helm dump init -l helm-dump=please my-chart /tmp/helm-dump-init-demo
```

Please note the usage of `-l helm-dump=please`: the `-l` option is equivalent to `kubectl`'s, so refer to `kubectl --help` for more information regarding its usage and semantics.

The <file:///tmp/helm-dump-init-demo/my-chart-0.1.0.tgz> file should be available, so let's inspect its contents.


## `Chart.yaml`

```yaml
apiVersion: v2
name: my-chart
version: 0.1.0
```


## `templates/nginx-deployment_apps_v1.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    deployment.kubernetes.io/revision: "1"
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"apps/v1","kind":"Deployment","metadata":{"annotations":{},"labels":{"app":"nginx"},"name":"nginx-deployment","namespace":"default"},"spec":{"replicas":3,"selector":{"matchLabels":{"app":"nginx"}},"template":{"metadata":{"labels":{"app":"nginx"}},"spec":{"containers":[{"image":"nginx:1.14.2","name":"nginx","ports":[{"containerPort":80}]}]}}}}
  labels:
    app: nginx
    app.kubernetes.io/instance: '{{ $.Release.Name }}'
    app.kubernetes.io/name: '{{ template "my-chart.fullname" $ }}'
    helm-dump: "true"
  name: nginx-deployment-{{ .Release.Name }}
  namespace: default
spec:
  progressDeadlineSeconds: 600
  replicas: 3
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: nginx
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      creationTimestamp: null
      labels:
	app: nginx
    spec:
      containers:
      - image: nginx:1.14.2
	imagePullPolicy: IfNotPresent
	name: nginx
	ports:
	- containerPort: 80
	  protocol: TCP
	resources: {}
	terminationMessagePath: /dev/termination-log
	terminationMessagePolicy: File
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      terminationGracePeriodSeconds: 30
```


## `templates/_helpers.tpl`

```text
{{/*
Expand the name of the chart.
*/}}
{{- define "my-chart.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "my-chart.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "my-chart.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "my-chart.labels" -}}
helm.sh/chart: {{ include "my-chart.chart" . }}
{{ include "my-chart.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "my-chart.selectorLabels" -}}
app.kubernetes.io/name: {{ include "my-chart.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "my-chart.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "my-chart.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
```


# Installing the Chart

Once the Helm Chart has been created, it can be used to install a new release:

```shell
helm install my-app my-chart-0.1.0.tgz
```

```text
NAME: my-app
LAST DEPLOYED: Mon Apr 11 12:33:33 2022
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

Everything going well, we should now see the release installed in the cluster:

```shell
helm list
```

```text
NAME  	NAMESPACE	REVISION	UPDATED                                 	STATUS  	CHART         	APP VERSION
my-app	default  	1       	2022-04-11 12:33:33.854705256 +0200 CEST	deployed	my-chart-0.1.0	           
```

Let's end this section by checking the resources we've created so far:

```shell
kubectl get all
```

```text
NAME                                          READY   STATUS    RESTARTS   AGE
pod/nginx-deployment-9456bbbf9-4jgfc          1/1     Running   0          2m29s
pod/nginx-deployment-9456bbbf9-mfpdc          1/1     Running   0          2m29s
pod/nginx-deployment-9456bbbf9-w5h4f          1/1     Running   0          2m29s
pod/nginx-deployment-my-app-9456bbbf9-4gkzx   1/1     Running   0          24s
pod/nginx-deployment-my-app-9456bbbf9-h2c2n   1/1     Running   0          24s
pod/nginx-deployment-my-app-9456bbbf9-rnmq2   1/1     Running   0          24s

NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
service/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   6d1h

NAME                                      READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nginx-deployment          3/3     3            3           2m29s
deployment.apps/nginx-deployment-my-app   3/3     3            3           24s

NAME                                                DESIRED   CURRENT   READY   AGE
replicaset.apps/nginx-deployment-9456bbbf9          3         3         3       2m29s
replicaset.apps/nginx-deployment-my-app-9456bbbf9   3         3         3       24s
```

As expected, there's one deployment `nginx-deployment` that was extracted by the `helm dump init` operation, and the deployment managed by Helm, `nginx-deployment-my-app`.

## License

Apache License Version 2.0
