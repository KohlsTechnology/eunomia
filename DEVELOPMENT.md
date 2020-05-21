# Development

## Prerequisites

### Access to a cluster

First of all, you will need access to a Kubernetes or an OpenShift cluster. The easiest way is to start a local minimal cluster in a virtual machine with the help of Minikube or Minishift.

- [VirtualBox](https://www.virtualbox.org/wiki/Downloads) - hypervisor to run a local cluster virtual machine in
- [Minikube](https://kubernetes.io/docs/setup/minikube/) - local Kubernetes cluster
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) - command line tool for controlling Kubernetes cluster
- [Minishift](https://github.com/minishift/minishift/releases) - local OpenShift cluster
- [openshift-cli](https://docs.openshift.com/container-platform/3.11/cli_reference/get_started_cli.html#installing-the-cli) - command line tool for controlling OpenShift cluster

#### Installing on a Mac

Every component can be easily installed via [Homebrew](https://brew.sh/):

```shell
brew cask install virtualbox
brew install minikube
brew install kubectl
brew cask install minishift
brew install openshift-cli
```

#### Installing on a Linux

You can install the applications via a distro-specific package manager, or download a binary, or install from source.

Vendor instructions on how to install on Linux:

- [VirtualBox](https://www.virtualbox.org/wiki/Linux_Downloads)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl-on-linux)
- [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
- [openshift-cli](https://docs.openshift.com/container-platform/3.11/cli_reference/get_started_cli.html#installing-the-cli)
- [Minishift](https://github.com/minishift/minishift/releases)

### Tools to build and test Eunomia

Apart from setting up access to a cluster, you will need some tools in order to build and test Eunomia.

Tools to build:

- [Git](https://git-scm.com/downloads)
- [Go 1.13+](https://golang.org/dl/)
- [Docker](https://docs.docker.com/install/)
- [Operator SDK v0.8.1](https://github.com/operator-framework/operator-sdk/blob/v0.8.1/doc/user/install-operator-sdk.md)

Tools to test:

- [make](https://www.gnu.org/software/make/manual/make.html)
- [shfmt](https://github.com/mvdan/sh)
- [golint](https://github.com/golang/lint)

#### Installing on a Mac

Again, all the components (except operator-sdk v0.8.1 and golint) can be easily installed via Homebrew:

```shell
brew install git
brew install go@1.13
brew install docker
brew install make  # it will be installed as "gmake"; follow the instructions that will appear to use it as "make"
brew install shfmt
```

#### Installing on a Linux

The components (except operator-sdk v0.8.1 and golint) can be installed via a distro-specific package manager. You can also download a binary, or install from source.

Here are vendor instructions on how to install the necessary components on Linux:

- [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
- [Go](https://golang.org/doc/install#tarball)
- [Docker](https://docs.docker.com/install/)
- [make](http://ftp.gnu.org/gnu/make/) (a tarball to download)
- [shfmt](https://github.com/mvdan/sh#shfmt)

#### Installing operator-sdk v0.8.1 from GitHub release

Unfortunately, operator-sdk version 0.8.1 isn't available via package managers, so you have to install it from the GitHub release as described in the [documentation](https://github.com/operator-framework/operator-sdk/blob/v0.8.1/doc/user/install-operator-sdk.md).

#### Installing golint

To install golint, follow the [installation instruction](https://github.com/golang/lint#installation).

### Checkout your fork

To check out this repository:

1. Create a [fork](https://help.github.com/en/articles/fork-a-repo) of this repo
2. Create the directories and clone your fork

```shell
git clone https://github.com/<YOUR_FORK>/eunomia.git
cd eunomia
```

## Running Locally for Development Purposes

The most efficient way to develop the operator locally is run the code on your local machine. This allows you to test code changes as you make them.

To achieve this, run the helper script (preferably in a separate console window as the script blocks):

```shell
./scripts/run-eunomia-locally.sh
```

If you want Eunomia to watch a particular namespace, just pass its name as an argument to the above script. Running the script without the argument defaults to Eunomia watching all namespaces.

### **Note on cluster being watched**

The above script uses `operator-sdk up local` command to run Eunomia operator binary locally. **It enforces the use of a Kubernetes cluster** - Eunomia will be watching namespace(s) in a Kubernetes cluster.

### **Note on template-processor images**

Eunomia binary being run locally still uses [template-processor images](./README.md#Template-Engine) to manage templates and parameters. If you are tweaking the images (anything in `template-processors` directory), make sure to deploy them to your local Minikube Docker registry beforehand. The easiest way is to use the helper script:

```shell
./scripts/deploy-to-local.sh minikube
```

By doing this you'll build your tweaked template-processor images and send them to Minikube Docker registry with the tag `dev`. You must also not forget to use this tag in `templateProcessorImage` in your GitOpsConfig. If not, the images will be downloaded from [Kohl's quay.io](https://quay.io/organization/kohlstechnology).

## Building the Operator Image

The Eunomia operator gets packaged as a container image for running on Kubernetes clusters. These instructions will walk you through building and testing the image in a cluster.

### Building the image on your local workstation

See https://golang.org/doc/install to install/setup your Go Programming environment if you have not already done this.

```shell
GOOS=linux make
```

From here you can build the eunomia-operator Docker image and manually push it to a registry, or run it in a local cluster (see [Using Minikube](#using-minikube) or [Using Minishift](#using-minishift)).

### Building the image and pushing to a remote registry

Run the following to build and push the eunomia-operator image as well as template-processors images:

```shell
export REGISTRY=<your_registry>
docker login $REGISTRY
./scripts/build-images.sh
```

## Testing

If you want to play with Eunomia deployed to Minikube or Minishift, here are some preliminary instructions.

### <a name="using-minikube"></a>Using Minikube

```shell
# Start minikube
minikube start --vm-driver virtualbox

# Build eunomia-operator (building eunomia-operator binary included) and template-processors
# images and store them in minikube docker registry
scripts/deploy-to-local.sh minikube

# Deploy the operator, use your locally-built image
helm template deploy/helm/eunomia-operator/ \
  --set eunomia.operator.image.tag=dev \
  --set eunomia.operator.image.pullPolicy=Never | kubectl apply -f -
```

### <a name="using-minishift"></a>Using Minishift

```shell
# Start minishift
minishift start --vm-driver virtualbox

# Log in to minishift as admin
oc login -u system:admin

# Build eunomia-operator (building eunomia-operator binary included) and template-processors
# images and store them in minishift docker registry
scripts/deploy-to-local.sh minishift

# Deploy the operator, use your locally-built image
helm template deploy/helm/eunomia-operator/ \
  --set eunomia.operator.image.tag=dev \
  --set eunomia.operator.image.pullPolicy=Never \
  --set eunomia.openshift.route.enabled=true | oc apply -f -
```

### After testing configure Docker CLI to use local Docker daemon again

After testing you might want to use your local Docker daemon again. To do it just issue

```shell
eval "$(docker-machine env -u)"
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
# Optional: Set the environment variable $EUNOMIA_TEST_ENV to minishift (default is minikube)
# Optional: Set the environment variable $EUNOMIA_TEST_PAUSE to yes (default is no)
make test-e2e
```

### Running All Tests

```shell
make test
```
