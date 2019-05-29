FROM quay.io/kohlstechnology/eunomia-base:v0.0.1

ENV kubectl=oc

USER root
RUN dnf remove -y kubectl && \
    dnf install -y origin-clients
COPY bin/processTemplates.sh /usr/local/bin/processTemplates.sh

USER ${USER_UID}
