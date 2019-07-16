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

set -e

export EUNOMIA_PATH=$GOPATH/src/github.com/KohlsTechnology/eunomia
export JOB_TEMPLATE=$EUNOMIA_PATH/templates/job.yaml
export CRONJOB_TEMPLATE=$EUNOMIA_PATH/templates/cronjob.yaml
export WATCH_NAMESPACE=""
export OPERATOR_NAME=eunomia-operator
export TEST_NAMESPACE=test-eunomia-operator

# Ensure minikube is running
#minikube start

# Ensure clean workspace
if [[ $(kubectl get namespace $TEST_NAMESPACE) ]]; then
    kubectl delete namespace $TEST_NAMESPACE
fi

kubectl create namespace $TEST_NAMESPACE
kubectl apply -f $EUNOMIA_PATH/deploy/crds/eunomia_v1alpha1_gitopsconfig_crd.yaml -n $TEST_NAMESPACE
kubectl create configmap eunomia-templates --from-file=$EUNOMIA_PATH/templates/cronjob.yaml --from-file=$EUNOMIA_PATH/templates/job.yaml -n $TEST_NAMESPACE
kubectl apply -f $EUNOMIA_PATH/deploy/kubernetes/service_account.yaml -n $TEST_NAMESPACE
kubectl apply -f $EUNOMIA_PATH/deploy/kubernetes/service.yaml -n $TEST_NAMESPACE
kubectl apply -f $EUNOMIA_PATH/deploy/kubernetes/role.yaml -n $TEST_NAMESPACE
kubectl apply -f $EUNOMIA_PATH/deploy/kubernetes/role_binding.yaml -n $TEST_NAMESPACE
operator-sdk test local ./test/e2e --namespace "$TEST_NAMESPACE" --up-local --no-setup
kubectl delete namespace $TEST_NAMESPACE
