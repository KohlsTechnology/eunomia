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

# Simple script to pull the images from quay and push them into a local registry
# It's intended for the unfortunate of us that have to deal with corporate proxies...
# This should probably be revised to use skopeo

VERSION="v0.0.1"
REGISTRY="${1}"

if [ -z "${REGISTRY}" ] ; then
    echo "Please specify the target registry and path as the only parameter"
    echo "Example: ${0} myawesomeregistry.eunomia.io/some/path"
    exit 1
fi

docker login $REGISTRY || exit 1

for IMAGE in "eunomia-operator" "eunomia-base" "eunomia-helm" "eunomia-ocp-templates" "eunomia-jinja"
do
    docker pull quay.io/kohlstechnology/$IMAGE:$VERSION || exit 1
    docker tag quay.io/kohlstechnology/$IMAGE:$VERSION $REGISTRY/$IMAGE:$VERSION || exit 1
    docker push $REGISTRY/$IMAGE:$VERSION || exit 1
done

