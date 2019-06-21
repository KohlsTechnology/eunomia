# Examples

## Simple Hello-World

```shell
kubectl create namespace eunomia-hello-world-demo
kubectl apply -f examples/templates/hello-world.yaml -n eunomia-hello-world-demo
minikube service -n eunomia-hello-world-demo hello-world
```

