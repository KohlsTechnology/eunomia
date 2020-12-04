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

TAG_OWNER="gitopsconfig.eunomia.kohls.io/owner"
TAG_APPLIED="gitopsconfig.eunomia.kohls.io/applied"

# this is needed because we want the current namespace to be set as default if a namespace is not specified.
function setContext() {
    # shellcheck disable=SC2154
    $kubectl config set-context current --namespace="$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)"
    $kubectl config use-context current
}

function kube() {
    $kubectl \
        -s https://kubernetes.default.svc:443 \
        --token "$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
        --certificate-authority=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt \
        "$@"
}

# addLabels OWNER TIMESTAMP - patches the YAML&JSON files in $MANIFEST_DIR,
# adding labels tracking the OWNER and TIMESTAMP. The labels are intended to be
# used later in function deleteByOldLabels.
function addLabels() {
    local owner="$1"
    local timestamp="$2"
    local tmpdir="$(mktemp -d)"
    # shellcheck disable=SC2044
    for file in $(find "$MANIFEST_DIR" -regextype posix-extended -iregex '.*\.(ya?ml|json)'); do
        cat "$file" |
            yq -y -s "map(select(.!=null)|setpath([\"metadata\",\"labels\",\"$TAG_OWNER\"]; \"$owner\"))|.[]" |
            yq -y -s "map(select(.!=null)|setpath([\"metadata\",\"labels\",\"$TAG_APPLIED\"]; \"$timestamp\"))|.[]" \
                >"$tmpdir/labeled"
        # We must use a helper file (can't do this in single step), as the file would be truncated if we read & write from it in one pipeline
        cat "$tmpdir/labeled" >"$file"
    done
}

# appendResourceVersion - patches the YAML&JSON files in $MANIFEST_DIR,
# adding the metadata.resourceVersion for each resource being managed.
# This is intended to serve as a locking mechanism when applying resources
# in which Kubernetes will fail the apply with a StatusConflict (HTTP status code 409)
# Ref https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
function appendResourceVersion() {
    local tmpdir="$(mktemp -d)"
    mkdir -p "$tmpdir/$MANIFEST_DIR"
    mkdir -p "$tmpdir/processed_files/$MANIFEST_DIR"
    # shellcheck disable=SC2044
    for file in $(find "$MANIFEST_DIR" -regextype posix-extended -iregex '.*\.(ya?ml|json)'); do
        mkdir -p "$tmpdir"/"$(dirname "$file")"
        mkdir -p "$tmpdir"/processed_files/"$(dirname "$file")"
        kube get --ignore-not-found -f "$file" -o yaml >"$tmpdir/$file"
        local kind="$(cat "$tmpdir"/"$file" | yq -y .kind)"
        if [[ "$kind" =~ "List" ]]; then
            #TODO: figure out how to handle this
            echo "THIS IS A LIST"
        else
            echo "THIS IS A $kind"
            local resourceVersion="$(cat "$tmpdir"/"$file" | yq -y .metadata.resourceVersion)"
            if [[ "${resourceVersion}" ]]; then
                cat "$file" |
                    yq -y -s "map(select(.!=null)|setpath([\"metadata\",\"resourceVersion\"]; \"$resourceVersion\"))|.[]" \
                        >"$tmpdir/labeled"

                cat "$tmpdir/labeled" >"$tmpdir/processed_files/$file"
            else
                echo "NO RESOURCE VERSION TO PATCH"
                cat "$file" >"$tmpdir/processed_files/$file"
            fi
        fi
    done
}

# deleteByOldLabels OWNER [TIMESTAMP] - deletes all kubernetes resources which have
# the OWNER label as provided [optional: but TIMESTAMP label different than provided].
function deleteByOldLabels() {
    if [ "$DELETE_MODE" == "None" ]; then
        echo "DELETE_MODE is set to None; Skipping deletion by old labels step."
        exit 0
    else
        local owner="$1"
        local timestamp="${2:-}"
        local allKinds="$(kube api-resources --verbs=list,delete -o name | paste -sd, -)"
        local ownedKinds="$(kube get "$allKinds" --ignore-not-found \
            -l "$TAG_OWNER==$owner" \
            -o jsonpath="{range .items[*]}{.kind} {.apiVersion}{'\n'}{end}" | # e.g. "Pod v1" OR "StorageClass storage.k8s.io/v1"
            sort -u |
            awk -F'[ /]' '{if (NF==2) {print $1} else {print $1"."$3"."$2}}' | # e.g. "Pod" OR "StorageClass.v1.storage.k8s.io"
            paste -sd, -)"
        if [ -z "$ownedKinds" ]; then
            return
        fi
        local filter="${TAG_OWNER}==${owner}"
        if [[ "${timestamp}" ]]; then
            filter="${filter},${TAG_APPLIED}!=${timestamp}"
            # Retrieve all resources owned by the GitOpsConfig that doesn't match the current jobs timestamp
            # Check the timestamp on all of the resources and ONLY delete the resource if the timestamp label is older than the current job timestamp
            # shellcheck disable=SC2005
            echo "$(kube get "${ownedKinds}" -l "${filter}" -o yaml)" >/tmp/check_deletion.yaml
            local resource_count=$(($(yq -y '.items | length' /tmp/check_deletion.yaml | head -qn 1) - 1))
            if [[ "$resource_count" -ge "0" ]]; then
                for i in $(seq 0 $resource_count); do
                    local resource_timestamp="$(yq -r -y '.items['"$i"'].metadata.labels."'$TAG_APPLIED'"' /tmp/check_deletion.yaml | head -qn 1)"
                    if [[ "$resource_timestamp" -lt "$timestamp" ]]; then
                        local delete_name=$(yq -y '.items['"$i"'].metadata.name' /tmp/check_deletion.yaml | head -qn 1)
                        local delete_kind=$(yq -y '.items['"$i"'].kind' /tmp/check_deletion.yaml | head -qn 1)
                        local delete_namespace_exists=$(yq -y '.items['"$i"'].metadata | has("namespace")' /tmp/check_deletion.yaml | head -qn 1)
                        if [[ "$delete_namespace_exists" =~ "false" ]]; then
                            kube delete --wait=false "$delete_kind" "$delete_name"
                        else
                            local delete_namespace=$(yq -y '.items['"$i"'].metadata.namespace' /tmp/check_deletion.yaml | head -qn 1)
                            kube delete --wait=false "$delete_kind" "$delete_name" -n "$delete_namespace"
                        fi
                    fi
                done
            fi
        else
            kube delete --wait=false "${ownedKinds}" -l "${filter}"
        fi
    fi
}

function createUpdateResources() {
    local owner="$1"
    local timestamp="$(date +%s)"
    # Check if directory contains only hidden files like .gitkeep, or .gitignore.
    # This would mean that user purposefully wanted to track an empty directory in git.
    # https://git.wiki.kernel.org/index.php/Git_FAQ#Can_I_add_empty_directories.3F
    if [[ -z $(ls "${MANIFEST_DIR}") ]]; then
        echo "Manifest directory empty, skipping"
        return
    elif [[ -z $(find "$MANIFEST_DIR" -regextype posix-extended -iregex '.*\.(ya?ml|json)') ]]; then
        echo "ERROR - no files with .yaml, .yml, or .json extension in manifest directory"
        exit 1
    fi
    case "$CREATE_MODE" in
    Apply)
        addLabels "$owner" "$timestamp"
        appendResourceVersion
        kube apply -R -f "$MANIFEST_DIR"
        deleteByOldLabels "$owner" "$timestamp"
        ;;
    Create)
        kube create -R -f "$MANIFEST_DIR"
        ;;
    Delete)
        kube delete --wait=false -R -f "$MANIFEST_DIR"
        ;;
    Patch)
        kube patch -R -f "$MANIFEST_DIR"
        ;;
    Replace)
        kube replace -R -f "$MANIFEST_DIR"
        ;;
    None) ;;
    esac
}

echo "Managing Resources"
setContext
# NOTE: Kubernetes currently requires that first *and last* character of
# label values are alphanumerical - we're adding the "own" prefix & suffix to
# ensure that. Also, Kubernetes requires it to be <=63 chars long, so we're
# taking a MD5 hash of actual name (MD5 hash is 33 chars long).
# See: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
owner="own.$(echo "$NAMESPACE $GITOPSCONFIG_NAME" | md5sum | awk '{print$1}').own"
case "$ACTION" in
create) createUpdateResources "$owner" ;;
delete) deleteByOldLabels "$owner" ;;
esac
