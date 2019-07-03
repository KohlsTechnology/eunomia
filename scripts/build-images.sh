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

REPOSITORY=${1}
if [ -z "${TRAVIS_TAG}" ] ; then
    IMAGE_TAG="latest"
else
    IMAGE_TAG=${TRAVIS_TAG}
fi

# building and pushing the operator images
docker build . -t ${REPOSITORY}/eunomia-operator:${IMAGE_TAG} -f build/Dockerfile
docker push ${REPOSITORY}/eunomia-operator:${IMAGE_TAG}

# building and pushing base template processor images
docker build template-processors/base -t ${REPOSITORY}/eunomia-base:${IMAGE_TAG}
docker push ${REPOSITORY}/eunomia-base:${IMAGE_TAG}

# building and pushing helm template processor images
docker build template-processors/helm -t ${REPOSITORY}/eunomia-helm:${IMAGE_TAG}
docker push ${REPOSITORY}/eunomia-helm:${IMAGE_TAG}

# building and pushing OCP template processor images
docker build template-processors/ocp-template -t ${REPOSITORY}/eunomia-ocp-templates:${IMAGE_TAG}
docker push ${REPOSITORY}/eunomia-ocp-templates:${IMAGE_TAG}

# building and pushing jinja template processor images
docker build template-processors/jinja -t ${REPOSITORY}/eunomia-jinja:${IMAGE_TAG}
docker push ${REPOSITORY}/eunomia-jinja:${IMAGE_TAG}
