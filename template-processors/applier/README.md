# OpenShift Applier Template Processor

This image allows for the use of [OpenShift Applier](https://github.com/redhat-cop/openshift-applier.git) with [Eunomia](https://github.com/KohlsTechnology/eunomia.git), a GitOps Operator for Kubernetes.

This image can be pulled from quay.io with:

    docker pull quay.io/redhat-cop/eunomia-applier

Building the image locally can be done by running:

    docker build -t quay.io/redhat-cop/eunomia-applier images/eunomia-applier/

A full end-to-end example of using Eunomia and OpenShift Applier together can be found in the [Operationalizing OpenShift Lab](https://github.com/redhat-cop/operationalizing-openshift-lab).
