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
set -e

set -e

REPOSITORY=${1}
if [ -z "${TRAVIS_TAG}" ] ; then
    IMAGE_TAG="latest"
else
    IMAGE_TAG=${TRAVIS_TAG}
fi
# Whether or not to push images. If set to anything, value will be true.
PUSH_IMAGES=${2:+true}

# Builds (and optionally pushes) a single image.
# Usage: build_image <context dir> <image url> <push image (0=true, 1=false)>
# Example: build_image template-processors/myimage quay.io/KohlsTechnology/myimage:latest 0
build_image() {
  context_dir=$1
  image_url=$2
  push=${3:-false}
  docker build ${context_dir} -t ${image_url}
  if $push; then docker push $image_url; fi
}

# building and pushing the operator images
build_image build ${REPOSITORY}/eunomia-operator:${IMAGE_TAG} ${PUSH_IMAGES}

# building and pushing base template processor images
build_image template-processors/base ${REPOSITORY}/eunomia-base:${IMAGE_TAG} ${PUSH_IMAGES}

# building and pushing helm template processor images
build_image template-processors/helm ${REPOSITORY}/eunomia-helm:${IMAGE_TAG} ${PUSH_IMAGES}

# building and pushing OCP template processor images
build_image template-processors/ocp-template ${REPOSITORY}/eunomia-ocp-templates:${IMAGE_TAG} ${PUSH_IMAGES}

# building and pushing Applier template processor image
# NOTE: this is based on the OCP template image, so this build must always come after that.
build_image template-processors/applier ${REPOSITORY}/eunomia-applier:${IMAGE_TAG} ${PUSH_IMAGES}

# building and pushing jinja template processor images
build_image template-processors/jinja ${REPOSITORY}/eunomia-jinja:${IMAGE_TAG} ${PUSH_IMAGES}
