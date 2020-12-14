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
from pathlib import Path
import subprocess
import sys
import tempfile
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

if __name__ == "__main__":
    logging.basicConfig(stream=sys.stdout, level=logging.INFO)
    logging.info("starting appendResourceVersion")
    manifest_dir = os.getenv('MANIFEST_DIR')
    with open('/var/run/secrets/kubernetes.io/serviceaccount/token') as x: token = x.read()
    files = list(Path(manifest_dir).rglob("*.yml")) + list(Path(manifest_dir).rglob("*.yaml")) + list(Path(manifest_dir).rglob("*.json"))
    for filename in files:
        logging.info("processing file {}".format(filename))
        try:
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
            if data is None:
                logging.error("no kubectl get output for file {}".format(filename))
                continue
            if "kind" in data and data["kind"] == "List":
                logging.info("file {} contains a list".format(filename))
                if "items" not in data:
                    logging.info("file {} has list with zero items".format(filename))
                    continue
                resource_version = {}
                for item in data["items"]:
                    if "metadata" in item and "resourceVersion" in item["metadata"]:
                        gvk_name = item["apiVersion"] + item["kind"] + item["metadata"]["name"]
                        resource_version[gvk_name] = item["metadata"]["resourceVersion"]
                        logging.info("got resource version {} for {}".format(item["metadata"]["resourceVersion"], gvk_name))
                    else:
                        logging.error("failed to get resource version for file {}".format(filename))
                        continue
                with open(filename, 'r+') as stream:
                    try:
                        new_docs = []
                        docs = yaml.safe_load_all(stream)
                        for doc in docs:
                            gvk_name = doc["apiVersion"] + doc["kind"] + doc["metadata"]["name"]
                            if gvk_name in resource_version:
                                doc["metadata"]["resourceVersion"] = resource_version[gvk_name]
                            else:
                                logging.error("failed to patch resource version for {} in file {}".format(gvk_name, filename))
                            new_docs.append(doc)
                        stream.seek(0)
                        stream.truncate()
                        yaml.safe_dump_all(new_docs, stream, explicit_start=True)
                    except yaml.YAMLError as exc:
                        logging.error("fatal error", exc_info=True)
            else:
                logging.info("file {} contains a {}".format(filename, data["kind"]))
                if "metadata" in data and "resourceVersion" in data["metadata"]:
                    with open(filename, 'r+') as stream:
                        try:
                            with_resource_version = yaml.safe_load(stream)
                            logging.info("got resource version {} for file {}".format(data["metadata"]["resourceVersion"], filename))
                            with_resource_version["metadata"]["resourceVersion"] = data["metadata"]["resourceVersion"]
                            stream.seek(0)
                            stream.truncate()
                            yaml.safe_dump(with_resource_version, stream)
                        except yaml.YAMLError as exc:
                            logging.error("fatal error", exc_info=True)
                else:
                    logging.error("failed to patch resource version for file {}".format(filename))
        except yaml.YAMLError as exc:
            logging.error("fatal error", exc_info=True)
