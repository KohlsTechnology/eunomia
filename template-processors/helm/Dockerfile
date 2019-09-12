FROM quay.io/kohlstechnology/eunomia-base:latest

ENV HELM_VERSION="v2.14.3"

USER root
RUN curl -L https://storage.googleapis.com/kubernetes-helm/helm-${HELM_VERSION}-linux-amd64.tar.gz | tar --strip-components 1 --directory /usr/bin -zxv linux-amd64/helm

COPY bin/processTemplates.sh /usr/local/bin/processTemplates.sh

USER ${USER_UID}
