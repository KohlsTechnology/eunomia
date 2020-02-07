#!/usr/bin/env bash

# Copyright 2020 Kohl's Department Stores, Inc.
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

export TEST_NAMESPACE=test-eunomia-operator
export GO111MODULE=on

# Check if minishift is running
minishift status || {
    echo "Minishift is not running, aborting tests"
    exit 1
}

eval "$(minishift oc-env)"
# FIXME: sort out permissions; for now using admin, to have something working as a first step
oc login -u system:admin
if ! oc get projects | grep -q test-eunomia-operator; then
    oc new-project test-eunomia-operator
fi
oc project test-eunomia-operator

# Ensure clean workspace
if [[ $(oc get namespace $TEST_NAMESPACE) ]]; then
    oc delete namespace $TEST_NAMESPACE
fi

eval "$(minishift docker-env)"
GOOS=linux make e2e-test-images

# Eunomia setup
helm template deploy/helm/eunomia-operator/ \
    --set eunomia.operator.openshift.enabled=true \
    --set eunomia.operator.image.tag=dev \
    --set eunomia.operator.image.pullPolicy=Never \
    --set eunomia.operator.namespace=$TEST_NAMESPACE | oc apply -f -

# Deployment test
oc wait --for=condition=available --timeout=30s deployment/eunomia-operator -n $TEST_NAMESPACE
podname=$(oc get pods -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' -n $TEST_NAMESPACE)
if oc exec "${podname}" date -n $TEST_NAMESPACE; then
    echo "Eunomia deployment successful"
else
    echo "Eunomia deployment failed"
    exit 1
fi

## Testing templates-processors/ocp-template

#Test ocp-hello-cr1
ocp_hello_cr1() {
    timeout=30
    # FIXME: below CR should be modified by the script to use PR branch ref instead of master
    oc apply -f test/e2e/testdata/simple/ocp-hello-cr1.yaml
    while ((--timeout)) && [[ "$(oc get po -n test-eunomia-operator -l app=helloworld -o=jsonpath="{range .items[*]}{.status.phase}{'\n'}{end}")" != "Running" ]]; do
        echo "waiting for ocp-hello-cr1 deployment: remaining $timeout sec..."
        sleep 1
    done
    if [[ $timeout == 0 ]]; then
        echo "Example ocp-hello-cr1 Test FAILED"
        exit 1
    fi
    echo "Example ocp-hello-cr1 Test Passed"
}

ocp_hello_cr1

# FIXME: the helloworld pod is still alive after below line; ownedKinds is empty, probably because missing RBAC policies
oc delete -f test/e2e/testdata/simple/ocp-hello-cr1.yaml

oc delete namespace test-eunomia-operator

