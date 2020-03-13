# Releasing Eunomia Version
This document provides the high level steps for releasing a new version of eunomia.

## Release to GitHub
1. Update file `version/version.go`
2. Update `CHANGELOG.md`
3. Create PR with above two changes
4. Create a GitHub release for the new version
5. Verify new container images were successfully pushed to quay.io
6. Send a message about the new release to the gitter channel

## Release to OperatorHub

### Prerequisites

For the release script to run successfully, you'll need the following:
- account on GitHub
- [operator-sdk](https://github.com/operator-framework/operator-sdk/blob/v0.8.1/doc/user/install-operator-sdk.md)
- [operator-courier](https://github.com/operator-framework/operator-courier#installation)
- [helm](https://helm.sh/docs/intro/install/)
- [hub](https://github.com/github/hub#installation)

### Release to OperatorHub

To push Eunomia to your fork of OperatorHub, run the script `scripts/operatorhub-deploy.sh`:

```shell
./scripts/operatorhub-deploy.sh VERSION
```

where `VERSION` is set to the version of Eunomia to be deployed to OperatorHub, e.g. `1.0.1`. You can now create a Pull Request manually.

Alternatively, if you pass the `--send-pr` flag:

```shell
./scripts/operatorhub-deploy.sh --send-pr VERSION
```

a Pull Request will be created automatically.
