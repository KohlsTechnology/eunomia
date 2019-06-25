FROM quay.io/kohlstechnology/eunomia-base:v0.0.1

USER root
RUN curl -ksL  https://storage.googleapis.com/kubernetes-helm/helm-v2.14.1-linux-amd64.tar.gz | tar --strip-components 1 --directory /usr/bin -zxv linux-amd64/helm

COPY bin/processTemplates.sh /usr/local/bin/processTemplates.sh

USER ${USER_UID}
