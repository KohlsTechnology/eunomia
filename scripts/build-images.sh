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

REGISTRY=${REGISTRY:-quay.io/kohlstechnology}

# building and pushing the operator images
#GOOS=linux operator-sdk build $REGISTRY/eunomia-operator:v0.0.1
#docker push $REGISTRY/eunomia-operator:v0.0.1


# building and pushing base template processor images
docker build template-processors/base -t $REGISTRY/eunomia-base:v0.0.1
docker push $REGISTRY/eunomia-base:v0.0.1

# building and pushing helm template processor images
docker build template-processors/helm -t $REGISTRY/eunomia-helm:v0.0.1
docker push $REGISTRY/eunomia-helm:v0.0.1

# building and pushing OCP template processor images
docker build template-processors/ocp-template -t $REGISTRY/eunomia-ocp-templates:v0.0.1
docker push $REGISTRY/eunomia-ocp-templates:v0.0.1

# building and pushing jinja template processor images
docker build template-processors/ocp-template -t $REGISTRY/eunomia-jinja:v0.0.1
docker push $REGISTRY/eunomia-jinja:v0.0.1
