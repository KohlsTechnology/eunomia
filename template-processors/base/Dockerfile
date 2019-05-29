FROM fedora

ENV USER_UID=1001 \
    USER_NAME=gitopsjob \
    kubectl=kubectl
 
COPY kubernetes.repo /etc/yum.repos.d/kubernetes.repo 

RUN dnf install -y git kubectl gettext python-pip jq findutils && \
    pip install yq

COPY bin /usr/local/bin
RUN /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}