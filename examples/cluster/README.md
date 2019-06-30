# Cluster Example

This example reflects a more real-world scenario. A "cluster seed" CR is used, which then:
- We trigger the initial install via command line, afterwards even the cluster-seed will get managed throuh Eunomia. No more manual activities! GitOps all the way!
- Provisions the "cluster seed" with a service account that has cluster-admin rights
- Provisions required components to run the cluster
- Creates GitOpsConfig CRs for teams to manage their own namespaces, with only access in their namespaces

# Repo structure
In order to keep things simple for this example, we're going to use the same git repo, but a folder structure under it. In the real world, you would break out the various team folders into at least one repo each. How exactly this would look like depends your requirements and organizational structure.

```shell
# Create the CRD
kubectl apply -f ./deploy/crds/eunomia_v1alpha1_gitopsconfig_crd.yaml

# Create the namespace
kubectl create namespace eunomia-operator

# Generate the configmap with the details for the runners
kubectl create configmap eunomia-templates --from-file=./templates/cronjob.yaml --from-file=./templates/job.yaml -n eunomia-operator

# Create the namespace for the cluster-seed
kubectl create namespace eunomia-cluster-seed

# Initial configuration of the cluster seed
helm template -f examples/cluster/teams/platform/cluster-seed/parameters/values.yaml examples/cluster/teams/platform/cluster-seed/templates/ | kubectl apply -f -

# Deploy the operator
helm template -f examples/cluster/teams/platform/eunomia-operator/parameters/values.yaml examples/cluster/teams/platform/eunomia-operator/templates/ | kubectl apply -f -

```

At this point the cluster should be "magically" configuring itself and within a few minutes all resources should be available.

```shell
# Watch the magic happening
kubectl get pods --all-namespaces -w 
```
