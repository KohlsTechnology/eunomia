# Getting started

In this document you will find a step-by-step guide on how to get Eunomia running in a local Minikube cluster.

Eunomia is watching GitOpsConfig Custom Resources. They and their handling by Eunomia will be discussed in general.

You will run a simple example where Eunomia will be acting upon local GitOpsConfig changes.

## Prerequisites

In order to setup Eunomia, you'll need access to a Kubernetes cluster:
- [VirtualBox](https://www.virtualbox.org/wiki/Downloads) - hypervisor to run the Minikube virtual machine in
- [Minikube](https://kubernetes.io/docs/setup/minikube/) - local cluster where you will deploy Eunomia
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) - command line tool for controlling Kubernetes cluster

### Installing on a Mac

```shell
brew cask install virtualbox
brew install minikube
brew install kubectl
```

### Installing on a Linux

All the components can be installed via a distro-specific package manager. If it's not possible, you can download a binary, or install from source.

Here are vendor instructions on how to install all the necessary components on Linux:

- [VirtualBox](https://www.virtualbox.org/wiki/Linux_Downloads)
- [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl-on-linux)

## Deploy Eunomia to a Minikube cluster

### Start a local Minikube cluster

Minikube is a minimal Kubernetes cluster run in a virtual machine (here in VirtualBox).

```shell
minikube start --vm-driver virtualbox
```

From now on your local Kubernetes client (kubectl) is configured to use your just started Minikube cluster. If any problems should occur, you might want to take a look at the [Minikube quickstart guide](https://kubernetes.io/docs/setup/learning-environment/minikube/#quickstart).

### Install Eunomia from OperatorHub

Eunomia can be easilly installed from OperatorHub. Just go to [Eunomia on OperatorHub](https://operatorhub.io/operator/eunomia) and click the blue `Install` button on the right. A pop-up will appear - follow the instructions there.

After performing this step Eunomia operator should be running in the `operators` namespace in your Kubernetes cluster.

## Hello-world example

We'll be using a very simple yaml files example which you can find in [`examples/hello-world-yaml/`](./examples/hello-world-yaml#simple-hello-world-with-plain-yaml-files) directory. Additionally, we'll discuss every single step performed to make sure that everything regarding Eunomia operation is clear.

### Create a new namespace

First, we need to create a namespace for our resources to be deployed in. This is for the sake of separation and keeping order:

```shell
kubectl create namespace eunomia-hello-world-yaml-demo
```

Eunomia by default watches all namespaces in the cluster for GitOpsConfigs.

### Create a new service account

We need the service account to provide access rights for all pods to be created without problems. The service account's name is `eunomia-runner-yaml`. It will be bound to `ClusterRole admin` to not complicate the example too much.

The service account will be taken from [`examples/hello-world-yaml/eunomia-runner-sa.yaml`](./examples/hello-world-yaml/eunomia-runner-sa.yaml).

```shell
kubectl apply -f https://raw.githubusercontent.com/KohlsTechnology/eunomia/master/examples/hello-world-yaml/eunomia-runner-sa.yaml -n eunomia-hello-world-yaml-demo
```

### Deploy the GitOpsConfig Custom Resource for the hello-world application

Eunomia is a Kubernetes native application ([an operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)). Its [Custom Resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) being watched for is a GitOpsConfig.

The one from [`examples/hello-world-yaml/cr/hello-world-cr1.yaml`](./examples/hello-world-yaml/cr/hello-world-cr1.yaml) will be applied.

```yaml
apiVersion: eunomia.kohls.io/v1alpha1
kind: GitOpsConfig
metadata:
  name: hello-world-yaml
spec:
  templateSource:
    uri: https://github.com/KohlsTechnology/eunomia
    ref: master
    contextDir: examples/hello-world-yaml/template1
  parameterSource:
    ref: master
    contextDir: examples/hello-world-yaml/parameters
  triggers:
  - type: Change
  serviceAccountRef: eunomia-runner-yaml
  templateProcessorImage: quay.io/kohlstechnology/eunomia-base:latest
  resourceHandlingMode: Apply
  resourceDeletionMode: Delete
```

Things to pay attention to:

|  field name  |  description  |
|:---|:---|
|  `kind`  |  set to `GitOpsConfig` - no Kubernetes application but Eunomia will be caring for it  |
|  `templateSource` section  |  it defines the source of the templates  |
|  - `uri`  |  repository where the templates are taken from; **NOTE:** This `uri` can be any repository with plain, or Helm, or Jinja Kubernetes manifest yaml files (depends on the `templateProcessorImage` field), but for the sake of simplicity let's use the Eunomia repository in this example  |
|  - `ref`  |  branch on the remote repository  |
|  - `contextDir`  |  directory where templates files will be taken from; in our example it's `examples/hello-world-yaml/template1/` - you can find there a single manifest file [`hello-world.yaml`](./examples/hello-world-yaml/template1/hello-world.yaml) with a Deployment and a Service inside  |
|  `parameterSource` section  |  it defines the source of the parameters for the templates specified in `templateSource` section  |
|  - `uri`  |  we didn't specify it - so by default it will be copied from the `templateSource` section  |
|  - `ref`  |  branch on the remote repository  |
|  - `contextDir`  |  directory where parameter files will be taken from; in our example `examples/hello-world-yaml/parameters/` - it's rather empty, and that's because in the example we don't parametrize anything (everything is set to a fixed value in the templates directory already)  |
|  `triggers`  |  a list of triggers which will cause Eunomia to take an action on GitOpsConfig-owned resources (start a reconciliation cycle, formally speaking). We specified only a `Change` trigger here - it will trigger every time a GitOpsConfig is changed (creation also counts)  |
|  `serviceAccountRef`  |  remember the service account we have deployed earlier? It will be used to [authenticate container processes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)  |
|  `templateProcessorImage`  |  the Docker image which will be used to process templates and parameters  |

### Deploy the first version of GitOpsConfig

```shell
kubectl apply -f https://raw.githubusercontent.com/KohlsTechnology/eunomia/master/examples/hello-world-yaml/cr/hello-world-cr1.yaml -n eunomia-hello-world-yaml-demo
```

After a while check if the application has been deployed successfully:

```shell
kubectl get pods -n eunomia-hello-world-yaml-demo
```

If everything went right, you should see output similar to this after a while:

```shell
NAME                                         READY   STATUS      RESTARTS   AGE
gitopsconfig-hello-world-yaml-ua5y7o-fs7k9   0/1     Completed   0          60s
hello-world-7b7f58765c-4wtnn                 1/1     Running     0          29s
```

As you can see, we have only one hello-world pod and that's because we've deployed a GitOpsConfig which points to `examples/hello-world-yaml/template1`. This directory defines a [Deployment with exactly one replica defined](./examples/hello-world-yaml/template1/hello-world.yaml).

To see the application deployed, you can run:

```shell
minikube service hello-world -n eunomia-hello-world-yaml-demo
```

On the newly opened website notice the `Version: 1.0.0`.

### Deploy the second version of GitOpsConfig

Now we can apply a new GitOpsConfig (by performing this action we will change the original GitOpsConfig). The updated one ([`examples/hello-world-yaml/cr/hello-world-cr2.yaml`](./examples/hello-world-yaml/cr/hello-world-cr2.yaml)) differs from the first one by pointing to template dir `examples/hello-world-yaml/template2` where in a Deployment [there are three replicas defined](./examples/hello-world-yaml/template2/hello-world.yaml).

```shell
kubectl apply -f https://raw.githubusercontent.com/KohlsTechnology/eunomia/master/examples/hello-world-yaml/cr/hello-world-cr2.yaml -n eunomia-hello-world-yaml-demo
```

Eunomia detected our change. It will now start a new reconciliation cycle to adjust the state of the cluster to match the updated GitOpsConfig.

We can check if all three replicas have been deployed:

```shell
kubectl get pods -n eunomia-hello-world-yaml-demo
```

Output should look like this:

```shell
NAME                                         READY   STATUS      RESTARTS   AGE
gitopsconfig-hello-world-yaml-1nhgz3-5t7d7   0/1     Completed   0          22s
gitopsconfig-hello-world-yaml-ua5y7o-fs7k9   0/1     Completed   0          6m4s
hello-world-7b7f58765c-4wtnn                 1/1     Running     0          5m33s
hello-world-7b7f58765c-7gn4d                 1/1     Running     0          7s
hello-world-7b7f58765c-wf47l                 1/1     Running     0          7s
```

Eunomia has reconciled the GitOpsConfig because you uploaded its new version (this triggered a new reconciliation cycle). Eunomia compared current status of the cluster with the specification and then adjusted resources to reflect the latter. After this operation the current status matches the specification.

### Deploy the third version of GitOpsConfig

```shell
kubectl apply -f https://raw.githubusercontent.com/KohlsTechnology/eunomia/master/examples/hello-world-yaml/cr/hello-world-cr3.yaml -n eunomia-hello-world-yaml-demo
```

We can check our test namespace:

```shell
kubectl get pods -n eunomia-hello-world-yaml-demo
```

Three new pods should supersede the old ones:

```shell
NAME                                         READY   STATUS      RESTARTS   AGE
gitopsconfig-hello-world-yaml-1nhgz3-5t7d7   0/1     Completed   0          6m49s
gitopsconfig-hello-world-yaml-6qqkns-kpt6m   0/1     Completed   0          46s
gitopsconfig-hello-world-yaml-ua5y7o-fs7k9   0/1     Completed   0          12m
hello-world-5df998b7bb-6zw99                 1/1     Running     0          33s
hello-world-5df998b7bb-744w4                 1/1     Running     0          29s
hello-world-5df998b7bb-r4zz9                 1/1     Running     0          31s
```

You should be able to see the `Version: 2.0.0` now:

```shell
minikube service hello-world -n eunomia-hello-world-yaml-demo
```

### Clean-up

Simply delete the test namespace, to remove all the resources.

```shell
kubectl delete ns eunomia-hello-world-yaml-demo
```
