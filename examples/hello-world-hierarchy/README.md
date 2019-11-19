# Simple Hello-World with Hierarchical Variables

This example uses a simple hello world application based on a helm chart with parameters that follow a hierarchical structure. The structure is defined in the file `hierarchy.lst`. In this example variables are first loaded from the default directory, then it merges the empty and level2 folders, and lastly add the variables from the demo folder (which is the same one as specified in parameterSource).

default -> empty -> level2 -> demo

```shell
# Create the namespace
kubectl create namespace eunomia-hello-world-demo

# Create the service account for the runners
kubectl apply -f examples/hello-world-helm/service_account_runner.yaml -n eunomia-hello-world-demo

# Deploy the CR for the hello-world application
kubectl apply -f examples/hello-world-hierarchy/cr/hello-world-cr.yaml -n eunomia-hello-world-demo

# You should also see one completed pod with a name starting with gitopsconfig-hello-world-hierarchy
kubectl get pods -n eunomia-hello-world-demo

# Make sure the hello-world pod start successfully
# You should be seeing 1 running hello-word pod with hiera in the name
kubectl get pods -n eunomia-hello-world-demo-hierarchy

# take a look at the CR
kubectl -n eunomia-hello-world-demo get GitOpsConfig
kubectl -n eunomia-hello-world-demo describe GitOpsConfig hello-world-hierarchy
```

## Cleanup

Simply delete the namespaces to clean up all deployed resources.

```shell
kubectl delete namespace eunomia-hello-world-demo eunomia-hello-world-demo-hierarchy
```
