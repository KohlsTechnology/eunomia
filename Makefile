
# Image URL to use all building/pushing image targets
REGISTRY ?= quay.io
REPOSITORY ?= $(REGISTRY)/kohlstechnology

VERSION := $(shell ./scripts/build/get-version.sh)

BUILD_COMMIT := $(shell ./scripts/build/get-build-commit.sh)
BUILD_TIMESTAMP := $(shell ./scripts/build/get-build-timestamp.sh)
BUILD_HOSTNAME := $(shell ./scripts/build/get-build-hostname.sh)

LDFLAGS := "-X github.com/KohlsTechnology/eunomia/version.Version=$(VERSION) \
	-X github.com/KohlsTechnology/eunomia/version.Vcs=$(BUILD_COMMIT) \
	-X github.com/KohlsTechnology/eunomia/version.Timestamp=$(BUILD_TIMESTAMP) \
	-X github.com/KohlsTechnology/eunomia/version.Hostname=$(BUILD_HOSTNAME)"

all: manager

# Run tests
native-test: generate fmt vet
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o build/_output/bin/eunomia  -ldflags $(LDFLAGS) github.com/KohlsTechnology/eunomia/cmd/manager

# Build manager binary
manager-osx: generate fmt vet
	GOOS=darwin go build -o build/_output/bin/eunomia -ldflags $(LDFLAGS) github.com/KohlsTechnology/eunomia/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install:
	cat deploy/crds/*crd.yaml | kubectl apply -f-

e2e-test-images: manager
	TRAVIS_TAG=v999.0.0 ./scripts/build-images.sh ${REPOSITORY}

# Deploy images to Quay.io
travis-deploy-images: manager
	docker login -u ${DOCKER_USER} -p ${DOCKER_PASSWORD} ${REGISTRY}
	./scripts/build-images.sh ${REPOSITORY} true
