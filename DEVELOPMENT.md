# Development

## Prerequisites

In order to build Eunomia, you'll need the following:

- [Git](https://git-scm.com/downloads)
- [Go](https://golang.org/dl/)
- [Dep](https://golang.github.io/dep/docs/installation.html)
- [Docker](https://docs.docker.com/install/)
- [Operator SDK v0.8.1](https://github.com/operator-framework/operator-sdk/blob/v0.8.1/doc/user/install-operator-sdk.md)
- Access to a Kubernetes cluster
  - [Minikube](https://kubernetes.io/docs/setup/minikube/) (optional)
  - [Minishift](https://www.okd.io/minishift/) (optional)

### Checkout your fork

To check out this repository:

1. Create a [fork](https://help.github.com/en/articles/fork-a-repo) of this repo
2. Create the directories and clone your fork

```
mkdir -p $GOPATH/src/github.com/KohlsTechnology
cd $GOPATH/src/github.com/KohlsTechnology
git clone https://github.com/<YOUR FORK>/eunomia.git
cd eunomia
```

### Installing on a Mac

All the components can easily be installed via [Homebrew](https://brew.sh/) on a Mac:

```shell
brew install git
brew install go
brew install dep
brew install docker
brew install operator-sdk
brew install minikube
brew install minishift
```

## Running Locally for Development Purposes

The most efficient way to develop the operator locally is run the code on your local machine. This allows you to test code changes as you make them.

```
minikube start
kubectl apply -f ./deploy/crds/eunomia_v1alpha1_gitopsconfig_crd.yaml
export JOB_TEMPLATE=./templates/job.yaml
export CRONJOB_TEMPLATE=./templates/cronjob.yaml
export WATCH_NAMESPACE=""
export OPERATOR_NAME=eunomia-operator
dep ensure
operator-sdk up local
```

## Building the Operator Image

The Eunomia operator gets packaged as a container image for running on Kubernetes clusters. These instructions will walk through building and testing the image.

### Building the image on your local workstation

See https://golang.org/doc/install to install/setup your Go Programming environment if you have not already done this.

Run "dep ensure" before building code. (if this is your first time running this use "dep ensure -v" ,this will take some time to complete.)

```shell
dep ensure
GOOS=linux operator-sdk build eunomia-operator
```

From here you could manually push the image to a registry, or run the image locally (out of scope for this doc).

### Building the image and pushing to a remote registry

Run the following to build and push the images:

```shell
export REGISTRY=<your registry>
docker login $REGISTRY
./scripts/build-images.sh
```

## Testing

### Using Minikube

Here are some preliminary instructions. This still needs a lot of TLC. Feel free to send in PRs.

```shell
minikube start
kubectl create namespace eunomia
kubectl apply -f ./deploy/crds/eunomia_v1alpha1_gitopsconfig_crd.yaml -n eunomia
kubectl delete configmap gitops-templates -n eunomia
kubectl create configmap gitops-templates --from-file=./templates/cronjob.yaml --from-file=./templates/job.yaml -n eunomia
kubectl apply -f ./deploy/kubernetes -n eunomia
```

### Using Openshift

Here are some preliminary instructions. This still needs a lot of TLC. Feel free to send in PRs.

```shell
oc create namespace eunomia
oc apply -f ./deploy/crds/gitops_v1alpha1_gitopsconfig_crd.yaml -n eunomia
oc delete configmap gitops-templates -n eunomia
oc create configmap gitops-templates --from-file=./templates/cronjob.yaml --from-file=./templates/job.yaml -n eunomia
oc apply -f ./deploy/kubernetes -f ./deploy/openshift -n eunomia
```

## Run Tests

For testing and CI purposes, we manage several set of tests. These tests can be run locally by following the below instructions. All test scripts assume that you are already logged into your minikube cluster.

### Running Unit Tests

```shell
./scripts/unit-tests.sh
```

### Running End-to-End Tests

```shell
./scripts/e2e-test.sh
```
