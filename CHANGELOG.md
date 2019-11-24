## v0.0.5 (Unreleased)

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
