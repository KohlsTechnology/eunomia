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

# This scripts helps with deploying images locally to minikube or minishift docker registry.
usage() {
    cat <<EOT
deploy-to-local.sh minikube|minishift|kind

Build template-processors and eunomia-operator images and deploy them to
a local minikube/minishift docker registry or load it to a kind node.
EOT
}

case "${1:-}" in
minikube) eval "$(minikube docker-env)" ;;
minishift) eval "$(minishift docker-env)" ;;
kind) ;;
*)
    usage
    exit 1
    ;;
esac

GOOS=linux make
"$(dirname "$0")/build-images.sh" quay.io/kohlstechnology

if [[ "${1:-}" == "kind" ]]; then
    echo "loading latest images into kind"
    IMAGES="$(docker images --filter reference='quay.io/kohlstechnology/eunomia*:dev' --format "{{.Repository}}:{{.Tag}}")"
    if [ -z "${IMAGES}" ]; then
        echo "Something went wrong, could get the list of eunomia images from docker"
        exit 1
    fi
    for IMAGE in ${IMAGES}; do
        kind load docker-image "${IMAGE}"
    done
fi
