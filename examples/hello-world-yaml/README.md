# Simple hello-world with plain yaml files

This example uses a simple hello world application based on a static yaml file. Scenario for this example is following: 
- Eunomia is already initialized in your cluster (ex.: for minikube please look [here](../../DEVELOPMENT.md#using-minikube))
- create new namespace
- create service account in this namespace
- apply first version of GitOpsConfig CR and observe 1 pod of hello-app created
- apply second version of GitOpsConfig CR and observe 2 more pods of hello-app spawned
- apply third version of GitOpsConfig CR and observe all 3 pods of hello-app being updated to version 2.0
- remove whole namespace with all created elements

## Example
```shell
# Create new namespace
kubectl create namespace eunomia-hello-world-yaml-demo

# Create new service account for the runners
kubectl apply -f examples/hello-world-yaml/eunomia-runner-sa.yaml -n eunomia-hello-world-yaml-demo

# Deploy the CR for the hello-world application
kubectl apply -f examples/hello-world-yaml/cr/hello-world-cr1.yaml -n eunomia-hello-world-yaml-demo

# Make sure the hello-world pod start successfully
# You should be seeing 1 running hello-word pod
# You should also see one completed pod with a name starting with gitopsconfig-hello-world-yaml
kubectl get pods -n eunomia-hello-world-yaml-demo

#If you are using minikube you can run:
minikube service hello-world -n eunomia-hello-world-yaml-demo

# take a look at the CR
kubectl -n eunomia-hello-world-yaml-demo get GitOpsConfig
kubectl -n eunomia-hello-world-yaml-demo describe GitOpsConfig hello-world-yaml

# Lets simulate a change in Git
# Your load increased and you now need 3 running pods
# So we are going to change `replicas: 1` to `replicas: 3`
# We will simply point at an updated version of the CR for the demo
# In the real world this would a pull request to the existing yaml file
kubectl apply -f examples/hello-world-yaml/cr/hello-world-cr2.yaml -n eunomia-hello-world-yaml-demo

# You should now be seeing 3 running hello-word pods
kubectl get pods -n eunomia-hello-world-yaml-demo

# Lets simulate another change in Git
# You updated your application and need to deploy a newer version of it
# So we are going to change the image tag from `1.0` to `2.0`
# We will simply point at an updated version of the CR for the demo
# In the real world this would a pull request to the existing yaml file
kubectl apply -f examples/hello-world-yaml/cr/hello-world-cr3.yaml -n eunomia-hello-world-yaml-demo

# You should now be see the old pods being replaced by 3 new ones
kubectl get pods -n eunomia-hello-world-yaml-demo

#If you are using minikube you can run to see updated version:
minikube service hello-world -n eunomia-hello-world-yaml-demo
```

## Cleanup
```shell
kubectl delete namespace eunomia-hello-world-yaml-demo
```
