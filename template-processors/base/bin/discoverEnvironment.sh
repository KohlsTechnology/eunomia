#!/usr/bin/env bash

# shellcheck disable=SC2154

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

function kube {
  kubectl \
    -s https://kubernetes.default.svc:443 \
    --token "$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
    --certificate-authority=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt \
    "$@"
}

function getClusterCAs {
    echo export CA_BUNDLE="/var/run/secrets/kubernetes.io/serviceaccount/ca.crt" >> "$HOME"/envs.sh

    # service-ca.crt is only included by default in OpenShift
    if [ -e /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt ]; then
        echo export SERVICE_CA_BUNDLE="/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt" >> "$HOME"/envs.sh
    fi
}

function getNamespace {
    echo export NAMESPACE="$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)" >> "$HOME"/envs.sh
}

function setContext {
  kube config set-context current --namespace="$NAMESPACE"
  kube config use-context current
}

echo Setting cluster-related environment variable
getClusterCAs
getNamespace
setContext
