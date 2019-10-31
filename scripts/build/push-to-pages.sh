#!/usr/bin/env bash
set -euxo pipefail

git -C "$GITHUB_PAGES_DIR" config user.email "travis@users.noreply.github.com"
git -C "$GITHUB_PAGES_DIR" config user.name travis
git -C "$GITHUB_PAGES_DIR" config core.sshCommand 'ssh -i github_deploy_key'
git -C "$GITHUB_PAGES_DIR" add .
git -C "$GITHUB_PAGES_DIR" status
git -C "$GITHUB_PAGES_DIR" commit -m "Published by Travis"
git -C "$GITHUB_PAGES_DIR" push origin "$GITHUB_PAGES_BRANCH"