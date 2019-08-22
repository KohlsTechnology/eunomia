#!/usr/bin/env bash

set -o nounset
set -o errexit

cd $CLONED_TEMPLATE_GIT_DIR/
ansible-galaxy install -r requirements.yml -p galaxy
ansible-playbook -i .applier/ galaxy/openshift-applier/playbooks/openshift-cluster-seed.yml
