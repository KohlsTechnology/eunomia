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

oc create namespace test-gitops-operator
oc project test-gitops-operator
oc create configmap gitops-templates --from-file=./templates/cronjob.yaml --from-file=./templates/job.yaml -n test-gitops-operator
oc apply -f ./test/deploy -n test-gitops-operator
operator-sdk test local ./test/e2e --namespace test-gitops-operator --no-setup
oc delete project test-gitops-operator

