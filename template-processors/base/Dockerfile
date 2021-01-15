FROM registry.access.redhat.com/ubi8/ubi:latest

ENV USER_UID=1001 \
    USER_NAME=gitopsjob \
    kubectl=kubectl \
    KUBECTL_VERSION="v1.19.5" \
    YQ_VERSION="2.11.1" \
    GOLANG_YQ_VERSION="3.3.2" \
    JQ_VERSION="1.6" \
    HIERARCHY_VERSION="0.1.1"


RUN yum install -y --disableplugin=subscription-manager git gettext python36-devel gcc python3-pip python3-setuptools && \
    curl https://github.com/stedolan/jq/releases/download/jq-${JQ_VERSION}/jq-linux64 -L -o /usr/bin/jq && \
    chmod +x /usr/bin/jq && \
    curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl -o /usr/bin/kubectl && \
    chmod +x /usr/bin/kubectl && \
    pip3 install yq==${YQ_VERSION} && \
    curl -L https://github.com/mikefarah/yq/releases/download/${GOLANG_YQ_VERSION}/yq_linux_amd64 -o /usr/bin/goyq && \
    chmod +x /usr/bin/goyq && \
    curl -L https://github.com/KohlsTechnology/hierarchy/releases/download/v${HIERARCHY_VERSION}/hierarchy_${HIERARCHY_VERSION}_Linux_x86_64.tar.gz | tar --directory /usr/bin -zxv hierarchy

COPY bin /usr/local/bin

RUN /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
