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

echo Processing Parameters

export VALUES_FILE=$CLONED_PARAMETER_GIT_DIR/values.yaml

# do a merge if there's more than one yaml file
if [ "$(ls -1 $CLONED_PARAMETER_GIT_DIR/*.yaml | wc -l)" -gt 1 ]; then
  echo "Merging all available yaml files"
  goyq merge $CLONED_PARAMETER_GIT_DIR/*.yaml > $CLONED_PARAMETER_GIT_DIR/eunomia_values_processed.yaml
  export VALUES_FILE=$CLONED_PARAMETER_GIT_DIR/eunomia_values_tmp.yaml
else
  if [ -e "$CLONED_PARAMETER_GIT_DIR/values.yaml" ]; then
    mv $CLONED_PARAMETER_GIT_DIR/values.yaml $CLONED_PARAMETER_GIT_DIR/eunomia_values_processed.yaml
  fi
fi
