#!/usr/bin/env bash

set -o nounset
set -o errexit

## we assume in $CLONED_TEMPLATE_GIT_DIR there are a set of templates with a .j2 extension
## we assume that in $CLONED_PARAMETER_GIT_DIR there is a parameter file called parameters.yaml
## the result is a set of file stored in the $MANIFEST_DIR, with the same name as the templates, but with no extension. 

for file in $CLONED_TEMPLATE_GIT_DIR/*.j2 ; do
  shortfile=$(basename -- "$file")
  # TODO consider improving by adding this filter: lib/ansible/plugins/filters/core.py
  j2 $file $CLONED_PARAMETER_GIT_DIR/eunomia_values_processed.yaml  \
    --import-env env \
    > $MANIFEST_DIR/"${shortfile%.*}";
done
