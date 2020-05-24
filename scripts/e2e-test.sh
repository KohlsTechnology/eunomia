#!/usr/bin/env bash

# Copyright 2019 Kohl's Department Stores, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

export OPERATOR_SDK_VERSION="v0.17.1"

usage() {
    cat <<EOT
e2e-test.sh [-e|--env=(minikube|minishift)] [-p|--pause]

Execute the end-to-end tests on a local minikube or minishift.

-e|--env=(minikube|minishift) sets the environment the tests will be run under
-p|--pause Pauses after each test step to help with debugging

You can also specify the settings via environment variables (command line parameters take precedence).
EUNOMIA_TEST_ENV=(minikube|minishift)
EUNOMIA_TEST_PAUSE=yes

EOT
}

function pause() {
    if [[ "${EUNOMIA_TEST_PAUSE:-}" == "yes" ]]; then
        read -r -s -n 1 -p "Press any key to continue . . ."
        echo ""
    fi
}

# Makes sure that all gitopsconfig jobs complete, before moving on to the next test
# Requires the namespace name as the only parameter
function wait_for_gitopsconfig_completion() {
    NAMESPACE="${1}"
    timeout=60
    ALL_GOOD=0
    JOBS=""
    # Count down `timeout` to 0, then fail to avoid locking situations
    while ((--timeout)) && [[ "${ALL_GOOD}" == "0" ]]; do
        JOBS=$(kubectl get jobs -n "${NAMESPACE}" -o name | sed 's/job.batch\///g')
        if [ -z "${JOBS}" ]; then
            echo "Something went wrong, received an empty list for jobs in namespace '${NAMESPACE}'"
            exit 1
        fi
        for JOB in ${JOBS}; do
            STATUS=$(kubectl get job -n "${NAMESPACE}" "${JOB}" -o=jsonpath="{.status.conditions[*].type}{'\n'}")
            ALL_GOOD=1
            if [ "${STATUS}" == "Complete" ]; then
                echo "Job ${JOB} is finished"
            else
                echo "Job ${JOB} is still running"
                ALL_GOOD=0
            fi
        done
        echo "waiting for GitOpsConfig jobs to finish: remaining $timeout sec..."
        sleep 1
    done
    if [[ $timeout == 0 ]]; then
        echo "Timeout waiting for GitOpsConfig jobs to finish"
        exit 1
    else
        echo "All GitOpsConfig jobs finished. List of jobs:"
        echo "${JOBS}"
    fi

}

# Checks how many gitopsconfig jobs where created and compares it to the expected number
# Usage: wait_for_gitopsconfig_completion <namespace> <expected-number>
function validate_job_count() {
    NAMESPACE="${1}"
    EXPECTED="${2}"
    COUNT=$(kubectl get jobs -n "${NAMESPACE}" | grep -c gitopsconfig)
    if [ "${COUNT}" -ne "${EXPECTED}" ]; then
        echo "Error, found ${COUNT} gitopsconfig jobs instead of ${EXPECTED}"
        echo "Found the following jobs in namespace ${NAMESPACE}"
        kubectl get jobs -n "${NAMESPACE}" -o=jsonpath="{range .items[*]}{.metadata.name}{': '}{.status.conditions[*].type}{'\n'}{end}"
        exit 1
    fi
}

# Returns the active replicas count
# Usage get_replica_count <namespace> <image> <labelname>
# Example: get_replica_count "eunomia-hello-world-yaml-demo" "gcr.io/google-samples/hello-app:2.0" "hello-world"
function get_replica_count() {
    NAMESPACE="${1}"
    IMAGE="${2}"
    NAME="${3}"
    kubectl get replicaset -n "${NAMESPACE}" -l name="${NAME}" -o=jsonpath="{range .items[?(@.spec.template.spec.containers[*].image=='${IMAGE}')]}{.status.readyReplicas}{'\n'}{end}"
}

EUNOMIA_PATH=$(
    cd "${0%/*}/.."
    pwd
)

OSDK_VERSION="$(operator-sdk version)"
if ! echo "${OSDK_VERSION}" | grep "${OPERATOR_SDK_VERSION}"; then
    echo "Error: You should be using Operator-SDK ${OPERATOR_SDK_VERSION}."
    echo "Found: ${OSDK_VERSION}"
    exit 1
fi

# Process the command line parameters
PARAMS=""
while (("$#")); do
    case "$1" in
    -p | --pause) # pause between tests
        export EUNOMIA_TEST_PAUSE=yes
        shift
        ;;
    -e | --env) # set the test environment
        if [ -n "$2" ] && [ "${2:0:1}" != "-" ]; then
            EUNOMIA_TEST_ENV=$2
            shift 2
        else
            echo "Error: Argument for $1 is missing" >&2
            exit 1
        fi
        ;;
    -h | --help) # help
        usage
        exit 1
        ;;
    --* | -*) # unsupported flags
        echo "Error: Unsupported flag $1" >&2
        usage
        exit 1
        ;;
    *) # preserve positional arguments
        PARAMS="$PARAMS $1"
        shift
        ;;
    esac
done
# set positional arguments in their proper place
eval set -- "$PARAMS"

# Default settings
export EUNOMIA_TEST_ENV=${EUNOMIA_TEST_ENV:-minikube}
export EUNOMIA_TEST_PAUSE=${EUNOMIA_TEST_PAUSE:-no}

case "${EUNOMIA_TEST_ENV:-}" in
minikube) ;;
minishift) ;;
*)
    echo "Error: invalid test environment '${EUNOMIA_TEST_ENV}' specified"
    usage
    exit 1
    ;;
esac

case "${EUNOMIA_TEST_PAUSE:-}" in
yes) ;;
no) ;;
*)
    echo "Error: invalid setting for pause: '${EUNOMIA_TEST_ENV}' specified"
    echo "It must be yes or no (or undefined)"
    usage
    exit 1
    ;;
esac

echo "Test environment set to : '${EUNOMIA_TEST_ENV}'"
echo "Pausing between tests: ${EUNOMIA_TEST_PAUSE}"

export JOB_TEMPLATE=${EUNOMIA_PATH}/build/job-templates/job.yaml
export CRONJOB_TEMPLATE=${EUNOMIA_PATH}/build/job-templates/cronjob.yaml
export WATCH_NAMESPACE=""
export OPERATOR_NAME=eunomia-operator
export OPERATOR_NAMESPACE=test-eunomia-operator

# If we're called as part of CI build on a PR, make sure we test the resources
# (templates etc.) from the PR, instead of the master branch of the main repo
if [ "${TRAVIS_PULL_REQUEST_BRANCH:-}" ]; then
    export EUNOMIA_URI="https://github.com/${TRAVIS_PULL_REQUEST_SLUG}"
    export EUNOMIA_REF="${TRAVIS_PULL_REQUEST_BRANCH}"
fi
echo "EUNOMIA_URI=${EUNOMIA_URI:-}"
echo "EUNOMIA_REF=${EUNOMIA_REF:-}"

# Check if minikube is running
if [[ "${EUNOMIA_TEST_ENV}" == "minikube" ]]; then
    minikube status || {
        echo "Minikube is not running, aborting tests"
        exit 1
    }
# Check if minishift is running
elif [[ "${EUNOMIA_TEST_ENV}" == "minishift" ]]; then
    minishift status || {
        echo "Minishift is not running, aborting tests"
        exit 1
    }
fi

# Ensure clean workspace
if [[ $(kubectl get namespace $OPERATOR_NAMESPACE) ]]; then
    kubectl delete namespace $OPERATOR_NAMESPACE
fi

# Pre-populate the Docker registry in minikube/minishift with images built from the current commit
# See also: https://stackoverflow.com/q/42564058
if [[ "${EUNOMIA_TEST_ENV}" == "minikube" ]]; then
    eval "$(minikube docker-env)"
elif [[ "${EUNOMIA_TEST_ENV}" == "minishift" ]]; then
    eval "$(minishift docker-env)"
fi
GOOS=linux make e2e-test-images

# Get minikube/minishift IP address
# shellcheck disable=SC2155
if [[ "${EUNOMIA_TEST_ENV}" == "minikube" ]]; then
    export MINIKUBE_IP=$(minikube ip)
elif [[ "${EUNOMIA_TEST_ENV}" == "minishift" ]]; then
    export MINIKUBE_IP=$(minishift ip)
fi

# TestReadinessAndLivelinessProbes is accessing operator via newly created service and
# it needs to know what is the port to connect to. This value should be consistent with
# livenessProbe and readinessProbe ports in deploy/helm/eunomia-operator/templates/deployment.yaml
export OPERATOR_WEBHOOK_PORT=8080

echo "Installing Eunomia Operator"
# Eunomia setup
helm template deploy/helm/eunomia-operator/ \
    --set eunomia.operator.image.tag=dev \
    --set eunomia.operator.image.pullPolicy=Never \
    --set eunomia.operator.namespace=$OPERATOR_NAMESPACE | kubectl apply -f -

# Deployment test
kubectl wait --for=condition=available --timeout=60s deployment/eunomia-operator -n $OPERATOR_NAMESPACE
podname=$(kubectl get pods -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' -n $OPERATOR_NAMESPACE)
if kubectl exec "${podname}" date -n $OPERATOR_NAMESPACE; then
    echo "Eunomia deployment successful"
else
    echo "Eunomia deployment failed"
    exit 1
fi

pause

# End-to-end tests
operator-sdk test local ./test/e2e \
    --namespaced-manifest /dev/null \
    --global-manifest /dev/null \
    --verbose \
    --go-test-flags "-tags e2e -timeout 40m"
pause

## Testing hello-world-yaml example
# Create new namespace
kubectl create namespace eunomia-hello-world-yaml-demo
# Create new service account for the runners
kubectl apply -f examples/hello-world-yaml/eunomia-runner-sa.yaml -n eunomia-hello-world-yaml-demo

#Test hello-world-yaml-cr1
hello_world_yaml_cr_1() {
    timeout=60
    kubectl apply -f examples/hello-world-yaml/cr/hello-world-cr1.yaml -n eunomia-hello-world-yaml-demo
    while ((--timeout)) && [[ "$(get_replica_count 'eunomia-hello-world-yaml-demo' 'gcr.io/google-samples/hello-app:1.0' "hello-world")" -ne "1" ]]; do
        echo "waiting for hello-world-yaml-cr1 deployment: remaining $timeout sec..."
        sleep 1
    done
    if [[ $timeout == 0 ]]; then
        echo "Example hello-world-yaml-cr1 Test FAILED"
        exit 1
    fi
    echo "Example hello-world-yaml-cr1 Test Passed"
}

#Test hello_world_yaml_cr_2
hello_world_yaml_cr_2() {
    timeout=60
    kubectl apply -f examples/hello-world-yaml/cr/hello-world-cr2.yaml -n eunomia-hello-world-yaml-demo
    while ((--timeout)) && [[ "$(get_replica_count 'eunomia-hello-world-yaml-demo' 'gcr.io/google-samples/hello-app:1.0' "hello-world")" -ne "3" ]]; do
        echo "waiting for hello-world-yaml-cr2 deployment: remaining $timeout sec..."
        sleep 1
    done
    if [[ $timeout == 0 ]]; then
        echo "Example hello-world-yaml-cr2 Test FAILED"
        exit 1
    fi
    echo "Example hello-world-yaml-cr2 Test Passed"
}

#Test hello_world_yaml_cr_3
hello_world_yaml_cr_3() {
    timeout=60
    kubectl apply -f examples/hello-world-yaml/cr/hello-world-cr3.yaml -n eunomia-hello-world-yaml-demo
    while ((--timeout)) && [[ "$(get_replica_count 'eunomia-hello-world-yaml-demo' 'gcr.io/google-samples/hello-app:2.0' "hello-world")" -ne "3" ]]; do
        echo "waiting for hello-world-yaml-cr3 deployment: remaining $timeout sec..."
        sleep 1
    done
    if [[ $timeout == 0 ]]; then
        echo "Example hello-world-yaml-cr3 Test FAILED"
        exit 1
    fi
    echo "Example hello-world-yaml-cr3 Test Passed"
}

hello_world_yaml_cr_1
echo "Waiting 15s to verify no other gitopsconfig gets started"
sleep 15
wait_for_gitopsconfig_completion eunomia-hello-world-yaml-demo
# don't enable validate_job_count until somebody fixes the bug for multiple jobs being started (issue #343)
#validate_job_count eunomia-hello-world-yaml-demo 1
pause

hello_world_yaml_cr_2
wait_for_gitopsconfig_completion eunomia-hello-world-yaml-demo
# don't enable validate_job_count until somebody fixes the bug for multiple jobs being started (issue #343)
#validate_job_count eunomia-hello-world-yaml-demo 2
pause

hello_world_yaml_cr_3
wait_for_gitopsconfig_completion eunomia-hello-world-yaml-demo
# don't enable validate_job_count until somebody fixes the bug for multiple jobs being started (issue #343)
#validate_job_count eunomia-hello-world-yaml-demo 3
pause

# Delete namespaces after Testing hello-world-yaml example
kubectl delete namespace eunomia-hello-world-yaml-demo

# Let things settle down just a bit more
sleep 15

## Testing hello-world-helm example
# Create new namespace
kubectl create namespace eunomia-hello-world-demo

# Create the service account for the runners
kubectl apply -f examples/hello-world-helm/service_account_runner.yaml -n eunomia-hello-world-demo

#Test hello_world_helm_cr1
hello_world_helm_cr1() {
    timeout=60
    kubectl apply -f examples/hello-world-helm/cr/hello-world-cr1.yaml -n eunomia-hello-world-demo
    while ((--timeout)) && [[ "$(get_replica_count 'eunomia-hello-world-demo' 'gcr.io/google-samples/hello-app:1.0' "hello-world")" -ne "1" ]]; do
        echo "waiting for hello-world-helm-cr1 deployment: remaining $timeout sec..."
        sleep 1
    done
    if [[ $timeout == 0 ]]; then
        echo "Example hello-world-helm-cr1 Test FAILED"
        exit 1
    fi
    echo "Example hello-world-helm-cr1 Test Passed"
}

#Test hello_world_helm_cr2
hello_world_helm_cr2() {
    timeout=60
    kubectl apply -f examples/hello-world-helm/cr/hello-world-cr2.yaml -n eunomia-hello-world-demo
    while ((--timeout)) && [[ "$(get_replica_count 'eunomia-hello-world-demo' 'gcr.io/google-samples/hello-app:1.0' "hello-world")" -ne "3" ]]; do
        echo "waiting for hello-world-helm-cr2 deployment: remaining $timeout sec..."
        sleep 1
    done
    if [[ $timeout == 0 ]]; then
        echo "Example hello-world-helm-cr2 Test FAILED"
        exit 1
    fi
    echo "Example hello-world-helm-cr2 Test Passed"
}

#Test hello_world_helm_cr3
hello_world_helm_cr3() {
    timeout=60
    kubectl apply -f examples/hello-world-helm/cr/hello-world-cr3.yaml -n eunomia-hello-world-demo
    while ((--timeout)) && [[ "$(get_replica_count 'eunomia-hello-world-demo' 'gcr.io/google-samples/hello-app:2.0' "hello-world")" -ne "3" ]]; do
        echo "waiting for hello-world-helm-cr3 deployment: remaining $timeout sec..."
        sleep 1
    done
    if [[ $timeout == 0 ]]; then
        echo "Example hello-world-helm-cr3 Test FAILED"
        exit 1
    fi
    echo "Example hello-world-helm-cr3 Test Passed"
}

hello_world_helm_cr1
echo "Waiting 15s to verify no other gitopsconfig gets started"
sleep 15
wait_for_gitopsconfig_completion eunomia-hello-world-demo
# don't enable validate_job_count until somebody fixes the bug for multiple jobs being started (issue #343)
#validate_job_count eunomia-hello-world-demo 1
pause

hello_world_helm_cr2
wait_for_gitopsconfig_completion eunomia-hello-world-demo
# don't enable validate_job_count until somebody fixes the bug for multiple jobs being started (issue #343)
#validate_job_count eunomia-hello-world-demo 1
pause

hello_world_helm_cr3
wait_for_gitopsconfig_completion eunomia-hello-world-demo
# don't enable validate_job_count until somebody fixes the bug for multiple jobs being started (issue #343)
#validate_job_count eunomia-hello-world-demo 1
pause

# Delete namespaces after Testing hello-world-helm example
kubectl delete namespace eunomia-hello-world-demo

# Let things settle down just a bit more
sleep 15

## Testing hello-world-hierarchy example
# Create new namespace
kubectl create namespace eunomia-hello-world-demo
# Create new service account for the runners
kubectl apply -f examples/hello-world-helm/service_account_runner.yaml -n eunomia-hello-world-demo

#Test hello_world_hierarchy_cr1
hello_world_hierarchy_cr1() {
    timeout=60
    kubectl apply -f examples/hello-world-hierarchy/cr/hello-world-cr.yaml -n eunomia-hello-world-demo
    while ((--timeout)) && [[ "$(get_replica_count 'eunomia-hello-world-demo-hierarchy' 'gcr.io/google-samples/hello-app:1.0' "hello-world-hierarchy")" -ne "1" ]]; do
        echo "waiting for hello-world-hierarchy-cr1 deployment: remaining $timeout sec..."
        sleep 1
    done
    if [[ $timeout == 0 ]]; then
        echo "Example hello-world-hierarchy-cr1 Test FAILED"
        exit 1
    fi
    echo "Example hello-world-hierarchy-cr1 Test Passed"
}

hello_world_hierarchy_cr1
echo "Waiting 15s to verify no other gitopsconfig gets started"
sleep 15
wait_for_gitopsconfig_completion eunomia-hello-world-demo
# don't enable validate_job_count until somebody fixes the bug for multiple jobs being started (issue #343)
#validate_job_count eunomia-hello-world-demo 1
pause

# Delete namespaces after Testing hello-world-hierarchy example
kubectl delete namespace eunomia-hello-world-demo eunomia-hello-world-demo-hierarchy

#Test git_submodules
git_submodules() {
    timeout=60
    kubectl apply -f test/e2e/testdata/submodule/hello-world-submodule-cr.yaml -n eunomia-hello-world-yaml-demo
    while ((--timeout)) && [[ "$(get_replica_count 'eunomia-hello-world-yaml-demo' 'gcr.io/google-samples/hello-app:1.0' "hello-world")" -ne "1" ]]; do
        echo "waiting for hello-world-submodule-cr deployment: remaining $timeout sec..."
        sleep 1
    done
    if [[ $timeout == 0 ]]; then
        echo "git_submodules Test FAILED"
        exit 1
    fi
    echo "git_submodules Test Passed"
}

# Create new namespace
kubectl create namespace eunomia-hello-world-yaml-demo
# Create new service account for the runners
kubectl apply -f examples/hello-world-yaml/eunomia-runner-sa.yaml -n eunomia-hello-world-yaml-demo

git_submodules
pause

# Delete namespaces after Testing hello-world-yaml example
kubectl delete namespace eunomia-hello-world-yaml-demo

echo "Deleting Eunomia Operator"

# Eunomia teardown
helm template deploy/helm/eunomia-operator/ \
    --set eunomia.operator.image.tag=dev \
    --set eunomia.operator.image.pullPolicy=Never \
    --set eunomia.operator.namespace=$OPERATOR_NAMESPACE | kubectl delete -f -
