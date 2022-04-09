# Copyright 2019 Orange
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# 	You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# 	See the License for the specific language governing permissions and
# limitations under the License.

################################################################################

# Name of this service/application
SERVICE_NAME := casskop

BUILD_FOLDER = .
MOUNTDIR = $(PWD)

BOOTSTRAP_IMAGE ?= ghcr.io/cscetbon/casskop-bootstrap:0.1.9
TELEPRESENCE_REGISTRY ?= datawire
KUBESQUASH_REGISTRY:=


VERSION ?= $(cat version/version.go | grep "Version =" | cut -d\   -f3)
# Default bundle image tag
BUNDLE_IMG ?= controller-bundle:$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

KUBECONFIG ?= ~/.kube/config

# The default action of this Makefile is to build the development docker image
default: build

.DEFAULT_GOAL := help
help:	
	@grep -E '(^[a-zA-Z_-]+:.*?##.*$$)|(^##)' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}{printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m##/[33m/'

clean:
	@rm -rf $(OUT_BIN) || true
	@rm -f apis/cassandracluster/v2/zz_generated.deepcopy.go || true


FIRST_VERSION = .spec.versions[0]
SPEC_PROPS = $(FIRST_VERSION).schema.openAPIV3Schema.properties.spec.properties

.PHONY: update-crds
update-crds:
	echo Update CRD - Remove protocol and set config type to object CRD
	@sed -i '/\- protocol/d' config/crd/bases/db.orange.com_cassandraclusters.yaml
	@yq -i e '$(SPEC_PROPS).config.type = "object"' config/crd/bases/db.orange.com_cassandraclusters.yaml
	@yq -i e '$(SPEC_PROPS).topology.properties.dc.items.properties.config.type = "object"' config/crd/bases/db.orange.com_cassandraclusters.yaml
	@yq -i e '$(SPEC_PROPS).topology.properties.dc.items.properties.rack.items.properties.config.type = "object"' config/crd/bases/db.orange.com_cassandraclusters.yaml
	for crd in config/crd/bases/*.yaml; do \
		crdname=$$(basename $$crd); \
		end=$$(expr $$(grep -n ^status $$crd|cut -f1 -d:) - 1); \
		cat $$(echo v1-crds/$$crdname|sed 's/.yaml/_crd.yaml/') > /tmp/$$crdname; \
		sed -e '1,/versions/d' -e "1,$${end}s/^..//" $$crd >> /tmp/$$crdname; \
		cp /tmp/$$crdname $$crd; \
		yq -i e '$(FIRST_VERSION).storage = false' $$crd; \
	done
	for chart in $(ls charts); do \
	  cp -v config/crd/bases/* charts/${chart}/crds/; \
	done

include shared.mk

define debug_telepresence
	export TELEPRESENCE_REGISTRY=$(TELEPRESENCE_REGISTRY) ; \
	echo "execute : cat casskop.env" ; \
	sudo mkdir -p /var/run/secrets/kubernetes.io ; \
	sudo ln -s /tmp/known/var/run/secrets/kubernetes.io/serviceaccount /var/run/secrets/kubernetes.io/ || true ; \
	tdep=$(shell kubectl get deployment -l app=casskop -o jsonpath='{.items[0].metadata.name}') ; \
  	echo kubectl get deployment -l app=casskop -o jsonpath='{.items[0].metadata.name}' ; \
	echo telepresence --swap-deployment $$tdep --mount=/tmp/known --env-file casskop.env $1 $2 ; \
  	telepresence --swap-deployment $$tdep --mount=/tmp/known --env-file casskop.env $1 $2
endef

debug-telepresence:
	$(call debug_telepresence)

debug-kubesquash:
	kubesquash --container-repo $(KUBESQUASH_REGISTRY)

# Run the development environment (in local go env) in the background using local ~/.kube/config
run:
	export POD_NAME=casskop; \
	go run ./main.go

#Generate dep for graph
UNAME := $(shell uname -s)

dep-graph:
ifeq ($(UNAME), Darwin)
	dep status -dot | dot -T png | open -f -a /Applications/Preview.app
endif
ifeq ($(UNAME), Linux)
	dep status -dot | dot -T png | display
endif

dgoss-bootstrap:
	 IMAGE_TO_TEST=$(BOOTSTRAP_IMAGE) ./docker/bootstrap/dgoss/runChecks.sh

configure-psp:
	kubectl get clusterrole psp:cassie -o yaml
	kubectl -n cassandra get rolebindings.rbac.authorization.k8s.io psp:sa:cassie -o yaml
	kubectl -n cassandra get rolebindings.rbac.authorization.k8s.io psp:sa:cassie -o yaml | grep -vE '(annotations|creationTimestamp|resourceVersion|uid|selfLink|last-applied-configuration)' | sed 's/cassandra/cassandra-e2e/' | kubectl apply -f -

# Usage example:
# REPLICATION_FACTOR=3 make cassandra-stress small
#
ifeq (cassandra-stress,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  STRESS_TYPE := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(STRESS_TYPE):;@:)
endif

REPLICATION_FACTOR ?= 1
DC ?= dc1
USERNAME ?= cassandra
PASSWORD ?= cassandra

.PHONY: cassandra-stress
cassandra-stress:
	kubectl delete configmap cassandra-stress-$(STRESS_TYPE) || true
	cp cassandra-stress/$(STRESS_TYPE)_stress.yaml /tmp/
	echo Using replication factor $(REPLICATION_FACTOR) with DC $(DC) in cassandra-stress profile file
	sed -i -e "s/'dc1': '3'/'$(DC)': '$(REPLICATION_FACTOR)'/" /tmp/$(STRESS_TYPE)_stress.yaml
	kubectl create configmap cassandra-stress-$(STRESS_TYPE) --from-file=/tmp/$(STRESS_TYPE)_stress.yaml
	kubectl delete -f cassandra-stress/cassandra-stress-$(STRESS_TYPE).yaml --wait=false || true
	while kubectl get pod cassandra-stress-$(STRESS_TYPE)>/dev/null; do echo -n "."; sleep 1 ; done
	cp cassandra-stress/cassandra-stress-$(STRESS_TYPE).yaml /tmp/
	sed -i -e 's/user=[a-zA-Z]* password=[a-zA-Z]*/user=$(USERNAME) password=$(PASSWORD)/' /tmp/cassandra-stress-$(STRESS_TYPE).yaml
ifdef CASSANDRA_IMAGE
	echo "using Cassandra image $(CASSANDRA_IMAGE)"
	sed -i -e 's#image:.*#image: $(CASSANDRA_IMAGE)#g' /tmp/cassandra-stress-$(STRESS_TYPE).yaml
endif

ifdef CASSANDRA_NODE
	sed -i -e 's/cassandra-demo/$(CASSANDRA_NODE)/g' /tmp/cassandra-stress-$(STRESS_TYPE).yaml
else
  ifneq ($(and $(CLUSTER_NAME),$(DC),$(RACK)),)
	sed -i -e 's/cassandra-demo/$(CLUSTER_NAME)-$(DC)-$(RACK)-0.$(CLUSTER_NAME)/g' /tmp/cassandra-stress-$(STRESS_TYPE).yaml
  endif

  ifdef CLUSTER_NAME
	sed -i -e 's/cassandra-demo/$(CLUSTER_NAME)/g' /tmp/cassandra-stress-$(STRESS_TYPE).yaml
  endif
endif

ifdef CONSISTENCY_LEVEL
	sed -i -e 's/cl=one/cl=$(CONSISTENCY_LEVEL)/g' /tmp/cassandra-stress-$(STRESS_TYPE).yaml
endif

	cat /tmp/cassandra-stress-$(STRESS_TYPE).yaml
	kubectl apply -f /tmp/cassandra-stress-$(STRESS_TYPE).yaml

# Generate bundle manifests and metadata, then validate generated files.
bundle: generate
	operator-sdk generate kustomize manifests -q;\
	VERSION=$$(cat ./version/version.go | grep -Po '(?<=Version =\s").*(?=")');\
	make kustomize; \
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $${VERSION} $(BUNDLE_METADATA_OPTS);\
	operator-sdk bundle validate ./bundle;\

# Build the bundle image.
bundle-build:
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

