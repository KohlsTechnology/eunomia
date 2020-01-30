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

## we assume in $CLONED_TEMPLATE_GIT_DIR there are a set of templates with a .j2 extension
## we assume that in $CLONED_PARAMETER_GIT_DIR there is a parameter file called parameters.yaml
## the result is a set of file stored in the $MANIFEST_DIR, with the same name as the templates, but with no extension.

for file in "${CLONED_TEMPLATE_GIT_DIR}"/*.j2; do
    shortfile=$(basename -- "$file")
    # TODO consider improving by adding this filter: lib/ansible/plugins/filters/core.py
    j2 "${file}" /tmp/eunomia_values_processed.yaml \
        --import-env env \
        >"${MANIFEST_DIR}/${shortfile%.*}"
done
