# Utilizing the openshift-provision Ansible role

The openshift-provision template processor uses the openshift-provision Ansible
role to provision resources in your cluster. Before using this template
processor, it is recommended you have some understanding of the role which can
be found in GitHub at [gnuthought/ansible-role-openshift-provision](https://github.com/gnuthought/ansible-role-openshift-provision).

## Repository Structures

_NOTE: Your vars and resources can be kept in the same repository as hierarchy
will likely be the same for both._

### Variables repository

#### Defining variables

Variables can be defined for your cluster as yaml or json files. The files must be
included at `$VARS_REPO/<hierarchy_key>/<hierarchy_value>/vars/`. The file names must
start with a alphanumeric character, and end with `.yaml`, `.yml`, or `.json`.

#### Required fields

* `cluster/<cluster_name>/vars/main.yml`
  * At a minimum this must define the [`hierarchy`](#defining-hierarchy) for your cluster.
* `cluster/<cluster_name>/base.yml`
  * This file defines your OpenShift cluster and is passed to the
    openshift-provision role as the variable
    [`openshift_clusters`](https://github.com/gnuthought/ansible-role-openshift-provision#openshift_provision-or-openshift_clusters).

#### Example vars structure

```shell
$VARS_REPO
+-- cloud
|   +-- gcpusc1
|   |   +-- vars
|   |       +-- main.yml
|   +-- awsuse1
|       +-- vars
|           +-- main.yml
+-- cluster
|   +-- my_cluster_name
|       +-- base.yml
|       +-- vars
|           +-- main.yml
+-- default
|   +-- vars
|       +-- main.yml
+-- openshift_release
    +-- 3.11
        +-- vars
            +-- main.yml
```
_See [Defining Hierarchy](#Defining-Hierarchy) for further explanation on configuring
your repository hierarchy_

### Resource repository

#### Defining resources

In each level of your hierarchy, you can provide resource definitions. These can be
defined as yaml definitions, or as Jinja templates. All you need to do is use the
`.j2` file extension of your resource file, and openshift-provision will process
it as Jinja template before taking action on the resource.

#### Pretasks and posttasks

In the top level of the repository with your cluster resource definitions, if
you include a `cluster-pretasks.yml` and/or `cluster-posttasks.yml` file that
can be used to perform Ansible tasks to run before or after the running of the
openshift-provision role. These should be in the format of a list of tasks,
not a full playbook, because that is how the `openshift-provision` role expects them.

#### Example resources structure

```shell
$RESOURCE_REPO
+-- cloud
|   +-- gcpusc1
|   |   +-- resources
|   |       +-- cloud_resource.yml
|   +-- awsuse1
|       +-- resources
|           +-- cloud_resource.yml
+-- cluster
|   +-- my_cluster_name
|       +-- resources
|           +-- resource_a.yml
|           +-- resource_b.yml.j2
+-- cluster-pretasks.yml
+-- cluster-posttasks.yml
+-- default
    +-- resources
        +-- resource_1.yml
        +-- resource_2.yml.j2
```
_See [Defining Hierarchy](#Defining-Hierarchy) for further explanation on configuring
your repository hierarchy_

## Defining hierarchy

The openshift-provision template processor allows you to define a hierarchy for
your resources and variables. The definition of this hierarchy must be defined
in `$VARS_REPO/cluster/<cluster_name>/vars/main.yml` as a variable named `hierarchy`
that is a list of key value pairs.

The key value pairs in your hierarchy correspond to directories in your vars and resource
repositories at the paths `$VARS_REPO/<key>/<value>/vars/` and `$RESOURCE_REPO/<key>/<value>/resources/`.

There MUST be a `vars` directory present for every level of your hierarchy. For example,
if your hierarchy contains a level where `env` is set to `lle`, there MUST be a
`$VARS_REPO/env/lle/vars/` directory present in the git repository.

For resources, there does not need to be a resources directory at every level of your hierarchy.

``` yaml
$ cat $VARS_REPO/cluster/my_cluster_name/vars/main.yml
---
hierarchy:
- openshift_release: 3.11
- cloud: gcpusc1
- env: lle
- cluster: my_cluster_name
```

In this example, the order of priority for variables and resource definitions
for `my_cluster_name` would be

1. `cluster` (Highest priority)
1. `env`
1. `cloud`
1. `openshift_release`
1. `default` (Lowest priority)

## Warnings/limitations

* **Currently if you include a secret resource, it will be outputted to the pod logs.**
* The openshift-provision role is only supported on OpenShift clusters.
* The service account that the template processor is running as either needs
  access to create namespaces or the namespace needs to be already exist. If the
  namespace does not exist the openshift-provision role will attempt to create
  the namespace.
* The template processor requires setting `resourceHandlingMode: None` because resources
  are provisioned during the `processTemplates.sh` stage. Issue #139 will move this
  into the template processor definition.
