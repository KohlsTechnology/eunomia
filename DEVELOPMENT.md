# Development

## Prerequisites

In order to build Eunomia, you'll need the following:

- [Git](https://git-scm.com/downloads)
- [Go 1.12+](https://golang.org/dl/)
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
export JOB_TEMPLATE=./build/job-templates/job.yaml
export CRONJOB_TEMPLATE=./build/job-templates/cronjob.yaml
export WATCH_NAMESPACE=""
export OPERATOR_NAME=eunomia-operator
export GO111MODULE=on
go mod vendor
operator-sdk up local --namespace="${WATCH_NAMESPACE}"
```

Setting WATCH_NAMESPACE to empty string, as above, results in Eunomia watching all namespaces. If you want a particular namespace to be watched, set it explicitly in the env variable.

## Building the Operator Image

The Eunomia operator gets packaged as a container image for running on Kubernetes clusters. These instructions will walk through building and testing the image.

### Building the image on your local workstation

See https://golang.org/doc/install to install/setup your Go Programming environment if you have not already done this.

```shell
export GO111MODULE=on
go mod vendor
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
# Start minikube
minikube start

# Deploy the operator
helm template deploy/helm/eunomia-operator/ | kubectl apply -f -
```

### Using Openshift

Here are some preliminary instructions. This still needs a lot of TLC. Feel free to send in PRs.

```shell
# Deploy the operator
helm template deploy/helm/eunomia-operator/ --set eunomia.openshift.route.enabled=true | oc apply -f -
```

## Run Tests

For testing and CI purposes, we manage several set of tests. These tests can be run locally by following the below instructions. All test scripts assume that you are already logged into your minikube cluster.

### Running Unit Tests
```shell
make test-unit
```

### Running End-to-End Tests
```shell
# Optional: Set the environment variable $EUNOMIA_URI to point to a specific git url for testing
# Optional: Set the environment variable $EUNOMIA_REF to point to a specific git reference for testing
make test-e2e
```

### Running All Tests
```shell
make test
```
