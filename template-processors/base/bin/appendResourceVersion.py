#!/usr/bin/env python3

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

import logging
import os
import subprocess
import yaml
# appendResourceVersion - patches the YAML&JSON files in $MANIFEST_DIR,
# adding the metadata.resourceVersion for each resource being managed.
# This is intended to serve as a locking mechanism when applying resources
# in which Kubernetes will fail the apply with a StatusConflict (HTTP status code 409)
# Ref https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
#
# Inputs:
#
# MANIFEST_DIR environment variable
# pylint: disable=W1203
LOG = logging.getLogger(__name__)

def get_files(manifest_dir, file_types=(".yml", ".yaml", ".json")):
    '''
    Get files of file_type(s) from manifest dir
    '''
    files = []
    for file in os.listdir(manifest_dir):
        if file.endswith(file_types):
            files.append(os.path.join(manifest_dir, file))
    return files

def get_kube_token():
    '''
    Get kubernetes auth token
    '''
    with open('/var/run/secrets/kubernetes.io/serviceaccount/token') as token_file:
        token = token_file.read()
        return token

def get_kubectl_data(filename, token):
    '''
    Get data using kubectl
    '''
    data = yaml.safe_load(subprocess.run(["kubectl",
                                          "-s",
                                          "https://kubernetes.default.svc:443",
                                          "--token",
                                          token,
                                          "--certificate-authority",
                                          "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
                                          "get",
                                          "--ignore-not-found",
                                          "-f",
                                          filename,
                                          "-o",
                                          "yaml"], stdout=subprocess.PIPE).stdout)
    LOG.debug(f"kubectl output: {data}")
    return data

def process_list(resource_list, filename):
    '''
    Process list of resource items
    '''
    resource_name_version_dict = {}
    for item in resource_list["items"]:
        if "metadata" in item and "resourceVersion" in item["metadata"]:
            kube_custom_resource_name = item["apiVersion"] + item["kind"] + item["metadata"]["name"]
            if "namespace" in item["metadata"]:
                kube_custom_resource_name += item["metadata"]["namespace"]
            resource_version = item["metadata"]["resourceVersion"]
            resource_name_version_dict[kube_custom_resource_name] = resource_version
            logging.info(f"Got resource version {resource_version} for {kube_custom_resource_name}")
        else:
            logging.error(f"Failed to get resource version for file {filename}")
            #Item does not have resource version, continue to next item
            continue
        with open(filename, 'r+') as stream:
            try:
                new_docs = []
                #Read existing document
                docs = yaml.safe_load_all(stream)
                for doc in docs:
                    LOG.debug(f"Existing file: {doc}")
                    custom_resource_name = doc["apiVersion"] + doc["kind"] + doc["metadata"]["name"]
                    #If resource has namespace metadata append namespace to custom_resource_name so resource can be uniquely identified
                    #This is to resolve an issue with identifying cluster wide resources
                    if "namespace" in item["metadata"]:
                        custom_resource_name += doc["metadata"]["namespace"]
                    if custom_resource_name in resource_name_version_dict.keys():
                        LOG.debug(f"Overwrite resource version for {custom_resource_name}")
                        doc["metadata"]["resourceVersion"] = resource_name_version_dict[custom_resource_name]
                    else:
                        LOG.info(f"Resource did not previously exist creating new resource {custom_resource_name} from file {filename}")
                    new_docs.append(doc)
                # Move pointer to beginning of file
                stream.seek(0)
                # Clear contents of file
                stream.truncate()
                # Dump resource with new version to file
                yaml.safe_dump_all(new_docs, stream, explicit_start=True)
                LOG.debug(f"New file: {new_docs}")
            except yaml.YAMLError as exc:
                LOG.error(f"Fatal error: {exc}", exc_info=True)



def process_files(token, files):
    '''
    Overwrite resource version in manifest file with existing metadata.resourceVersion from Kubernetes.
    This is intended to serve as a locking mechanism when applying resources
    in which Kubernetes will fail the apply with a StatusConflict (HTTP status code 409)
    Ref: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
    '''
    for filename in files:
        kube_data = get_kubectl_data(filename, token)
        if kube_data is None:
            LOG.error(f"No kubectl get output for file {filename}")
            #No kube_data to process break from for loop
            continue
        elif "kind" in kube_data and kube_data["kind"] == "List":
            if "items" in kube_data and not kube_data["items"]:
                LOG.info(f"File {filename} has list with zero items")
                #No kube_data to process break from for loop
                continue
            LOG.info(f"file {filename} contains a list")
            process_list(kube_data, filename)
        else:
            #not a list, do not need to loop through kube_data
            LOG.info(f"File {filename} contains a {kube_data['kind']}")
            if "metadata" in kube_data and "resourceVersion" in kube_data["metadata"]:
                with open(filename, 'r+') as stream:
                    try:
                        file_resource_version = yaml.safe_load(stream)
                        LOG.debug(f"Existing file, non-list: {file_resource_version}")
                        LOG.info(f"got resource version {kube_data['metadata']['resourceVersion']} for file {filename}")
                        # Overwrite file resource version with kube resource version, to prevent runtime conflict
                        file_resource_version["metadata"]["resourceVersion"] = kube_data["metadata"]["resourceVersion"]
                        # Move pointer to beginning of file
                        stream.seek(0)
                        # Clear contents of file
                        stream.truncate()
                        # Dump resource with new version to file
                        yaml.safe_dump(file_resource_version, stream)
                        LOG.debug(f"New file: {file_resource_version}")
                    except yaml.YAMLError as exc:
                        LOG.error(f"Fatal error: {exc}", exc_info=True)
            else:
                LOG.error(f"failed to patch resource version for file {filename}")

def main(manifest_dir):
    '''
    Main method
    '''
    #Get all manifest files
    files = get_files(manifest_dir)
    #Get token to call Kubernetes
    token = get_kube_token()
    #Overwrite resource version in manifest file with existing resource version from kubernetes
    process_files(token, files)

if __name__ == "__main__":
    logging.basicConfig(level=logging.DEBUG)
    LOG.info("Starting appendResourceVersion...")
    manifest_dir = os.getenv('MANIFEST_DIR')
    main(manifest_dir)
