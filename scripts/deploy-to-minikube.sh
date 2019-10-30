#!/usr/bin/env bash

#This scripts helps with deploying images locally to minikube docker registry.

set -euxo pipefail

eval $(minikube docker-env)
export TRAVIS_TAG=latest
"$(dirname "$0")/build-images.sh" quay.io/kohlstechnology
