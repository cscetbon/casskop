---
title: Operations Issues
sidebar_label: Operations Issues
---

## Operator can't perform the Action

If you ask to scale up or add a new DC, or ask for more resources, CassKop will ask Kubernetes to schedule as you
requested.
But sometimes it is not possible to achieve the change because of a lack of resources (memory/cpus) or because
constraints can't be satisfied (Kubernetes nodes with specific labels available...)

CassKop make uses of the PodDisruptionBudget to prevent CassKop to make some change on the CassandraCluster that could
make more than 1 Cassandra node at a time. 

If you have a Pod stuck in **pending** state, then you have at least 1 Pod in Disruption, and the PDB object will
prevent you to make changes on statefulset because that mean that you will have more than 1 Cassandra down at a time.

```console
$ kubectl get poddisruptionbudgets
NAME             MIN AVAILABLE   MAX UNAVAILABLE   ALLOWED DISRUPTIONS   AGE
cassandra-demo   N/A             1                 0                     12m
```

The Operator logs this line when there is disruption on the Cassandra cluster:

```logs
INFO[3037] Cluster has Disruption on Pods, we wait before applying any change to statefulset  cluster=cassandra-demo dc-rack=dc1-rack1
```


### Can't ScaleUp

In this example I ask a ScaleUp but it can't perform :

```console
$ kubectl get pods
NAME                                                              READY   STATUS    RESTARTS   AGE
cassandra-demo-dc1-rack1-0                                        1/1     Running   0          16h
cassandra-demo-dc1-rack1-1                                        0/1     Pending   0          12m
cassandra-demo-dc1-rack2-0                                        1/1     Running   0          16h
cassandra-demo-dc1-rack3-0                                        1/1     Running   0          16h
```

the cassandra-demo-dc1-rack1-1 pod is Pending and can't be scheduled.

If we looked at the pod status we will see this message :

```
Warning  FailedScheduling   51s (x17 over 14m)    default-scheduler   0/6 nodes are available: 4 Insufficient cpu, 4 node(s) didn't match node selector.
```

Kubernetes can't find any Pod with sufficient cpu and matching kubernetes nodes labels we asked in the topology section.

To fix this, we can either:
- reduce memory/cpu limits
- add more kubernetes nodes that will satisfied our requirements.
- rollback the scaleUp Operation

> At this point, CassKop will wait indefinitely to the case to be Fix


#### Rollback ScaleUp operation

In order to rollback the operation, we need to revert the change on the `nodesPerRacks` parameter.

> This is not sufficient

Because CassKop is actually performing another action on the cluster (ScaleUp) we can't scheduled a new operation to
rollback since it has not finished.
We introduced a new parameter in the CRD to allow such changes when all the pods can't be scheduled:
- `Spec.unlockNextOperation: true`

> _**:triangular_flag_on_post: Warning** This is not a regular parameter and it must be used with very good care!!._


By adding this parameter in our cluster definition, CassKop will allow to trigger a new operation.

> Once CassKop has scheduled the new operation, it will reset this parameter to the default `false` value.
> value. If you need more operation, you will need to reset the parameter to force another Operation.
> Keep in mind that CassKop is mean to do only 1 operation at a time.


If this is not already done, you can now rollback the scaleUp updating `nodesPerRacks: 1`

```
WARN[3348] ScaleDown detected on a pending Pod. we don't launch decommission  cluster=cassandra-demo pod=cassandra-demo-dc1-rack1-1 rack=dc1-rack1
INFO[3350] Cluster has 1 Pod Disrupted but that may be normal as we are decommissioning  cluster=cassandra-demo
dc-rack=dc1-rack1
...
INFO[3354] [cassandra-demo][dc1-rack1]: StatefulSet(ScaleDown): Replicas Number OK: ready[1] 
INFO[3354] ScaleDown not yet Completed: Waiting for Pod operation to be Done  cluster=cassandra-demo rack=dc1-rack1
INFO[3354] Decommission done -> we delete PVC            cluster=cassandra-demo pvc=data-cassandra-demo-dc1-rack1-1 rack=dc1-rack1
INFO[3354] PVC deleted                                   cluster=cassandra-demo pvc=data-cassandra-demo-dc1-rack1-1 rack=dc1-rack1
DEBU[3354] Waiting Rack to be running before continuing, we break ReconcileRack Without Updating Statefulset  cluster=cassandra-demo dc-rack=dc1-rack1 err="<nil>"
INFO[3354] ScaleDown is Done                             cluster=cassandra-demo rack=dc1-rack1
```


### Can't add new rack in new DC

In this example, I ask to add dc called dc2

```
kubectl get pods
NAME                                                              READY   STATUS    RESTARTS   AGE
cassandra-demo-dc1-rack1-0                                        1/1     Running   0          17h
cassandra-demo-dc1-rack2-0                                        1/1     Running   0          17h
cassandra-demo-dc1-rack3-0                                        1/1     Running   0          17h
cassandra-demo-dc2-rack1-0                                        1/1     Running   0          5m46s
cassandra-demo-dc2-rack2-0                                        1/1     Running   0          4m20s
cassandra-demo-dc2-rack3-0                                        0/1     Pending   0          2m37s
```

But the last one can't be scheduled because of insufficient cpu on k8s nodes.

We can either add the wanted resources in the k8s cluster or make a rollback.

#### Solution1: rollback adding the DC

To rollback the add of the new DC, we first need to scale down to 0 for the nodes that already have join the ring.
We need to allow disruption as we do in previous section.

then We first need to ask the dc2 to scaleDown to 0 because it has already add 2 racks, and we add the
spec.unlockNextOperation to true.

```
...
  unlockNextOperation: true
  ...
    name: dc2
    nodesPerRacks: 0
```

This will allow CassKop to make the scale down. because it will start with the first rack, it will free some space and
the last pod which was pending will be joining. then it will be decommissioned by CassKop.

we can see in CassKop logs when it deal with the rack with unscheduled pods:

```
WARN[0667] Aborting Initializing..., start ScaleDown                      cluster=cassandra-demo rack=dc2-rack3
INFO[0667] The Operator Waits 20 seconds for the action to start correctly  cluster=cassandra-demo rack=dc2-rack3
WARN[0667] ScaleDown detected on a pending Pod. we don't launch decommission  cluster=cassandra-demo pod=cassandra-demo-dc2-rack3-0 rack=dc2-rack3
INFO[0667] Cluster has 1 Pod Disrupted but that may be normal as we are decommissioning
cluster=cassandra-demodc-rack=dc2-rack3
INFO[0667] Template is different:  {...}
```

It will also ScaleDown any pods that was part of the new DC.

> Once ScaleDown is done, you can delete the DC from the Spec.

#### Solution2: change the topology for dc2 (remove the 3rd unschedulable rack)

we get back in the previous section before making the rolling back of adding the dc2.

If only one of the Racks can't schedule any pods, we can change the topology to remove the Rack ONLY if there was not already
pods deployed in the Rack. If this is not the case, then you will need to process ScaleDown instead of removing rack.

```
    - labels:
        failure-domain.beta.kubernetes.io/region: europe-west1
      name: dc2
      nodesPerRacks: 1
      config:
        cassandra-yaml:
          num_tokens: 256
      rack:
      - labels:
          failure-domain.beta.kubernetes.io/zone: europe-west1-d
        name: rack1
      - labels:
          failure-domain.beta.kubernetes.io/zone: europe-west1-c
        name: rack2
      - labels:
          failure-domain.beta.kubernetes.io/zone: europe-west1-d
        name: rack3
```

let's remove the rack3 from dc2.

the operator will log :

```
WARN[0347] We asked to remove Rack dc2-rack3 with unschedulable pod  cluster=cassandra-demo
INFO[0347] [cassandra-demo]: Delete PVC[data-cassandra-demo-dc2-rack3-0] OK 
```

The rack3 (and its statefulset) has been removed, and the associated (empty) pvc deleted
