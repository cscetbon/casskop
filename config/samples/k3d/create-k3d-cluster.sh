#!/bin/bash

CLUSTER=local-casskop
k3d cluster delete $CLUSTER
k3d cluster create --image rancher/k3s:v1.20.15-k3s1 $CLUSTER
. $(dirname $0)/setup-requirements.sh
