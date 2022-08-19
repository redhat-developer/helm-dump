# helm-dump: a Helm plugin to create a Helm chart from existing cluster resources

![](https://img.shields.io/github/v/release/redhat-developer/helm-dump)
![](https://img.shields.io/github/workflow/status/redhat-developer/helm-dump/release)

## Overview of the helm-dump plugin

Consider a case where as a Helm chart developer, you have configured a workload on either a Kubernetes or an OpenShift cluster. The requirement is to have this workload replicated either in the same or another OpenShift cluster by using a Helm chart. Use the `helm-dump` plug-in to create and export Helm charts based on the resources of a deployed workload available in an OpenShift or Kubernetes cluster. You can run the `helm-dump` plug-in against a namespace and generate a Helm chart for resources in the namespace that match a particular filter.

The `helm-dump` plug-in helps you transfer scalar values such as numbers or strings from a Helm chart template into the chart's `values.yaml` file.

## Quick start

The quickstart showcases the `helm-dump` plug-in. You must install the plug-in and use it to create and export Helm charts. The quickstart consists of the following procedures:

1. Installing the `helm-dump` plug-in
2. Installing the template deployment 
3. Labeling the `Deployment` resource
4. Extracting a Helm chart
5. Installing the newly created Helm chart into a cluster

### Installing the `helm-dump` plug-in

Binary downloads of the `helm-dump` plug-in are available on the Releases page.

#### Prerequisites

You must have GoReleaser to manage the project's build process and compile this software.

#### Procedure

1. Download either the `helm-dump_<VERSION>.tar.gz` or `helm-dump_<VERSION>.zip` file.

2. Unpack its contents in the `$HELM_PLUGINS` directory:
```
# for HELM_PLUGINS environment variable
eval $(helm env)

# unpack the tarball
tar xvfz ~/Downloads/helm-dump_0.2.1.tar.gz -C "$HELM_PLUGINS"
# or the zip file
unzip -d "$HELM_PLUGINS" ~/Downloads/helm-dump_0.2.1.zip
```

3. After the bundle file is unpacked, you should view an output similar to the following: 
```
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
The output verifies that the plugin is now available to use.

4. Build the `helm-dump` binary:
```
$ make build
```

**Example output:**
```
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
5. Bundle the`helm-dump` plug-in:
```
$ make plugin
```
 
**Example output:**
```
Building plugin in /home/isuttonl/Documents/src/helm-dump/dist/plugin/dump... Done!
Creating helm-dump_0.2.1.tar.gz... Done!
Creating helm-dump_0.2.1.zip... Done!
Calculating checksum for plugin bundles... Done!
```

The plug-in bundle has been built and stored at the `./dist/plugin/dump` location and now you can install it in the system.

6. Install the plug-in bundle:
```
$ make install
```
 
**Example output:**
```
Installed plugin: dump
```

7. Verify that the plug-in installation is successful:
```
$ helm plugin list
```
 
**Example output:**
```
NAME    VERSION DESCRIPTION                                                           
dump    0.2.1   A Helm plugin that creates a chart from a cluster's existing resources
```

### Installing the template deployment
This procedure uses the following example resource for the workload configuration in the form of a `nginx-deployment.yaml` file:
```
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

Use this workload configuration to install a template deployment. This workload serves as a deployment that can get replicated in other environments based on the requirements.

#### Procedure

1. Install a simple deployment resource to apply the workload configuration into an available cluster:

For Kubernetes cluster:
```
kubectl apply -f nginx-deployment.yaml
```

For OpenShift cluster:
```
oc apply -f nginx-deployment.yaml
```

**Example output:**
```
deployment.apps/nginx-deployment created
```

2. Verify that the installation is successful:

For Kubernetes cluster:
```
kubectl get pods -n default
```

For OpenShift cluster:
```
oc get pods -n default
```

**Example output:**
```
NAME                    READY   STATUS    RESTARTS   AGE
nginx-deployment-9456bbbf9-8wgjb   1/1     Running   0          9s
nginx-deployment-9456bbbf9-bnmnk   1/1     Running   0          9s
nginx-deployment-9456bbbf9-fmszw   1/1     Running   0          9s
```
The output verifies that all pods are running.

### Labeling the `Deployment` resource

You must label the `Deployment` resource so it can be included in the Helm chart that you are going to create and extract using the `helm-dump` plug-in.

#### Prerequisites

- You must have validated that the workload is properly running by exposing the deployment and the service.
- You must have validated that the workload is indeed working by using the `curl` command. This validation ensures that the `nginx-deployment` web server is successfully installed and working, and you have accessed the web server. 

**Note:**
- If you have exposed the deployment and service successfuly, you must see a route for the `nginx-deployment` service similar to the following example:

**Example:**
```
route.route.openshift.io/nginx exposed
```
- You can label multiple resources, but this quickstart only focuses on the `Deployment` resource.

- Create a chart using helm dump and the labeled `Deployment` resource

After the workload is confirmed to work, the resource can be labeled to be collected by helm-dump at a later stage. 

### Procedure

- Label the `Deployment` resource:
```
kubectl label deployment nginx-deployment helm-dump=please
```
Example output:
```
deployment.apps/nginx-deployment labeled
```
The output verifies that you have labeled the resource successfully.

### Extracting a Helm chart
Now that the `Deployment` resource is installed in the cluster and properly labeled, you can now extract this resource bundled in a Helm chart.

**Procedure**

1. Use the `helm dump init` command to extract the `Deployment` resource to a Helm chart:
```
helm dump init -l helm-dump=please my-chart /tmp/helm-dump-init-demo
```
**Note:**
- When you use `-l helm-dump=please`, the `-l` option has the same semantics as the `kubectl` option, so refer to `kubectl --help` for more information regarding its usage and semantics.

After extraction, the `my-chart-0.1.0.tgz` file is available at the `/tmp/helm-dump-init-demo/` directory with the following chart and resource templates as its contents:

- Chart.yaml:
```yaml
apiVersion: v2
name: my-chart # (1)
version: 0.1.0
```
1. Specifies the chart name as informed by the user when generating a chart using the `helm-dump` plug-in.
 
- templates/nginx-deployment_apps_v1.yaml:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    deployment.kubernetes.io/revision: "1"
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"apps/v1","kind":"Deployment","metadata":{"annotations":{},"labels":{"app":"nginx"},"name":"nginx-deployment","namespace":"default"},"spec":{"replicas":3,"selector":{"matchLabels":{"app":"nginx"}},"template":{"metadata":{"labels":{"app":"nginx"}},"spec":{"containers":[{"image":"nginx:1.14.2","name":"nginx","ports":[{"containerPort":80}]}]}}}}
  labels: # (1)
    app: nginx 
    app.kubernetes.io/instance: '{{ $.Release.Name }}'
    app.kubernetes.io/name: '{{ template "my-chart.fullname" $ }}'
    helm-dump: "please"
  name: nginx-deployment-{{ .Release.Name }} # (2)
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
1. Specifies the labels added by the `helm-dump` plug-in according to the Helm guidelines.
2. Specifies the name field that is modified according to the Helm guidelines to include the release name at the time of deployment.

### Installing the newly created Helm chart into a cluster

After you create and extract the Helm chart, now proceed to install the modified Helm chart into the cluster. 

#### Procedure
1. Use the Helm chart to install a new release:
```
helm install my-app my-chart-0.1.0.tgz
```

**Example output:**
```
NAME: my-app
LAST DEPLOYED: Mon Apr 11 12:33:33 2022
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

2. Verify that the release is successfully installed in the cluster:
```
helm list
```

**Example output:**
```
NAME    NAMESPACE   REVISION    UPDATED                                     STATUS      CHART           APP VERSION
my-app  default     1           2022-04-11 12:33:33.854705256 +0200 CEST    deployed    my-chart-0.1.0
```

3. Verify the resources declared in the cluster:
```
kubectl get all
```

**Example output:**
```
NAME                                          READY   STATUS    RESTARTS   AGE
pod/nginx-deployment-9456bbbf9-4jgfc          1/1     Running   0          2m29s (1)
pod/nginx-deployment-9456bbbf9-mfpdc          1/1     Running   0          2m29s (1)
pod/nginx-deployment-9456bbbf9-w5h4f          1/1     Running   0          2m29s (1)
pod/nginx-deployment-my-app-9456bbbf9-4gkzx   1/1     Running   0          24s
 
pod/nginx-deployment-my-app-9456bbbf9-h2c2n   1/1     Running   0          24s
 
pod/nginx-deployment-my-app-9456bbbf9-rnmq2   1/1     Running   0          24s
 
NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
service/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   6d1h
 
NAME                                      READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nginx-deployment          3/3     3            3           2m29s (2)
deployment.apps/nginx-deployment-my-app   3/3     3            3           24s   (3)
 
NAME                                                DESIRED   CURRENT   READY   AGE
replicaset.apps/nginx-deployment-9456bbbf9          3         3         3       2m29s
replicaset.apps/nginx-deployment-my-app-9456bbbf9   3         3         3       24s
```
1. Replicas from the resource that created the chart.
2. The resource that created the chart.
3. Resource created with the recently created Helm chart.

The output verifies that there is `nginx-deployment` deployment resource with three replicas that served as template, and `nginx-deployment-my-app` deployment resource with three replicas installed using the newly generated Helm chart.

The `nginx-deployment` deployment resource is extracted by the `helm dump init` command and the `nginx-deployment-my-app` deployment is managed by Helm.

## License
Apache License Version 2.0