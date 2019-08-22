# OpenShift Applier Template Processor

This image allows for the use of [OpenShift Applier](https://github.com/redhat-cop/openshift-applier.git) with Eunomia.

This image can be pulled from quay.io with:

    docker pull quay.io/KohlsTechnology/eunomia-applier

Building the image locally can be done by running:

    docker build -t quay.io/KohlsTechnology/eunomia-applier template-processors/applier/

A full end-to-end example of using Eunomia and OpenShift Applier together can be found in the [Operationalizing OpenShift Lab](https://github.com/redhat-cop/operationalizing-openshift-lab).
