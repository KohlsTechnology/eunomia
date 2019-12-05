## v0.0.6 (Unreleased)

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
