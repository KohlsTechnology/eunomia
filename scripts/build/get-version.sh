#!/usr/bin/env bash
set -euo pipefail

GIT_V_TAG=$(git tag --list 'v[0-9]*' --points-at HEAD)

if [ "$GIT_V_TAG" ]; then
    echo "$GIT_V_TAG"
else
    echo "dev-$(git rev-parse --short HEAD)"
fi

