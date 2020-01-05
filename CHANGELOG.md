## v0.1.0 (January 2, 2020)
Starting with this release the v1alpha1 API is complete. Going forward changes will not be made to the v1alpha1 API. Work on the v1alpha2 API will start soon.

FEATURES:
* Helm install now supports k8s Ingress for the GitHub webhook [GH-225](https://github.com/KohlsTechnology/eunomia/pull/225)
* Support passing arguments to template processors [GH-229](https://github.com/KohlsTechnology/eunomia/pull/229)

CHANGES:
* **BREAKING** Remove resource handling mode CreateOrMerge [GH-223](https://github.com/KohlsTechnology/eunomia/pull/223)
* Improve local development workflow [GH-228](https://github.com/KohlsTechnology/eunomia/pull/228)

BUG FIXES:
* Fix documentation for using SSH keys to access private repos [GH-227](https://github.com/KohlsTechnology/eunomia/pull/227)
* Set CPU and memory requests and limits for operator [GH-224](https://github.com/KohlsTechnology/eunomia/pull/224)
* Fix issues with resource deletion [GH-52](https://github.com/KohlsTechnology/eunomia/pull/52)
* Fix typo in CRD description [GH-97](https://github.com/KohlsTechnology/eunomia/pull/97)

## v0.0.6 (December 19, 2019)
FEATURES:
* Create k8s events on Job/CronJob success and failure [GH-212](https://github.com/KohlsTechnology/eunomia/pull/212)

CHANGES:
* **BREAKING** Remove resource handling mode CreateOrUpdate [GH-149](https://github.com/KohlsTechnology/eunomia/pull/149)

BUG FIXES:
* Allow empty hierarchy directories [GH-198](https://github.com/KohlsTechnology/eunomia/pull/198)
* Do not constantly spawn new k8s jobs [GH-209](https://github.com/KohlsTechnology/eunomia/pull/209)
* Fix issue accessing private git repos [GH-195](https://github.com/KohlsTechnology/eunomia/pull/195)
* Fix k8s label length limit [GH-207](https://github.com/KohlsTechnology/eunomia/pull/207)

## v0.0.5 (December 5, 2019)
BUG FIXES:
* Multiple documentation fixes [GH-187](https://github.com/KohlsTechnology/eunomia/pull/187) [GH-191](https://github.com/KohlsTechnology/eunomia/pull/191) [GH-192](https://github.com/KohlsTechnology/eunomia/pull/192)

FEATURES:
* Delete k8s objects when removed from git repo [GH-157](https://github.com/KohlsTechnology/eunomia/pull/157)
* Add new openshift-provision template processor [GH-147](https://github.com/KohlsTechnology/eunomia/pull/147)
* Add initial fields to the GitOpsConfig status section [GH-163](https://github.com/KohlsTechnology/eunomia/pull/163)

CHANGES:
* Allow scheduling operator pod on all OpenShift nodes [GH-170](https://github.com/KohlsTechnology/eunomia/pull/170)
* Default Job and CronJob templates are not ConfigMaps anymore [GH-177](https://github.com/KohlsTechnology/eunomia/pull/177)

## v0.0.4 (November 21, 2019)
BUG FIXES:
* Fix proxy support in base template processor [GH-155](https://github.com/KohlsTechnology/eunomia/pull/155)
* Improve error handling and logging in base template processor [GH-151](https://github.com/KohlsTechnology/eunomia/pull/151)

FEATURES:
* Support Ansible htpasswd module in applier template processor [GH-159](https://github.com/KohlsTechnology/eunomia/pull/159)
* Initial support for parameters variable hierarchy [GH-156](https://github.com/KohlsTechnology/eunomia/pull/156)

CHANGES:
* Document Prometheus metrics support [GH-150](https://github.com/KohlsTechnology/eunomia/pull/150)

## v0.0.3 (November 7, 2019)
BUG FIXES:
* Fix k8s finalizer [GH-128]

FEATURES:
* n/a

IMPROVEMENTS:
* Automate helm releases to GitHub pages [GH-120]

CHANGES:
* Pin kubectl to 1.11.x in base template processor [GH-140]

## v0.0.2 (October 30, 2019)
* Fix release automation
* Add OpenShift Applier template processor
* Use latest Go Lang 1.12.x version
* Switch to UBI 8 base container image
* Add step to merge YAML files
* Remove DEFAULT_ROUTE_DOMAIN and REGISTRY_ROUTE environment variables

## v0.0.1 (August 28, 2019)
* Initial release
* Dead on arrival, do not use
