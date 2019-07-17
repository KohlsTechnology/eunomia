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

set -o nounset
set -o errexit

function setContext {
  $kubectl config set-context current --namespace=$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)
  $kubectl config use-context current
}

function kube {
  $kubectl -s https://kubernetes.default.svc:443  --token $(cat /var/run/secrets/kubernetes.io/serviceaccount/token) --certificate-authority=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt $@
}

function getClusterCAs {
    SECRET=$(kube describe sa default -n default | grep 'Tokens:' | awk '{print $2}')
    echo export CA_BUNDLE=$(kube get secret $SECRET -n default -o "jsonpath={.data['ca\.crt']}") >> $HOME/envs.sh
    echo export SERVICE_CA_BUNDLE=$(kube get secret $SECRET -n default -o "jsonpath={.data['service-ca\.crt']}") >> $HOME/envs.sh
}

function getDefaultRouteDomain {
    REGISTRY_ROUTE=$(kube get route docker-registry --no-headers -n default | awk '{print $2}')
    echo export DEFAULT_ROUTE_DOMAIN=${REGISTRY_ROUTE#*.} >> $HOME/envs.sh
}

function getNamespace {
    echo export NAMESPACE=$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace) >> $HOME/envs.sh
}

echo Setting cluster-related environment variable
setContext
getClusterCAs
getDefaultRouteDomain
getNamespace
