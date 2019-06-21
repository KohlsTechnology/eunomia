# Examples

## Simple Hello-World

### Test everything works without Eunomia
```shell
kubectl create namespace hello-world-demo
kubectl apply -f examples/templates/hello-world.yaml -n hello-world-demo
minikube service -n hello-world-demo hello-world
```

### Cleanup
```shell
kubectl delete namespace hello-world-demo
```

### Let's do this with Eunomia
```shell
# nuke the namespace if it already exists
kubectl delete namespace eunomia-hello-world-demo

kubectl create namespace eunomia-hello-world-demo

kubectl apply -f ./deploy/crds/eunomia_v1alpha1_gitopsconfig_crd.yaml -n eunomia-hello-world-demo
kubectl create configmap eunomia-templates --from-file=./templates/cronjob.yaml --from-file=./templates/job.yaml -n eunomia-hello-world-demo
kubectl apply -f ./deploy/kubernetes -n eunomia-hello-world-demo
kubectl get pods -n eunomia-hello-world-demo

kubectl apply -f examples/hello-world-yaml/hello-world-cr.yaml -n eunomia-hello-world-demo

# take a look at the CR
kubectl -n eunomia-hello-world-demo get GitOpsConfig
kubectl -n eunomia-hello-world-demo describe GitOpsConfig hello-world-yaml

# Make sure the hello-world pod gets started
kubectl get pods -n eunomia-hello-world-demo -w

# Access the service
minikube service hello-world -n eunomia-hello-world-demo
```
