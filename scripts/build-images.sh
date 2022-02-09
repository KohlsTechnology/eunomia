#!/usr/bin/env bash

# shellcheck disable=SC2002,SC2155

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

REPOSITORY=${1}
if [ -z "${CONTAINER_IMAGE_TAG:-}" ]; then
    IMAGE_TAG="dev"
else
    IMAGE_TAG=${CONTAINER_IMAGE_TAG}
fi
# Whether or not to push images. If set to anything, value will be true.
PUSH_IMAGES=${2:+true}

# Builds (and optionally pushes) a single image.
# Usage: build_image <context dir> <image name>
# Example: build_image template-processors/myimage/ myimage
build_image() {
    local context_dir=$1
    local image_url=${REPOSITORY}/$2
    local push=${PUSH_IMAGES:-false}

    if [ -f "${context_dir}/Dockerfile.in" ]; then
        cat "${context_dir}/Dockerfile.in" |
            sed "s|@REPOSITORY@|${REPOSITORY}|g" |
            sed "s|@IMAGE_TAG@|${IMAGE_TAG}|g" \
                >"${context_dir}/Dockerfile"
    fi

    # Make sure "dev" docker tag always points to the most recently changed commit
    docker build "${context_dir}" -t "${image_url}:dev"
    if $push; then
        docker push "${image_url}:dev"
    fi

    # If adjusting a git tag, set the appropriate tag in docker as well
    if [ "${IMAGE_TAG}" != "dev" ]; then
        docker tag "${image_url}:dev" "${image_url}:${IMAGE_TAG}"
        if $push; then
            docker push "${image_url}:${IMAGE_TAG}"
        fi
    fi

    # Do we have to move the "latest" tag in the docker repository?
    # Do it only if the newly tagged version is the newest one in git, according to basic SemVer semantics (vMAJOR.MINOR.PATCH).
    local latest=v$(git tag --list "v[0-9]*" | sed 's/^v//' | sort -t . -n -k 1,1 -k 2,2 -k 3,3 -r | head -n1)
    if [ "${IMAGE_TAG}" = "${latest}" ]; then
        docker tag "${image_url}:${IMAGE_TAG}" "${image_url}:latest"
        if $push; then
            docker push "${image_url}:latest"
        fi
    fi
}

# building and pushing the operator images
build_image build/ eunomia-operator

# building and pushing base template processor images
build_image template-processors/base/ eunomia-base

# building and pushing helm template processor images
build_image template-processors/helm/ eunomia-helm

# building and pushing OCP template processor images
build_image template-processors/ocp-template/ eunomia-ocp-templates

# building and pushing Applier template processor image
# NOTE: this is based on the OCP template image, so this build must always come after that.
build_image template-processors/applier/ eunomia-applier

# building and pushing jinja template processor images
build_image template-processors/jinja/ eunomia-jinja

# building and pushing openshift-provision template processor image
build_image template-processors/openshift-provision/ eunomia-openshift-provision
