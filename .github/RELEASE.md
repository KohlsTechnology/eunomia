# Operator Hub Release steps

Set some environment variables used during the process

```shell
export new_version=<new-version>
export old_version=<old-version>
export quay_test_repo=<quay-test-repo>
export community_fork=<a-fork-of-community-operator>
```

An example could be:-

```shell
export new_version=0.1.1
export old_version=0.1.0
export quay_test_repo=kohls_technology
export community_fork=KohlsTechnology
```

## Create new CSV

Update the [`deploy/operator.yaml`](../deploy/operator.yaml) with the image tag of the version you are about to release. Also update anything else that might have change in this release in the manifests.

run the following:

```shell
operator-sdk olm-catalog gen-csv --csv-version $new_version --from-version $old_version
cp deploy/crds/eunomia_v1alpha1_gitopsconfig_crd.yaml deploy/olm-catalog/eunomia/${new_version}/
```

verify the created csv:

```shell
operator-courier --verbose verify deploy/olm-catalog/eunomia
operator-courier --verbose verify --ui_validate_io deploy/olm-catalog/eunomia
```

Reference link:-

https://github.com/operator-framework/operator-sdk/blob/master/doc/user/olm-catalog/generating-a-csv.md#configuration

## Test new CSV
Test what the operator would look like in OperatorHub, by going to this [site](https://operatorhub.io/preview) and pasting content of the newly-generated csv file

Test the operator deployment process from OperatorHub

```shell
AUTH_TOKEN=$(curl -sH "Content-Type: application/json" -XPOST https://quay.io/cnr/api/v1/users/login -d '
{
    "user": {
        "username": "'"${QUAY_USERNAME}"'",
        "password": "'"${QUAY_PASSWORD}"'"
    }
}' | jq -r '.token')
```

Push the catalog to the quay application registry (this is different than a container registry).

```shell
operator-courier push deploy/olm-catalog/eunomia $quay_test_repo eunomia $new_version "${AUTH_TOKEN}"
```

Deploy the operator source

```shell
envsubst < deploy/olm-catalog/operator-source.yaml | oc apply -f -
```

Reference Link:-

https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md

## Pushing the new CSV to OperatorHub

```shell
[[ -z "${TMPDIR}" ]] && TMPDIR=/tmp
git_context="git -C ${TMPDIR}/community-operators"
git -C ${TMPDIR} clone https://github.com/operator-framework/community-operators
$git_context remote add tmp https://github.com/${community_fork}/community-operators
$git_context checkout -b eunomia-${new_version}
cp -R -f deploy/olm-catalog/eunomia ${TMPDIR}/community-operators/upstream-community-operators/
$git_context add .
$git_context commit -m "eunomia release ${new_version}" --signoff
$git_context push tmp
hub -C ${TMPDIR}/community-operators pull-request -m "eunomia release ${new_version}"
rm -rf ${TMPDIR}/community-operators
```

