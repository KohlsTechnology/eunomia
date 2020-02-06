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

set -euxo pipefail

EUNOMIA_PATH=$(
    cd "${0%/*}/.."
    pwd
)

export JOB_TEMPLATE=${EUNOMIA_PATH}/build/job-templates/job.yaml
export CRONJOB_TEMPLATE=${EUNOMIA_PATH}/build/job-templates/cronjob.yaml
export WATCH_NAMESPACE=""
export OPERATOR_NAME=eunomia-operator
export OPERATOR_NAMESPACE=test-eunomia-operator
export GO111MODULE=on

# If we're called as part of CI build on a PR, make sure we test the resources
# (templates etc.) from the PR, instead of the master branch of the main repo
if [ "${TRAVIS_PULL_REQUEST_BRANCH:-}" ]; then
    export EUNOMIA_URI="https://github.com/${TRAVIS_PULL_REQUEST_SLUG}"
    export EUNOMIA_REF="${TRAVIS_PULL_REQUEST_BRANCH}"
fi
echo "EUNOMIA_URI=${EUNOMIA_URI:-}"
echo "EUNOMIA_REF=${EUNOMIA_REF:-}"

# Check if minikube is running
minikube status || {
    echo "Minikube is not running, aborting tests"
    exit 1
}

# Ensure clean workspace
if [[ $(kubectl get namespace $OPERATOR_NAMESPACE) ]]; then
    kubectl delete namespace $OPERATOR_NAMESPACE
fi

# Pre-populate the Docker registry in minikube with images built from the current commit
# See also: https://stackoverflow.com/q/42564058
eval "$(minikube docker-env)"
GOOS=linux make e2e-test-images

# Get minikube IP address
# shellcheck disable=SC2155
export MINIKUBE_IP=$(minikube ip)

# TestReadinessAndLivelinessProbes is accessing operator via newly created service
# and it needs to know what is the port to connect to.
export OPERATOR_WEBHOOK_PORT=8080

# Eunomia setup
helm template deploy/helm/eunomia-operator/ \
    --set eunomia.operator.image.tag=dev \
    --set eunomia.operator.image.pullPolicy=Never \
    --set eunomia.operator.namespace=$OPERATOR_NAMESPACE | kubectl apply -f -

# Deployment test
kubectl wait --for=condition=available --timeout=30s deployment/eunomia-operator -n $OPERATOR_NAMESPACE
podname=$(kubectl get pods -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' -n $OPERATOR_NAMESPACE)
if kubectl exec "${podname}" date -n $OPERATOR_NAMESPACE; then
    echo "Eunomia deployment successful"
else
    echo "Eunomia deployment failed"
    exit 1
fi

# End-to-end tests
operator-sdk test local ./test/e2e \
    --namespaced-manifest /dev/null \
    --global-manifest /dev/null \
    --verbose \
    --go-test-flags "-tags e2e -timeout 20m"

## Testing hello-world-yaml example
# Create new namespace
kubectl create namespace eunomia-hello-world-yaml-demo
# Create new service account for the runners
kubectl apply -f examples/hello-world-yaml/eunomia-runner-sa.yaml -n eunomia-hello-world-yaml-demo

#Test hello-world-yaml-cr1
hello_world_yaml_cr_1() {
    timeout=30
    kubectl apply -f examples/hello-world-yaml/cr/hello-world-cr1.yaml -n eunomia-hello-world-yaml-demo
    while ((--timeout)) && [[ "$(kubectl get po -n eunomia-hello-world-yaml-demo -l name=hello-world -o=jsonpath="{range .items[*]}{.status.phase}{'\n'}{end}")" != "Running" ]]; do
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
    timeout=30
    kubectl apply -f examples/hello-world-yaml/cr/hello-world-cr2.yaml -n eunomia-hello-world-yaml-demo
    while ((--timeout)) && [[ "$(kubectl get replicaset -n eunomia-hello-world-yaml-demo -l name=hello-world -o=jsonpath="{range .items[*]}{.status.readyReplicas}{'\n'}{end}")" != "3" ]]; do
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
    timeout=30
    kubectl apply -f examples/hello-world-yaml/cr/hello-world-cr3.yaml -n eunomia-hello-world-yaml-demo
    while ((--timeout)) && [[ "$(kubectl get deployment -n eunomia-hello-world-yaml-demo -o=jsonpath="{range .items[*]}{.status.observedGeneration}{'\n'}{end}")" != "3" ]]; do
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
hello_world_yaml_cr_2
hello_world_yaml_cr_3

# Delete namespaces after Testing hello-world-yaml example
kubectl delete namespace eunomia-hello-world-yaml-demo

## Testing hello-world-helm example
# Create new namespace
kubectl create namespace eunomia-hello-world-demo

# Create the service account for the runners
kubectl apply -f examples/hello-world-helm/service_account_runner.yaml -n eunomia-hello-world-demo

#Test hello_world_helm_cr1
hello_world_helm_cr1() {
    timeout=60
    kubectl apply -f examples/hello-world-helm/cr/hello-world-cr1.yaml -n eunomia-hello-world-demo
    while ((--timeout)) && [[ "$(kubectl get po -n eunomia-hello-world-demo -l name=hello-world -o=jsonpath="{range .items[*]}{.status.phase}{'\n'}{end}")" != "Running" ]]; do
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
    while ((--timeout)) && [[ "$(kubectl get replicaset -n eunomia-hello-world-demo -l name=hello-world -o=jsonpath="{range .items[*]}{.status.readyReplicas}{'\n'}{end}")" != "3" ]]; do
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
    while ((--timeout)) && [[ "$(kubectl get deployment -n eunomia-hello-world-demo -o=jsonpath="{range .items[*]}{.status.observedGeneration}{'\n'}{end}")" != "3" ]]; do
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
hello_world_helm_cr2
hello_world_helm_cr3

# Delete namespaces after Testing hello-world-helm example
kubectl delete namespace eunomia-hello-world-demo

## Testing hello-world-hierarchy example
# Create new namespace
kubectl create namespace eunomia-hello-world-demo
# Create new service account for the runners
kubectl apply -f examples/hello-world-helm/service_account_runner.yaml -n eunomia-hello-world-demo

#Test hello_world_hierarchy_cr1
hello_world_hierarchy_cr1() {
    timeout=90
    kubectl apply -f examples/hello-world-hierarchy/cr/hello-world-cr.yaml -n eunomia-hello-world-demo
    while ((--timeout)) && [[ "$(kubectl get po -n eunomia-hello-world-demo-hierarchy -o=jsonpath="{range .items[*]}{.status.phase}{'\n'}{end}")" != "Running" ]]; do
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

# Delete namespaces after Testing hello-world-hierarchy example
kubectl delete namespace eunomia-hello-world-demo eunomia-hello-world-demo-hierarchy

# Eunomia teardown
helm template deploy/helm/eunomia-operator/ \
    --set eunomia.operator.image.tag=dev \
    --set eunomia.operator.image.pullPolicy=Never \
    --set eunomia.operator.namespace=$OPERATOR_NAMESPACE | kubectl delete -f -
