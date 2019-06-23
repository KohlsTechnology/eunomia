FROM alpine:3.10

ENV USER_UID=1001 \
    USER_NAME=gitopsjob \
    kubectl=kubectl

COPY bin /usr/local/bin

RUN \
    apk add --no-cache bash curl ca-certificates git gettext jq findutils py-pip && \
    curl -L https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl -o /usr/bin/kubectl && \
    chmod +x /usr/bin/kubectl && \
    pip install yq && \
    /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
