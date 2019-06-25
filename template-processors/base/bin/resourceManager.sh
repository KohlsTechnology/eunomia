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

# this is needed because we want the current namespace to be set as default if a namespace is not specified.
function setContext {
  $kubectl config set-context current --namespace="${NAMESPACE}"
  $kubectl config use-context current
}

function kube {
  $kubectl -s https://kubernetes.default.svc:443  --token $(cat /var/run/secrets/kubernetes.io/serviceaccount/token) --certificate-authority=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt -n "${NAMESPACE}" $@
}

function deleteResources {
    #first we need to delete the GitOpsConfig resources whose finalizer might not work otherwise
    for file in find $MANIFEST_DIR -iregex '.*\.yaml'; do
      cat $file | yq 'select(.kind == "GitOpsConfig")' | kube delete -f - --wait=true
    done
    set +u
    kube delete -R -f $MANIFEST_DIR
    set -u
}

function createUpdateResources {
  if [ $CREATE_MODE == "CreateOrMerge" ]; then
    kube apply -R -f $MANIFEST_DIR
  fi
  if [ $CREATE_MODE == "CreateOrUpdate" ]; then
    set +u
    kube create -R -f $MANIFEST_DIR
    set -u
    kube update -R -f $MANIFEST_DIR
  fi
  if [ $CREATE_MODE == "Patch" ]; then
    kube patch -R -f $MANIFEST_DIR
  fi

}

echo Managing Resources
setContext

if [ $ACTION == "create" ] 
then
  createUpdateResources
fi

if [ $ACTION == "delete" ] 
then
  deleteResources
fi  
