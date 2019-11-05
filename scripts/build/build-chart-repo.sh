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

echo '>> Building charts...'
find "$HELM_CHARTS_SOURCE" -mindepth 1 -maxdepth 1 -type d | while read -r chart; do
  version=${TRAVIS_TAG} envsubst < "$chart"/Chart.yaml.tpl  > "$chart"/Chart.yaml
  echo ">>> helm lint $chart"
  helm lint "$chart"
  chart_dest=$HELM_CHART_DEST/"$(basename "$chart")"
  echo ">>> helm package -d $chart_dest $chart"
  mkdir -p "$chart_dest"
  helm package -d "$chart_dest" "$chart"
done
echo '>>>' "helm repo index --url https://$(dirname "$GITHUB_PAGES_REPO").github.io/$(basename "$GITHUB_PAGES_REPO") $HELM_CHART_DEST"
helm repo index --url "https://$(dirname "$GITHUB_PAGES_REPO").github.io/$(basename "$GITHUB_PAGES_REPO")" "$HELM_CHART_DEST"