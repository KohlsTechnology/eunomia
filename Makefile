
# Image URL to use all building/pushing image targets
REGISTRY ?= quay.io
REPOSITORY ?= $(REGISTRY)/kohlstechnology

COMMIT := $(shell git rev-parse HEAD)
BRANCH := $(shell git symbolic-ref --short -q HEAD || echo HEAD)
DATE := $(shell date -u +%Y%m%d-%H:%M:%S)
VERSION_PKG = github.com/KohlsTechnology/eunomia/version
LDFLAGS := "-X ${VERSION_PKG}.Branch=${BRANCH} -X ${VERSION_PKG}.BuildDate=${DATE} \
	-X ${VERSION_PKG}.GitSHA1=${COMMIT}"

export GITHUB_PAGES_DIR ?= /tmp/helm/publish
export GITHUB_PAGES_BRANCH ?= gh-pages
export GITHUB_PAGES_REPO ?= KohlsTechnology/eunomia
export HELM_CHARTS_SOURCE ?= deploy/helm/eunomia-operator
export HELM_CHART_DEST ?= $(GITHUB_PAGES_DIR)
export TEST_ENV ?= minikube

.PHONY: all
all: build

.PHONY: clean
clean:
	rm -rf build/_output

generate:
	docker build ./scripts -f ./scripts/operator-sdk.docker -t 'operator-sdk:old'
	go mod vendor
	cp go.mod go.mod.bak
	docker run \
		-u "$(shell id -u)" \
		-v "$(shell go env GOCACHE):/gocache" \
		-v "$(PWD):/gopath/src/github.com/KohlsTechnology/eunomia" \
		-v "$(shell go env GOPATH | sed 's/:.*//' )/pkg:/gopath/pkg" \
		operator-sdk:old
	cp go.mod.bak go.mod  # undo changes made by go1.12 required for ancient operator-sdk.docker

# Build binary
.PHONY: build
build:
	go mod vendor
	go build -o build/_output/bin/eunomia -ldflags $(LDFLAGS) github.com/KohlsTechnology/eunomia/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run:
	go run ./cmd/manager/main.go

# Run some stuff that should be run before committing, then verify that there are no accidental modifications in the repo,
# which could result in different code being actually compiled than expected based on reading the source.
.PHONY: test-dirty
test-dirty: generate
	git diff --exit-code
	# TODO: also check that there are no untracked files, e.g. extra .go and .yaml ones

.PHONY: test
test: check-fmt lint vet check-shfmt shellcheck test-unit test-e2e

.PHONY: test-e2e
test-e2e:
	./scripts/e2e-test.sh

.PHONY: test-unit
test-unit:
	go test -v -coverprofile=coverage.txt ./...

# Install CRDs into a cluster
.PHONY: install
install:
	cat deploy/crds/*crd.yaml | kubectl apply -f-

# Check if gofmt against code is clean
.PHONY: check-fmt
check-fmt:
	test -z "$(shell gofmt -l . | grep -v ^vendor)"

.PHONY: lint
lint:
	LINT_INPUT="$(shell go list ./... | grep -v /vendor/)"; golint -set_exit_status $$LINT_INPUT

# Run go vet against code
.PHONY: vet
vet:
	VET_INPUT="$(shell go list ./... | grep -v /vendor/)"; go vet $$VET_INPUT

# TODO: improve the command to also check scripts outside the scripts and template-processors dirs
.PHONY: check-shfmt
check-shfmt:
	shfmt -i 4 -d ./scripts
	shfmt -i 4 -d ./template-processors

# TODO: improve the command to also check scripts without .sh extension
# get_helm.sh is ignored because it is getting download from internet
.PHONY: shellcheck
shellcheck:
	shellcheck $(shell find . -not -path "./vendor/*" -not -path "./get_helm.sh" -not -path "./ci/*" -name "*.sh")

.PHONY: e2e-test-images
e2e-test-images: build
	TRAVIS_TAG=latest ./scripts/build-images.sh ${REPOSITORY}

# Deploy images to Quay.io
.PHONY: travis-deploy-images
travis-deploy-images: build
	docker login -u ${DOCKER_USER} -p ${DOCKER_PASSWORD} ${REGISTRY}
	./scripts/build-images.sh ${REPOSITORY} true

.PHONY: publish-chart-repo
publish-chart-repo:
	./scripts/build/publish-chart-repo.sh

.PHONY: travis-release
travis-release: travis-deploy-images publish-chart-repo
