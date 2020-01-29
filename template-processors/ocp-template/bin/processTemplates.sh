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

## we assume in $CLONED_TEMPLATE_GIT_DIR there is a template called template.yaml
## we assume that in $CLONED_PARAMETER_GIT_DIR there is a parameter file called parameters.ini

envsubst < "$CLONED_PARAMETER_GIT_DIR/parameters.ini" > "$CLONED_PARAMETER_GIT_DIR/parameters_subst.ini"
oc process -f "$CLONED_TEMPLATE_GIT_DIR/template.yaml" --param-file="$CLONED_PARAMETER_GIT_DIR/parameters_subst.ini" > "$MANIFEST_DIR/manifests.yaml"
