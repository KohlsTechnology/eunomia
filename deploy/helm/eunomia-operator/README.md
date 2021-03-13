# Eunomia GitOps Operator for Kubernetes

<!---
This doesn't work yet. Maybe someday it will.

## TL;DR:

```shell
helm repo add eunomia-operator https://kohlstechnology.github.io/eunomia/
helm install my-release eunomia-operator/eunomia-operator
```

-->

## Introduction

This chart bootstraps a eunomia operator deployment on [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Kubernetes 1.16+

## Installing the Chart

This command deploys _eunomia_ on a Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

```shell
# Deploy the operator
helm template deploy/helm/eunomia-operator/ | kubectl apply -f -
```

### Installing with Kubernetes Ingress

Update [values.yaml](values.yaml) file for ingress configuration. If you don't want to change the file, you can enable ingress from command line as well.

For running with default configuration

```shell
# Enabling eunomia ingress
helm template deploy/helm/eunomia-operator/ --set eunomia.operator.ingress.enabled=true | kubectl apply -f - -n eunomia-operator
```

Also, you can pass the ingress configuration in command line itself. For example:-

```shell
# Updating eunomia ingress configuration
helm template deploy/helm/eunomia-operator/ --set eunomia.operator.ingress.enabled=true \
--set eunomia.operator.ingress.hosts[0].host=hello-eunomia.info \
--set eunomia.operator.ingress.hosts[0].paths[0].path=/ \
--set eunomia.operator.ingress.hosts[0].paths[0].portName=webhook \
--set eunomia.operator.ingress.hosts[0].paths[1].path=/metrics \
--set eunomia.operator.ingress.hosts[0].paths[1].portName=metrics | kubectl apply -f - -n eunomia-operator
```

Replace the host `hello-eunomia.info` with suitable DNS name.

### Installing with Cloud Load Balancers

Update [values.yaml](values.yaml) file for the service configuration. If you don't want to change the file, you can enable ingress from command line as well.

For running with default configuration

```shell
# Enabling eunomia with cloud load balancer
helm template deploy/helm/eunomia-operator/ \
  --set eunomia.operator.service.type=LoadBalancer \
  --set eunomia.operator.service.annotations."cloud\.google\.com\/load-balancer-type"=Internal \
  | kubectl apply -f - -n eunomia-operator
```

### Installing with PSP

```shell
# Enabling eunomia PodSecurityPolicy
helm template deploy/helm/eunomia-operator/ --set eunomia.operator.podSecurityPolicy.enabled=true | kubectl apply -f - -n eunomia-operator
```

### Installing on OpenShift

Use the below command to install Eunomia on OpenShift. This will also give you the route for the ingress webhook.

```shell
# Deploy the operator
helm template deploy/helm/eunomia-operator/ --set eunomia.operator.openshift.route.enabled=true | oc apply -f -
```

<!---
This doesn't work yet. Maybe someday it will.

## Installing the Chart

To install the chart with the release name `my-release`:

```shell
helm install my-release eunomia-operator/eunomia-operator
```

The command deploys _eunomia_ on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```shell
helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

-->

## Configuration

The following table lists the configurable parameters of the _eunomia_ chart and their default values.

| Parameter                                    | Description                                                                                                           | Default                              |
| -------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- | ------------------------------------ |
| `eunomia.operator.affinity`                  | Set `affinity` field in operator pod spec                                                                             | `nil`                                |
| `eunomia.operator.deployment.clusterViewer`  | Create eunomia-cluster-list ClusterRole                                                                               | `true`                               |
| `eunomia.operator.deployment.enabled`        | Create operator Deployment                                                                                            | `true`                               |
| `eunomia.operator.deployment.nsRbacOnly`     | Only create RBAC objects                                                                                              | `false`                              |
| `eunomia.operator.deployment.operatorHub`    | Do not create objects managed by OperatorHub install                                                                  | `false`                              |
| `eunomia.operator.image.name`                | Operator container image name                                                                                         | `kohlstechnology/eunomia-operator`   |
| `eunomia.operator.image.pullPolicy`          | Operator container image pull policy                                                                                  | `Always`                             |
| `eunomia.operator.image.repository`          | Operator container image registry name                                                                                | `quay.io`                            |
| `eunomia.operator.image.tag`                 | Operator contianer image tag                                                                                          | `latest`                             |
| `eunomia.operator.ingress.annotations`       | Set .metadata.annotaions for Ingress                                                                                  | `nil`                                |
| `eunomia.operator.ingress.enabled`           | Create Ingress for operator webhook                                                                                   | `false`                              |
| `eunomia.operator.ingress.hosts`             | Set Ingress .spec.rules                                                                                               | _see values.yaml_                    |
| `eunomia.operator.ingress.tls`               | Set Ingress .spec.tls                                                                                                 | `nil`                                |
| `eunomia.operator.namespace`                 | Namespace for operator deployment                                                                                     | `eunomia-operator`                   |
| `eunomia.operator.nodeSelector`              | Set `nodeSelector` in operator pod spec                                                                               | `nil`                                |
| `eunomia.operator.openshift.enabled`         | If `true`, enable installation on OpenShift                                                                           | `false`                              |
| `eunomia.operator.openshift.route.enabled`   | If `true`, create OpenShift Route                                                                                     | `false`                              |
| `eunomia.operator.podSecurityPolicy.enabled` | If `true`, create PodSecurityPolicy and RBAC resources                                                                | `false`                              |
| `eunomia.operator.replicas`                  | Operator Deploy .spec.replicas                                                                                        | `1`                                  |
| `eunomia.operator.resources`                 | Set operator container requests a limits                                                                              | _see values.yaml_                    |
| `eunomia.operator.service.annotaions`        | Set .metadata.annotations for Service                                                                                 | `nil`                                |
| `eunomia.operator.service.type`              | Set .spec.type for Service                                                                                            | `nil`                                |
| `eunomia.operator.serviceAccount`            | Name of servie account to run the operator pod                                                                        | `eunomia-operator`                   |
| `eunomia.operator.tolerations`               | Set `tolerations` field on operator pod spec                                                                          | `nil`                                |

