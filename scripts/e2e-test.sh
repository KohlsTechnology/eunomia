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

export EUNOMIA_PATH=$(cd "${0%/*}/.." ; pwd)

export JOB_TEMPLATE=${EUNOMIA_PATH}/build/job-templates/job.yaml
export CRONJOB_TEMPLATE=${EUNOMIA_PATH}/build/job-templates/cronjob.yaml
export WATCH_NAMESPACE=""
export OPERATOR_NAME=eunomia-operator
export TEST_NAMESPACE=test-eunomia-operator
export GO111MODULE=on

# If we're called as part of CI build on a PR, make sure we test the resources
# (templates etc.) from the PR, instead of the master branch of the main repo
if [ "${TRAVIS_PULL_REQUEST_BRANCH:-}" ]; then
  export EUNOMIA_URI="https://github.com/${TRAVIS_PULL_REQUEST_SLUG}"
  export EUNOMIA_REF="${TRAVIS_PULL_REQUEST_BRANCH}"
fi
echo "EUNOMIA_URI=${EUNOMIA_URI:-}"
echo "EUNOMIA_REF=${EUNOMIA_REF:-}"

# Ensure minikube is running
#minikube start

# Ensure clean workspace
if [[ $(kubectl get namespace $TEST_NAMESPACE) ]]; then
    kubectl delete namespace $TEST_NAMESPACE
fi

# Pre-populate the Docker registry in minikube with images built from the current commit
# See also: https://stackoverflow.com/q/42564058
eval $(minikube docker-env)
make e2e-test-images

helm template deploy/helm/eunomia-operator/ \
  --set eunomia.operator.deployment.enabled= \
  --set eunomia.operator.namespace=$TEST_NAMESPACE | kubectl apply -f -

operator-sdk test local ./test/e2e --namespace "$TEST_NAMESPACE" --up-local --no-setup

helm template deploy/helm/eunomia-operator/ \
  --set eunomia.operator.deployment.enabled= \
  --set eunomia.operator.namespace=$TEST_NAMESPACE | kubectl delete -f -

helm template deploy/helm/eunomia-operator/ \
  --set eunomia.operator.image.tag=dev \
  --set eunomia.operator.namespace=$TEST_NAMESPACE | kubectl apply -f -

kubectl wait --for=condition=available --timeout=30s deployment/eunomia-operator -n $TEST_NAMESPACE

podname=$(kubectl get pods  -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' -n $TEST_NAMESPACE)

if kubectl exec "${podname}" date -n $TEST_NAMESPACE
then
  echo "Pod is Healthy"
else
  echo "Pod is not Healthy"
  exit 1
fi

helm template deploy/helm/eunomia-operator/ \
  --set eunomia.operator.image.tag=dev \
  --set eunomia.operator.namespace=$TEST_NAMESPACE | kubectl delete -f -

