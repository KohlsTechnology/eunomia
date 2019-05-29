FROM quay.io/kohlstechnology/eunomia-base:v0.0.1

USER root
RUN pip install j2cli[yaml]

COPY bin/processTemplates.sh /usr/local/bin/processTemplates.sh

USER ${USER_UID}
