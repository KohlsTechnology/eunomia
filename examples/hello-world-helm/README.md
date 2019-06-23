# Simple Hello-World HELM

This example uses a simple hello world application based on a helm chart with parameters.

```shell
# Create the namespace
kubectl create namespace eunomia-hello-world-demo

# Create the service account for the runners
kubectl apply -f examples/service_account_runner.yaml -n eunomia-hello-world-demo

# Deploy the CR for the hello-world application
kubectl apply -f examples/hello-world-helm/cr/hello-world-cr1.yaml -n eunomia-hello-world-demo

# Make sure the hello-world pod start successfully
# You should be seeing 1 running hello-word pod
# You should also see one completed pod with a name starting with gitopsconfig-hello-world-helm
kubectl get pods -n eunomia-hello-world-demo

# take a look at the CR
kubectl -n eunomia-hello-world-demo get GitOpsConfig
kubectl -n eunomia-hello-world-demo describe GitOpsConfig hello-world-helm

# Lets simulate a change in Git
# Your load increased and you now need 3 running pods
# So we are going to change `replicas: 1` to `replicas: 3`
# We will simply point at an updated version of the CR for the demo
# In the real world this would a pull request to the existing yaml file
kubectl apply -f examples/hello-world-helm/cr/hello-world-cr2.yaml -n eunomia-hello-world-demo

# You should now be seeing 3 running hello-word pods
kubectl get pods -n eunomia-hello-world-demo

# Lets simulate another change in Git
# You updated your application and need to deploy a newer version of it
# So we are going to change the image tag from `1.0` to `2.0`
# We will simply point at an updated version of the CR for the demo
# In the real world this would a pull request to the existing yaml file
kubectl apply -f examples/hello-world-helm/cr/hello-world-cr3.yaml -n eunomia-hello-world-demo

# You should now be see the old pods being replaced by 3 new ones
kubectl get pods -n eunomia-hello-world-demo

# Access the service
minikube service hello-world -n eunomia-hello-world-demo
```

## Cleanup

Simply delete the namespace to clean up all deployed resources.

```shell
kubectl delete namespace eunomia-hello-world-demo
```
