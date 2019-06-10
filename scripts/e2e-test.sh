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

export JOB_TEMPLATE=$GOPATH/src/github.com/KohlsTechnology/eunomia/templates/job.yaml
export CRONJOB_TEMPLATE=$GOPATH/src/github.com/KohlsTechnology/eunomia/templates/cronjob.yaml
export WATCH_NAMESPACE=""
export OPERATOR_NAME=eunomia-operator

kubectl create namespace test-eunomia-operator
kubectl apply -f $GOPATH/src/github.com/KohlsTechnology/eunomia/deploy/crds/eunomia_v1alpha1_gitopsconfig_crd.yaml -n test-eunomia-operator
kubectl create configmap eunomia-templates --from-file=$GOPATH/src/github.com/KohlsTechnology/eunomia/templates/cronjob.yaml --from-file=.$GOPATH/src/github.com/KohlsTechnology/eunomia/templates/job.yaml -n test-eunomia-operator
kubectl apply -f $GOPATH/src/github.com/KohlsTechnology/eunomia/deploy/kubernetes/service_account.yaml -n test-eunomia-operator
kubectl apply -f $GOPATH/src/github.com/KohlsTechnology/eunomia/deploy/kubernetes/service.yaml -n test-eunomia-operator
kubectl apply -f $GOPATH/src/github.com/KohlsTechnology/eunomia/deploy/kubernetes/role.yaml -n test-eunomia-operator
kubectl apply -f $GOPATH/src/github.com/KohlsTechnology/eunomia/deploy/kubernetes/role_binding.yaml -n test-eunomia-operator
operator-sdk test local ./test/e2e --namespace test-eunomia-operator --up-local --no-setup
kubectl delete namespace test-eunomia-operator