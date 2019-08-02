FROM quay.io/kohlstechnology/eunomia-base:latest

ENV kubectl=oc

USER root

RUN curl -ksL https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit.tar.gz | tar --strip-components 1 --directory /usr/bin -zxv 

COPY bin/processTemplates.sh /usr/local/bin/processTemplates.sh

USER ${USER_UID}
