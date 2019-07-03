# Examples

Here are some basic examples for you to try out and get a feel for what Eunomia can do for you. As always...the sky is the limit!

## Install the Operator

Before you can execute any of the examples, you need to install the operator first.

```shell
# Create the CRDs
kubectl apply -f ./deploy/crds/eunomia_v1alpha1_gitopsconfig_crd.yaml

# Create the namespace
kubectl create namespace eunomia-operator

# Generate the configmap with the details for the runners
kubectl create configmap eunomia-templates --from-file=./templates/cronjob.yaml --from-file=./templates/job.yaml -n eunomia-operator

# Deploy the operator
kubectl apply -f ./deploy/kubernetes -n eunomia-operator

# Make sure the operator pod is running
kubectl get pods -n eunomia-operator

# Once it is running, check the logs to make sure there are no errors
kubectl -n eunomia-operator logs `kubectl get pods -n eunomia-operator -o name | sed 's/pod\///g'`
```

## Simple hello-world

[Static yaml file](hello-world-yaml/README.md) 

[Helm chart with parameters](hello-world-helm/README.md) 

## Real world-ish

[Boostrapping a whole cluster via a cluster-seed](cluster/README.md)
