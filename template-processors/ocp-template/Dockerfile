FROM quay.io/kohlstechnology/eunomia-base:latest

ENV OC_VERSION=4.1.8

USER root

RUN curl -O http://mirror.openshift.com/pub/openshift-v4/clients/ocp/${OC_VERSION}/openshift-client-linux-${OC_VERSION}.tar.gz \
  && tar --directory /usr/bin -zxvf openshift-client-linux-${OC_VERSION}.tar.gz oc \
  && ln -s /lib /lib64 \
  && ln -s /lib64/ld-musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY bin/processTemplates.sh /usr/local/bin/processTemplates.sh

USER ${USER_UID}
