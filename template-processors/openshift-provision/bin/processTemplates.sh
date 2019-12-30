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

echo "Applying resources in ${CLONED_TEMPLATE_GIT_DIR}"
# TODO: make cluster_name configurable after GitHub issue #134 is complete
ANSIBLE_ROLES_PATH=/files/roles ansible-playbook /files/processTemplates.yml \
  ${TEMPLATE_PROCESSOR_ARGS:-} \
  -e template_directory="${CLONED_TEMPLATE_GIT_DIR}" \
  -e parameter_file="${CLONED_PARAMETER_GIT_DIR}/eunomia_values_processed.yaml"
