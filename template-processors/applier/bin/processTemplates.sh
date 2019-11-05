#!/usr/bin/env bash

set -euxo pipefail

cd $CLONED_TEMPLATE_GIT_DIR/
ansible-galaxy install -r requirements.yml -p galaxy
ansible-playbook -i .applier/ galaxy/openshift-applier/playbooks/openshift-cluster-seed.yml
