#!/usr/bin/env bash

# shellcheck disable=SC2001

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

echo "Processing Parameters"

FOLDERS=""

if [ -e "${CLONED_PARAMETER_GIT_DIR}/hierarchy.lst" ]; then
    echo "Generating hierarchy"
    # a hierarchy list was provided, so lets process it
    while IFS= read -r DIR; do
        # remove everything after # to allow comments and nuke all whitespaces
        DIR="$(echo "${DIR}" | sed -e 's/[[:space:]]*#.*//')"

        if [ -n "${DIR}" ]; then
            # add the folder to the end of the list
            DIR="$(
                cd "${CLONED_PARAMETER_GIT_DIR}/${DIR}"
                pwd
            )"
            FOLDERS="${FOLDERS} ${DIR}"
        fi
    done <<<"$(envsubst <"${CLONED_PARAMETER_GIT_DIR}/hierarchy.lst")"
else
    # No hierarchy provided, so lets just use the current folder
    FOLDERS="${CLONED_PARAMETER_GIT_DIR}"
fi

# process the folders
if [ -n "${FOLDERS}" ]; then
    # ensure our base file exists with one document to make yq happy
    VALUES_FILE="/tmp/eunomia_values_processed1.yaml"
    echo "---" >"${VALUES_FILE}"

    for DIR in ${FOLDERS}; do
        echo "Processing files in ${DIR}"

        # get the list of yaml files to process
        YAML_FILES="$(find "${DIR}" -maxdepth 1 -name \*.json -o -name \*.yaml -o -name \*.yml)"

        # merge the files
        if [ "${YAML_FILES}" ]; then
            # shellcheck disable=SC2086
            goyq merge -i -x "${VALUES_FILE}" ${YAML_FILES}
        fi
    done
else
    echo "ERROR - no folders found for processing"
    exit 1
fi

# Replace variables from enviroment
# This allows determining things like cluster names, regions, etc.
if [ -e "${VALUES_FILE}" ]; then
    envsubst <"${VALUES_FILE}" >/tmp/eunomia_values_processed.yaml
else
    echo "ERROR - missing parameter files"
    exit 1
fi
