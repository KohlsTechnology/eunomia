#!/usr/bin/env bash

# Copyright 2020 Kohl's Department Stores, Inc.
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

set -euo pipefail

# Colors for better user experience
WHITE='\033[1;37m'
NC='\033[0m' # No color

# This script helps in deployment of new version of Eunomia to OperatorHub
usage() {
    cat <<EOT
Usage: operatorhub-deploy.sh [--send-pr] VERSION

Creates a fork of https://github.com/operator-framework/community-operators
and pushes changes with a new Eunomia version to it. Then a PR can be
manually created.

VERSION in format x.y.z, e.g. 0.1.2

With --send-pr flag set, a Pull Request with a new Eunomia version in
https://github.com/operator-framework/community-operators will be
created.

The script will modify the contents of the deploy/ directory - a
directory structure deploy/olm-catalog/eunomia will be created. Inside
files necessary to deploy Eunomia to OperatorHub will be created. When
the script completes, the generated files get cleaned up.
EOT
}

if [[ "${1:-}" == --send-pr ]]; then
    send_pr=true
    shift
elif [[ "${1:-}" =~ ^-.*$ ]]; then
    cat <<EOT

error: You issued operatorhub-deploy.sh $@
Did you want to set the --send-pr flag?

EOT
    usage
    exit 1
fi

if ! [[ "${1:-}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "\nerror: VERSION argument not passed, or in wrong format\n" >&2
    usage
    exit 1
fi
new_version="$1"
eunomia_root=$(
    cd "$(dirname "$0")/.."
    pwd
)

# "operator-sdk olm-catalog gen-csv" command requires a specific project layout:
# https://github.com/operator-framework/operator-sdk/blob/v0.8.1/doc/user/olm-catalog/generating-a-csv.md#configuration
echo -e "${WHITE}Creating directory structure to satisfy operator-sdk"
echo -e "\n- Generating operator.yaml and role.yaml${NC}"
# Create temporary directory to store intermediate helm-generated files
tmp_dir=$(mktemp -d)
helm template "${eunomia_root}/deploy/helm/eunomia-operator/" \
    --set eunomia.operator.deployment.operatorHub=true \
    --set eunomia.operator.image.tag="v${new_version}" \
    --output-dir "$tmp_dir"
# Create operator.yaml and role.yaml
cat "${tmp_dir}/eunomia-operator/templates/deployment.yaml" \
    >"${eunomia_root}/deploy/operator.yaml"
cat "${tmp_dir}/eunomia-operator/templates/clusterrole-operators.yaml" \
    "${tmp_dir}/eunomia-operator/templates/role-operator.yaml" \
    >"${eunomia_root}/deploy/role.yaml"
# Delete temporary directory
rm -rf "$tmp_dir"

# Get previous versions from OperatorHub
echo -e "\n${WHITE}- Downloading previous releases of Eunomia from OperatorHub${NC}"
tmp_dir=$(mktemp -d)

git -C "$tmp_dir" clone https://github.com/operator-framework/community-operators
git_context="git -C ${tmp_dir}/community-operators"
hub_context="hub -C ${tmp_dir}/community-operators"

rm -rf "${eunomia_root}/deploy/olm-catalog"
mkdir -p "${eunomia_root}/deploy/olm-catalog"
cp -R "${tmp_dir}/community-operators/upstream-community-operators/eunomia" \
    "${eunomia_root}/deploy/olm-catalog"
echo -e "${WHITE}Files successfully generated${NC}"

# Find latest version
latest=$(grep currentCSV "${eunomia_root}/deploy/olm-catalog/eunomia/eunomia.package.yaml" | awk -F"eunomia.v" '{print $2}')

# Generate ClusterServiceVersion (CSV) of a new version
echo -e "\n${WHITE}Generating a new ClusterServiceVersion yaml file${NC}"
(
    cd "$eunomia_root" &&
        operator-sdk olm-catalog gen-csv --csv-version "$new_version" --from-version "$latest"
)
echo -e "\n${WHITE}- Clean up temp files${NC}"
rm "${eunomia_root}/deploy/operator.yaml" "${eunomia_root}/deploy/role.yaml"

# By default copy gitopsconfig crd from latest version
cp "${eunomia_root}/deploy/olm-catalog/eunomia/${latest}/gitopsconfigs.eunomia.kohls.io.crd.yaml" \
    "${eunomia_root}/deploy/olm-catalog/eunomia/${new_version}/"

# Make currentCSV in eunomia.package.yaml point to the new version
# Note: using -i.bkup in order for it to work both on Mac and Linux
# https://stackoverflow.com/a/22084103
sed -i.bkup "s/eunomia\.v.*$/eunomia\.v${new_version}/g" "${eunomia_root}/deploy/olm-catalog/eunomia/eunomia.package.yaml"
rm -f "${eunomia_root}/deploy/olm-catalog/eunomia/eunomia.package.yaml.bkup"
echo -e "${WHITE}ClusterServiceVersion yaml file successfully generated${NC}"

# Stop execution of the script for the user to be able to verify, or modify the files
cat <<EOT

Check if everything is alright with the new Eunomia OperatorHub release.
If you want to change something in the description, edit the new CSV file
that can be found in deploy/olm-catalog/eunomia/${new_version} directory.

To see how it will look like in OperatorHub, paste the contents of
the newly-generated CSV file to https://operatorhub.io/preview

EOT
read -r -p "Press Enter to continue or Ctrl-C to exit the script"

# Verify the generated CSV
echo -e "\n${WHITE}Verifying the newly-generated CSV${NC}"
operator-courier --verbose verify "${eunomia_root}/deploy/olm-catalog/eunomia"
operator-courier --verbose verify --ui_validate_io "${eunomia_root}/deploy/olm-catalog/eunomia"
echo -e "${WHITE}Verification successful${NC}"

# Push changes to remote fork of https://github.com/operator-framework/community-operators
echo -e "\n${WHITE}Pushing changes to your remote fork of operator-framework/community-operators${NC}"
branch_name="eunomia-v${new_version}"

# Go to cloned operator-framework/community-operators repo and fork it
$hub_context fork --remote-name tmp
$git_context checkout -b "$branch_name"
cp -R -f "${eunomia_root}/deploy/olm-catalog/eunomia" "${tmp_dir}/community-operators/upstream-community-operators/"
$git_context add .
$git_context commit -m "Eunomia release $new_version" --signoff
$git_context push --force tmp
echo -e "\n${WHITE}Successfully pushed changes to remote fork${NC}"

echo -e "\n${WHITE}Clean-up${NC}"
rm -rf "${tmp_dir}/community-operators"
rm -rf "${eunomia_root}/deploy/olm-catalog"

if [[ -n "${send_pr:-}" ]]; then
    echo -e "\n${WHITE}Creating a Pull Request in OperatorHub${NC}"
    $hub_context pull-request -m "Eunomia release $new_version"
    echo -e "${WHITE}Successfully created a PR in OperatorHub${NC}\n"
else
    cat <<EOT

If you want to deploy Eunomia v${new_version} to OperatorHub, go to the
branch ${branch_name} of your fork of
https://github.com/operator-framework/community-operators and create
a Pull Request.

EOT
fi
