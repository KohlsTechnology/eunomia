FROM quay.io/kohlstechnology/eunomia-ocp-templates:latest

USER root
RUN PACKAGES=".build-deps gcc musl-dev python-dev libffi-dev openssl-dev libpthread-stubs" \
  && apk add --no-cache --virtual $PACKAGES \
  && pip install ansible openshift==0.9.0 kubernetes==9.0.0 jmespath \
  && apk del $PACKAGES 

COPY bin/processTemplates.sh /usr/local/bin/processTemplates.sh

USER ${USER_UID}
