# Examples

Here are some basic examples for you to try out and get a feel for what Eunomia can do for you. As always...the sky is the limit!

## Simple hello-world

### Install the Operator

Before you can execute any of the hello-world examples, you need to install the operator first.

```shell
# Deploy the operator pre-requisites, which require cluster-admin access
helm template deploy/helm/prereqs/ | kubectl apply -f -

# Deploy the operator
helm template deploy/helm/operator/ | kubectl apply -f -

# Make sure the operator pod is running
kubectl get pods -n eunomia-operator

# Once it is running, check the logs to make sure there are no errors
kubectl -n eunomia-operator logs deployment/eunomia-operator
```
## Try it out

[Static yaml file](hello-world-yaml) 

[Helm chart with parameters](hello-world-helm) 

## Real world-ish

[Boostrapping a whole cluster via a cluster-seed](cluster/README.md)
