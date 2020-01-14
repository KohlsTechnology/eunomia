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

go test ./test/e2e/... -tags xd -v

export TEST_NAMESPACE=test-eunomia-operator

helm template deploy/helm/eunomia-operator/ \
  --set eunomia.operator.deployment.enabled= \
  --set eunomia.operator.namespace=$TEST_NAMESPACE | kubectl delete -f -

helm template deploy/helm/eunomia-operator/ \
  --set eunomia.operator.image.tag=dev \
  --set eunomia.operator.image.pullPolicy=Never \
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

