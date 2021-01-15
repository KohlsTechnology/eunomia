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

export HOME=/tmp

case "$ACTION" in
create)
    /usr/local/bin/gitClone.sh
    /usr/local/bin/discoverEnvironment.sh
    # shellcheck disable=SC1090
    source $HOME/envs.sh
    hierarchy -b "${CLONED_PARAMETER_GIT_DIR}" -f "hierarchy.lst" -o "/tmp/eunomia_values_processed.yaml"
    /usr/local/bin/processTemplates.sh
    /usr/local/bin/resourceManager.sh
    ;;
delete)
    /usr/local/bin/discoverEnvironment.sh
    /usr/local/bin/resourceManager.sh
    ;;
esac
