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

# This script helps in running Eunomia development binary locally.
# Eunomia watches for GitOpsConfig in local Minikube cluster,
# the logs are being dumped to standard output.

# The first script argument is a namespace to be watched by Eunomia operator.
# Default value is "" (empty) which means Eunomia watching all namespaces.
watch_ns="${1:-}"

# if minikube is not running, start it
minikube status || minikube start

kubectl apply -f ./deploy/crds/eunomia.kohls.io_gitopsconfigs_crd.yaml
export JOB_TEMPLATE=./build/job-templates/job.yaml
export CRONJOB_TEMPLATE=./build/job-templates/cronjob.yaml
export OPERATOR_NAME=eunomia-operator
operator-sdk up local --namespace="${watch_ns}"
