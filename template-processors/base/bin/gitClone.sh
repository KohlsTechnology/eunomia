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

function pullFromTemplatesRepo() {
    set +u
    if [ "$TEMPLATE_GIT_HTTP_PROXY" ]; then
        local http_proxy="$TEMPLATE_GIT_HTTP_PROXY"
    fi
    if [ "$TEMPLATE_GIT_HTTPS_PROXY" ]; then
        local https_proxy="$TEMPLATE_GIT_HTTPS_PROXY"
    fi
    if [ "$TEMPLATE_GIT_NO_PROXY" ]; then
        local no_proxy="$TEMPLATE_GIT_NO_PROXY"
    fi
    if [ "$TEMPLATE_GITCONFIG" ] && [ -d "$TEMPLATE_GITCONFIG" ]; then
        cp -r "${TEMPLATE_GITCONFIG}/." "${HOME}/"
    else
        export GIT_SSL_NO_VERIFY=true
    fi
    set -u
    mkdir -p "$TEMPLATE_GIT_DIR"
    (
        export http_proxy
        export https_proxy
        export no_proxy
        git clone --depth 1 --shallow-submodules -b "$TEMPLATE_GIT_REF" "$TEMPLATE_GIT_URI" "$TEMPLATE_GIT_DIR"
        pushd "$TEMPLATE_GIT_DIR"
        git submodule init
        git submodule update --recursive --remote
        popd
    )
}

function pullFromParametersRepo() {
    set +u
    if [ "$PARAMETER_GIT_HTTP_PROXY" ]; then
        local http_proxy="$PARAMETER_GIT_HTTP_PROXY"
    fi
    if [ "$PARAMETER_GIT_HTTPS_PROXY" ]; then
        local https_proxy="$PARAMETER_GIT_HTTPS_PROXY"
    fi
    if [ "$PARAMETER_GIT_NO_PROXY" ]; then
        local no_proxy="$PARAMETER_GIT_NO_PROXY"
    fi
    if [ "$PARAMETER_GITCONFIG" ] && [ -d "$PARAMETER_GITCONFIG" ]; then
        cp -r "${PARAMETER_GITCONFIG}/." "${HOME}/"
    else
        export GIT_SSL_NO_VERIFY=true
    fi
    set -u
    mkdir -p "$PARAMETER_GIT_DIR"
    (
        export http_proxy
        export https_proxy
        export no_proxy
        git clone --depth 1 --shallow-submodules -b "$PARAMETER_GIT_REF" "$PARAMETER_GIT_URI" "$PARAMETER_GIT_DIR"
        pushd "$PARAMETER_GIT_DIR"
        git submodule init
        git submodule update --recursive --remote
        popd
    )
}

echo Cloning Repositories
pullFromTemplatesRepo
# In git, if directory contains no files, it isn't tracked:
# https://git.wiki.kernel.org/index.php/Git_FAQ#Can_I_add_empty_directories.3F
if ! [[ -d "$CLONED_TEMPLATE_GIT_DIR" ]]; then
    echo "ERROR - directory ${CLONED_TEMPLATE_GIT_DIR#/git/templates/} does not exist in the remote repository.
If you want an empty directory to be tracked by git, add a .gitkeep file inside" >&2
    exit 1
fi
pullFromParametersRepo
mkdir -p "$MANIFEST_DIR"
