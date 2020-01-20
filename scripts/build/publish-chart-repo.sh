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

# checkout

mkdir -p "$GITHUB_PAGES_DIR"
git -C "$GITHUB_PAGES_DIR" clone -b "$GITHUB_PAGES_BRANCH" "git@github.com:$GITHUB_PAGES_REPO.git" .

# build

helm init --client-only
version=${TRAVIS_TAG} envsubst < "$HELM_CHARTS_SOURCE"/Chart.yaml.tpl  > "$HELM_CHARTS_SOURCE"/Chart.yaml
helm lint "$HELM_CHARTS_SOURCE"
chart_dest=$HELM_CHART_DEST/"$(basename "$HELM_CHARTS_SOURCE")"
mkdir -p "$chart_dest"
helm package -d "$chart_dest" "$HELM_CHARTS_SOURCE"
helm repo index --url "https://$(dirname "$GITHUB_PAGES_REPO").github.io/$(basename "$GITHUB_PAGES_REPO")" "$HELM_CHART_DEST"

# publish

git -C "$GITHUB_PAGES_DIR" config user.email "travis@users.noreply.github.com"
git -C "$GITHUB_PAGES_DIR" config user.name travis
git -C "$GITHUB_PAGES_DIR" config core.sshCommand 'ssh -i github_deploy_key'
git -C "$GITHUB_PAGES_DIR" add .
git -C "$GITHUB_PAGES_DIR" status
git -C "$GITHUB_PAGES_DIR" commit -m "Published by Travis"
git -C "$GITHUB_PAGES_DIR" push origin "$GITHUB_PAGES_BRANCH"
