FROM alpine:3.10

ENV USER_UID=1001 \
    USER_NAME=gitopsjob \
    kubectl=kubectl \
    KUBECTL_VERSION="v1.15.0" \
    YQ_VERSION="2.7.2" \
    GOLANG_YQ_VERSION="2.4.0"

COPY bin /usr/local/bin

RUN \
    apk add --no-cache bash curl ca-certificates git gettext jq findutils py-pip && \
    curl -kL https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl -o /usr/bin/kubectl && \
    chmod +x /usr/bin/kubectl && \
    pip install yq==${YQ_VERSION} && \
    curl -kL https://github.com/mikefarah/yq/releases/download/${GOLANG_YQ_VERSION}/yq_linux_amd64 -o /usr/bin/goyq && \
    chmod +x /usr/bin/goyq && \
    /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
