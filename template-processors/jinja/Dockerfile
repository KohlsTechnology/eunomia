FROM quay.io/kohlstechnology/eunomia-base:latest

USER root
RUN pip install j2cli[yaml]

COPY bin/processTemplates.sh /usr/local/bin/processTemplates.sh

USER ${USER_UID}
