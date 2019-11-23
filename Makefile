
# Image URL to use all building/pushing image targets
REGISTRY ?= quay.io
REPOSITORY ?= $(REGISTRY)/kohlstechnology

BUILD_COMMIT := $(shell ./scripts/build/get-build-commit.sh)
BUILD_TIMESTAMP := $(shell ./scripts/build/get-build-timestamp.sh)
BUILD_HOSTNAME := $(shell ./scripts/build/get-build-hostname.sh)

export GITHUB_PAGES_DIR ?= /tmp/helm/publish
export GITHUB_PAGES_BRANCH ?= gh-pages
export GITHUB_PAGES_REPO ?= KohlsTechnology/eunomia
export HELM_CHARTS_SOURCE ?= deploy/helm/eunomia-operator
export HELM_CHART_DEST ?= $(GITHUB_PAGES_DIR)

LDFLAGS := "-X github.com/KohlsTechnology/eunomia/version.Vcs=$(BUILD_COMMIT) \
	-X github.com/KohlsTechnology/eunomia/version.Timestamp=$(BUILD_TIMESTAMP) \
	-X github.com/KohlsTechnology/eunomia/version.Hostname=$(BUILD_HOSTNAME)"

.PHONY: all build

all: build

clean:
	rm -rf build/_output

# Build binary
build:
	GO111MODULE=on go mod vendor
	GO111MODULE=on go build -o build/_output/bin/eunomia -ldflags $(LDFLAGS) github.com/KohlsTechnology/eunomia/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run:
	go run ./cmd/manager/main.go

test: fmt lint vet test-unit test-e2e

test-e2e:
	./scripts/e2e-test.sh

test-unit:
	./scripts/unit-tests.sh

# Install CRDs into a cluster
install:
	cat deploy/crds/*crd.yaml | kubectl apply -f-

# Run gofmt against code
fmt:
	test -z "$(shell gofmt -l . | grep -v ^vendor)"

lint:
	LINT_INPUT="$(shell go list ./... | grep -v /vendor/)"; golint -set_exit_status $$LINT_INPUT

# Run go vet against code
vet:
	VET_INPUT="$(shell go list ./... | grep -v /vendor/)"; GO111MODULE=on go vet $$VET_INPUT

e2e-test-images: build
	TRAVIS_TAG=v999.0.0 ./scripts/build-images.sh ${REPOSITORY}

# Deploy images to Quay.io
travis-deploy-images: build
	docker login -u ${DOCKER_USER} -p ${DOCKER_PASSWORD} ${REGISTRY}
	./scripts/build-images.sh ${REPOSITORY} true

publish-chart-repo:
	./scripts/build/publish-chart-repo.sh

travis-release: travis-deploy-images publish-chart-repo
