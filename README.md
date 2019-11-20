![eunomia-logo-small.png](./media/eunomia-logo-small.png)

# Eunomia - a GitOps Operator for Kubernetes

[![Join the chat at https://gitter.im/KohlsTechnology/eunomia](https://badges.gitter.im/KohlsTechnology/eunomia.svg)](https://gitter.im/KohlsTechnology/eunomia?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) [![Build Status](https://travis-ci.com/KohlsTechnology/eunomia.svg?branch=master)](https://travis-ci.com/KohlsTechnology/eunomia) [![Docker Repository on Quay](https://quay.io/repository/kohlstechnology/eunomia-operator/status "Docker Repository on Quay")](https://quay.io/repository/kohlstechnology/eunomia-operator) [![Go Report Card](https://goreportcard.com/badge/github.com/KohlsTechnology/eunomia)](https://goreportcard.com/report/github.com/KohlsTechnology/eunomia)

## Who is Eunomia

According to Wikipedia:

>Eunomia (Greek: Εὐνομία) was a minor Greek goddess of law and legislation (her name can be translated as "good order", "governance according to good laws"), as well as the spring-time goddess of green pastures (eû means "well, good" in Greek, and νόμος, nómos, means "law", while pasturelands are called nomia).

## What is GitOps

GitOps is all about turning day 2 operations into code! Not just that, it means you start thinking about day 2 on day 1. This is a dream come true for any Operations team!
GitOps leverages the strength of automation and combines it with the power of git based workflows. It is a natural evolution beyond infrastructure-as-code and builds on top of DevOps best practices.

### Next Generation Change Management

Especially in large Enterprises, Change Management is usually a painful experience. GitOps allows to take a lot of that pain out and streamline the process itself. It does so by still providing what the process tries to accomplish (and thus still meet audit requirements), but does so in a way that is much faster, much more secure, and much more reliable.

Your changes now all of a sudden provide:

- Version Control
- Peer Reviews
- Approvals
- Audit Trail
- Reproducibility
- Consistency
- Reliability

What's your backout plan for your change? How about simply moving back to the previous commit and getting EXACTLY what you had before?

## Purpose

The Eunomia provides the ability to implement these git-based flows for any resources in Kubernetes. Eunomia does not care if you have a plain Kubernetes, a cloud based Kubernetes (like GKE), or a complete PaaS platform based on Kubernetes (like OpenShift). Eunomia also does not care how you want to structure your data, how many repos you want to use, or which templating engine is your favourite.

Eunomia can handle straight-up (static) yaml files with the complete definition or create dynamic ones based on your templating engine of choice. Eunomia already supports *Helm Charts*, *OpenShift Templates*, and *Jinja2 Templates*, but can easily be extended to support others.

These templates will be merged and processed with a set of environment-specific parameters to get a list of resource manifests. Then these manifest can be created/updated/deleted in Kubernetes.

## Vision

While this controller can certainly be used to directly populate an individual namespace with a configuration stored in git, the vision is that a hierarchy of controllers will be used to populate multiple namespaces. Ideally this approach will be used to bring a newly created cluster to a desired configured state. Only the initial seeding CR should have cluster-level permissions. Any sub-CRs should have progressively less access assigned to their service accounts.

Here is a very simple example of how the hierarchy might look like:

![Gitops-hierarchy.png](./media/Gitops-hierarchy.png)

The main sections of the GitOpsConfig CRD are described below.

## Example

The configuration is described in the GitOpsConfig CRD, here is an example:

```yaml
apiVersion: eunomia.kohls.io/v1alpha1
kind: GitOpsConfig
metadata:
  name: simple-test
spec:
  # Add fields here
  templateSource:
    uri: https://github.com/KohlsTechnology/eunomia
    ref: master
    contextDir: simple/templates
    secretRef: template-gitconfig
  parameterSource:
    contextDir: simple/parameters
    secretRef: parameter-gitconfig  
  triggers:
  - type: Change
  - type: Webhook
  - type: Periodic
    cron: "0 * * * *"
  ServiceAccountRef:      "mysvcaccount",
  templateProcessorImage: mydockeregistry.io:5000/gitops/eunomia-base:latest
  ResourceDeletionMode:   "Cascade",
  ResourceHandlingMode:   "CreateOrMerge",
```

## TemplateSource and ParameterSource

The `TemplateSource` and `ParameterSource` specify where the templates and the parameters are stored. The exact contents of these locations depend on the templating engine that has been selected.

The fields of this section are:

```yaml
  templateSource:
    uri: https://github.com/KohlsTechnology/eunomia
    ref: master
    contextDir: simple/templates
    HTTPProxy: <http proxy>
    HTTPSProxy: <https proxy>
    NOProxy: <no Proxy>
    SecretRef: <gitconfig and credentials secret>

  parameterSource:
    uri: https://github.com/KohlsTechnology/eunomia
    ref: master
    contextDir: seed/parameters
    HTTPProxy: <http proxy>
    HTTPSProxy: <https proxy>
    NOProxy: <no Proxy>
    SecretRef: <gitconfig and credentials secret>
```

These are the mandatory constraints and default behaviors of the fields:

| field name  | mandatory  | default  |
|:---|:---:|:---|
| uri  | yes  | N/A  |
| ref  | no  | `master`  |
| contextDir  | no  | `.`  |
| HTTPProxy  | no  |   |
| HTTPSProxy  | no  |   |
| NOProxy  | no  |   |
| SecretRef  | no  |   |

If a secret is provided, then it is assumed that the connection to Git requires authentication. See the [Git Authentication](#git\ authentication) section below for more details.

If the `uri` is not specified in the `parameterSource` section, then it will default to the `uri` specified under `templateSource`.

### parameterSource processing
Eunomia uses the [yq command](http://mikefarah.github.io/yq/) to merge all yaml files in the specified folder. You have to be careful, if you have the same variable name in multiple files. Dictionaries will merge, lists will get overwritten. 

#### Variable Hierarchy
You can provide a file `hierarchy.lst`, to allow a variable hierarchy. This will allow you to specify a default value and overwrite it on an environment level if necessary. This will greatly simplify your configuration and allows for deduplication of data, making your operational life a lot easier.

The contents of the file are simply relative path names, with the base being `contextDir`.

Example `hierarchy.lst`:
```
../defaults    #this is first ... lowest priority
../marketing   #this is second
../development #this is third ... highest priority
```

In this case it will load all yaml files from `../defaults`, then merge it with everything in `../marketing`, and lastly merges it with everything in `../development`. You can also use the relative path `./`, which means it'll also load the variables defined in `contextDir` directly (same folder that as `hierarchy.lst`).

#### Upcoming features
Once [issue #4](https://github.com/KohlsTechnology/eunomia/issues/4) is resolved, you will be able to specify variable names to dynamically determine the correct folder. This will allow you to only have one `hierarchy.lst`. (Technically, it is actually already possible to use environment variables, but without #4, there are just none set that would be of any practical use in hierarchy.lst.)

### Git Authentication

Specifying a `SecretRef` will automatically turn on git authentication. The secrets for the template and parameter repos will be mounted respectively in the `/template-gitconfig` and `/parameter-gitconfig` of the job pod.
The referenced secrets must be available and how they are provisioned is beyond the scope of this operator. See the [Vision](#vision) paragraph on how to build a hierarchical structure, where the resources needed to run a given GitOpsConfig are configured by a predecessor GitpOpsConfig instance.

This secret will be linked from `~/` of the used running the pod. The secret *must* contain a `.gitconfig` file and may contain other files. The passed `.gitconfig` will be used during the git operations. It is advised to reference any additional files via the absolute path.

#### Username and password authentication

For username and password based authentication create the following `.gitconfig`:

```ini
[http]
    sslCAInfo = /template-gitconfig/ca.crt

[user]
    name = gitconfig

[credential]
    helper = store
```

When the credential helper is of type `store`, credentials are by default retrieved from the `~/.git-credentials` file. This file should also be added to the secret and has the following format:

```text
https://<username>:<password>@<git server fqdn>
```

Don't forget to provide the `ca.crt` file to the secret.

#### Certificate based authentication

For certificate based authentication, create the following `.gitconfig`:

```ini
[core]
    sshCommand = 'ssh -i /template-gitconfig/mykey.rsa -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no'
```

and add the `mykey.rsa` file to the secret.

## Triggers

You can enable one or multiple triggers.

| Name  | Description  |
|:---|:---|
|`Change` | This triggers every time the CR is changed, including when it is created.|
|`Periodic` | Periodically apply the configuration. This can be used to either schedule changes for a specific time, use it for drift management to revert any changes, or as a safeguard in case webhooks were missed. It uses a cron-style expression.
|`Webhook` | This triggers when something on git changes. You have to configure the webhook yourself.

## Template Engine

When it's time to apply a configuration, the GitOps controller runs a job pod. The image of the job pod can be specified in the `templateProcessorImage` field.
This is the plugin mechanism to support multiple template engines.
A base image is provided that can be inherited to simplify the process of adding support for a new templating engine.
The base image provides the following workflow:

1. `gitClone.sh` : This will clone the template and parameter repos. It is expected that there will be no need to customize this. Any required changes are most likely worthy of a pull-request upstream.
2. `discoverEnvironment.sh` : This will create a set of environment variables that are specific to the target Kubernetes environment. Currently the following variable are supported:

    | Name  | Description  |
    |:---|:---|
    | `CA_BUNDLE`  | Platform-level CA bundle  |
    | `SERVICE_CA_BUNDLE`  | Service-level CA bundle  |
    | `NAMESPACE`  | Current namespace  |
3. `processParameters.sh` : This script processes all the parameter files and generates `/tmp/eunomia_values_processed.yaml`. This script currently supports the following features:
    - Merging of all existing yaml files in the `CLONED_PARAMETER_GIT_DIR` location, into a single file for processing by the templating engine. 
    - Substitution of variables with environment variables.

    This script can be further enhanced to e.g. support secrets injection.

4. `processTemplate.sh` : This file needs to be overwritten in order to support a different templating engine. The contract is the following:

    - Templates are available at the location specified by the variable: `CLONED_TEMPLATE_GIT_DIR`
    - Parameters are available at the location specified by the variable: `CLONED_PARAMETER_GIT_DIR`
    - After the template processing completes, the processed manifests should be stored at the location of this variable: `MANIFEST_DIR`

5. `resourceManager.sh` :  Processes the resources in `MANIFEST_DIR`. One or more files can be present, and all will be processed.

Currently the following templating engines are supported (follow the link to see examples of how new template processors can be added):

- [Raw Manifests](./template-processors/base)
- [OpenShift Templates](./template-processors/ocp-template)
- [Helm Charts](./template-processors/helm)
- [Jinja Templates](./template-processors/jinja)
- [OpenShift Applier](./template-processors/applier)

## serviceAccountRef

This is the service account used by the job pod that will process the resources. The service account must be present in the same namespace as the one where the GitOpsConfig CR is and must have enough permission to manage the resources. It is out of scope of this controller how that service account is provisioned, although you can use a different GitOpsConfig CR to provision it (seeding CR).

## Resource Handling Mode

This field specifies how resources should be handled, once the templates are processed. The following modes are currently supported:

1. `CreateOrMerge`, which is roughly equivalent to `kubectl apply`.
2. `CreateOrUpdate`, which will overwrite any existing configuration.
3. `Patch`. Patch requires objects to already exists and will patch them. It's useful when customizing objects that are provided through other means.
4. `None`. In some cases there may be template processors or automation frameworks where the processing of templates and handling of generated resources are a single step. In that case, Eunomia can be configured to skip the built-in resource handling step.

## Resource Deletion Mode

This field specifies how to handle resources when the GitOpsConfig object is deleted. Two options are available:

1. `Retain`, resources previously created are left intact.
2. `Delete`, resources are delete with the `cascade` option.
3. `None`, resource deletion is not handled at all.

## Installing Eunomia

### Installing on Kubernetes

Simply use the helm chart to install it on your flavor of Kubernetes.

```shell
# Deploy the operator
helm template deploy/helm/eunomia-operator/ | kubectl apply -f -
```

### Installing on OpenShift

Use the below command to install Eunomia on OpenShift. This will also give you the route for the ingress webhook.

```shell
# Deploy the operator
helm template deploy/helm/eunomia-operator/ --set openshift.route.enabled=true | oc apply -f -
```

## Examples / Demos

We've created several examples for you to test out Eunomia. See [EXAMPLES](examples/README.md) for details.

## Monitoring with Prometheus

[Prometheus](https://prometheus.io/) is an open-source systems monitoring and alerting toolkit.

Prometheus collects metrics from monitored targets by scraping metrics HTTP endpoints.

- [configuring-prometheus](https://prometheus.io/docs/introduction/first_steps/#configuring-prometheus)

- `scrape_configs` controls what resources Prometheus monitors.

- `kubernetes_sd_configs` Kubernetes SD configurations allow retrieving scrape targets. Please see [kubernetes_sd_configs](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#endpoints) for details.

- Additionally, `relabel_configs` allow advanced modifications to any target and its labels before scraping.

By default, the metrics in Operator SDK are exposed on `0.0.0.0:8383/metrics`

For more information, see [Metrics in Operator SDK](https://github.com/operator-framework/operator-sdk/blob/v0.8.1/doc/user/metrics/README.md)

### Usage:

```
scrape_configs:
  - job_name: 'kubernetes-service-endpoints'
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
      - source_labels: [__meta_kubernetes_namespace]
        action: keep
        regex: test-eunomia-operator
```
You can find additional examples on their [GitHub page](https://github.com/prometheus/prometheus/blob/master/documentation/examples/prometheus-kubernetes.yml).

### Verify metrics port:
kubectl exec `POD-NAME` curl localhost:8383/metrics  -n `NAMESPACE`

(e.g. `kubectl exec eunomia-operator-5b9b664cfc-6rdrh curl localhost:8383/metrics  -n test-eunomia-operator`)

## Development

Please see our [development documentation](DEVELOPMENT.md) for details.

## Troubleshooting

Please see our [troubleshooting guide](TROUBLESHOOTING.md) for details.

## License

See [LICENSE](LICENSE) for details.

## Code of Conduct

See [CODE_OF_CONDUCT.md](.github/CODE_OF_CONDUCT.md)
for details.
