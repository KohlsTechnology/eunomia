# Troubleshooting guide

## Resources are not getting deleted

This happens sometimes and is usually related to the finalizers.

```shell
# Find the CR causing the problem
kubectl api-resources --verbs=list --namespaced -o name | xargs -n 1 kubectl get --show-kind --ignore-not-found -n <NAMESPACE>

# Patch the finalizer to be able to nuke it
kubectl patch <RESOURCE> -p '{"metadata":{"finalizers": []}}' --type=merge -n <NAMESPACE>
```

#### Example

```
kubectl api-resources --verbs=list --namespaced -o name | xargs -n 1 kubectl get --show-kind --ignore-not-found -n eunomia-hello-world-demo

NAME                                             AGE
gitopsconfig.eunomia.kohls.io/hello-world-yaml   1h

kubectl patch gitopsconfig.eunomia.kohls.io/hello-world-yaml -p '{"metadata":{"finalizers": []}}' --type=merge -n eunomia-hello-world-demo
```

## CRD can't be deleted

More fun with the finalizer most likely. Use the below commands to clean it up.

```shell
# remove the CRD finalizer blocking on custom resource cleanup
kubectl patch crd/gitopsconfigs.eunomia.kohls.io -p '{"metadata":{"finalizers":[]}}' --type=merge

# now delete it
kubectl delete -f ./deploy/crds/eunomia.kohls.io_gitopsconfigs_crd-k8s-pre-1.16.yaml
```
