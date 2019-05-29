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

## we assume in $CLONED_TEMPLATE_GIT_DIR there is a helm chart
## the helm chart may need updating

envsubst < $CLONED_PARAMETER_GIT_DIR/values.yaml > $CLONED_PARAMETER_GIT_DIR/values_subst.yaml
helm init --client-only
helm repo update $CLONED_TEMPLATE_GIT_DIR
helm template -f $CLONED_PARAMETER_GIT_DIR/values_subst.yaml --output-dir $MANIFEST_DIR --namespace $NAMESPACE $CLONED_TEMPLATE_GIT_DIR
