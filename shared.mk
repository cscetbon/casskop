UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	GOOS = linux
endif
ifeq ($(UNAME_S),Darwin)
	GOOS = darwin
endif

ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

ifeq (, $(shell which controller-gen))
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

ifeq (, $(shell which kustomize))
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

CONTROLLER_GEN_OPTIONS=crd paths=./api/... output:dir=./config/crd/bases schemapatch:manifests=./config/crd/bases

# Shell to use for running scripts
SHELL := $(shell which bash)

# Generate code
generate-k8s:
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

OPERATOR_SDK_VERSION=v1.13.0
# workdir
WORKDIR := /go/casskop

.PHONY: generate
generate:
	echo "Generate zzz-deepcopy objects"
	$(MAKE) generate-k8s
	@rm -f */crds/*
	$(CONTROLLER_GEN) $(CONTROLLER_GEN_OPTIONS)
	$(MAKE) update-crds

# Build the bundle image.
bundle-build:
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .
