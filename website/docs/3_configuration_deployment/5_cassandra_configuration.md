---
title: Cassandra Configuration
sidebar_label: Cassandra Configuration
---

## Cassandra bootstrap

The configuration of the Cassandra image has been delegated to two init-containers that are executed prior to the Cassandra
container when the pod is started.

There is a specific bootstrap image that is build from the docker directory and contains all required files or scripts
to work with CassKop.

### Initcontainer 1 : base-config-builder

The init container is responsible for the following actions :

- copy the default Cassandra configuration from user's provided Cassandra image.

Configuration:

- The init-config container is by default the baseImage Cassandra image
- The default command executed by the init-container is:

```bash
cp -vr /etc/cassandra/* /bootstrap
```

### Initcontainer 2 : config-builder

This container is responsible for the following actions :

- Generate all the cassandra configuration (cassandra.yaml, java. etc..) from the different config entries in the
  CassandraCluster object level (spec, datacenters and racks)

It uses [cass-config-builder](https://github.com/datastax/cass-config-builder) which is a tool developed by Datastax.
CassKop ships its own build of that image, `ghcr.io/cscetbon/cass-config-builder`, that adds the definitions of the
Cassandra versions no longer published upstream (4.1 and 5.0), see
[config-builder](https://github.com/cscetbon/casskop/tree/master/docker/config-builder). It can be changed using
`Spec.configBuilderImage`.

### Initcontainer 3 : bootstrap

The bootstrap Container :

- Adding files and additional jar from the bootstrap image to the default configuration
- Applying a configmap on top of the default configuration
- Modifying the configuration to be suitable to run with CassKop:
  * Updating the cluster name
  * Setting the seedlist
  * Adding a Cassandra exporter and Jolokia agent
  * ..

We provide the bootstrap image, but you can change it using `Spec.bootstrapImage` but you need to comply with the
required actions, see [Bootstrap](https://github.com/cscetbon/casskop/tree/master/docker/bootstrap).

## Nodes per rack

One of the requirements for CassKop is to always keep the same number of nodes in each of its racks per Cassandra DCs. The number of nodes used for the Cassandra Cluster is configured using the `CassandraCluster.spec.nodesPerRacks`
property.

> If you have not specify a Cluster Topology, then you'll have a default datacenter named `dc1` and a default rack named
> `rack1`

CassKop will keep the same number of nodes in each Cassandra rack as it is a good practice for Cassandra.

You can define different numbers of replicas for racks in different Cassandra DataCenters using the `nodesPerRacks` property at the datacenter level. (in `CassandraCluster.spec.topology.dc[<idx>].nodesPerRacks`). If
specified on the datacenter level, this parameter takes priority over the global `CassandraCluster.spec.nodesPerRacks`.

Example:
example to scale up the nodesPerRacks in DC2 :

```yaml
  topology:
    dc:
      - name: dc1
        rack:
          - name: rack1
          - name: rack2
      - name: dc2
        nodesPerRacks: 3        <--- We increase by one this value
        rack:
          - name: rack1
```

> The number of Cassandra nodes will be the multiplication of the number of racks * the nodesPerRacks value.

If we changes on of these properties then CassKop will trigger either a [ScaleUp](/casskop/docs/operations/cluster_operations#scaleup)
or a [ScaleDown](/casskop/docs/operations/cluster_operations#scaledown) operation.


## Configuration embedded in CassandraCluster

The configuration of the different components of Cassandra can be changed in the object and be overriden at different
levels. Here is an example showing how to update some Java and Cassdnra parameters as well and overwrite them at the
datacenter and rack level:

```yaml
apiVersion: db.orange.com/v2
kind: CassandraCluster
metadata:
  name: cassandra-demo
spec:
  nodesPerRacks: 1
  cassandraImage: cassandra:3.11.7
  restartCountBeforePodDeletion: 3
  dataStorageClass: local-storage
  hardAntiAffinity: false
  deletePVC: true
  autoPilot: true
  resources:
    limits:
      cpu: 1
      memory: 2Gi
  config:
    cassandra-yaml:
      num_tokens: 64
  topology:
    dc:
      - name: dc1
        config:
          cassandra-yaml:
            num_tokens: 32
        resources:
          limits:
            cpu: 3
            memory: 3Gi
        labels:
          location.dfy.orange.com/site : mts
        rack:
          - name: rack1
            labels:
              location.dfy.orange.com/street : street1
          - name: rack2
            labels:
              location.dfy.orange.com/street : street2
            config:
              cassandra-yaml:
                num_tokens: 16
      - name: dc2
        nodesPerRacks: 1
        labels:
          location.dfy.orange.com/site : mts
        rack:
          - name: rack1
            labels:
              location.dfy.orange.com/street : street3
```

## Cassandra versions

CassKop is tested with the official `cassandra` images of the 3.11, 4.0, 4.1 and 5.0 lines. The configuration is
generated for the version found in the tag of `Spec.cassandraImage` (`cassandra:4.1.5` is 4.1.5, `cassandra:5.0`
is 5.0.0) unless `Spec.serverVersion` is set, which is only needed when the tag does not carry the version.

### Cassandra 4.1 and later

Cassandra 4.1 renamed most of the `cassandra.yaml` settings so that their value carries a unit
(`read_request_timeout_in_ms: 5000` became `read_request_timeout: 5000ms`). From 4.1 the configuration is generated
with the new names and the legacy names found in `config` sections are translated, so a configuration written for
4.0 keeps working. Both styles can be used, the new one wins when a setting is given twice.

### Cassandra 5.0

Cassandra 5.0 nodes are started with `storage_compatibility_mode: CASSANDRA_4`, the mode required to upgrade a
cluster from 4.x. Once every node runs 5.0 and `nodetool upgradesstables` has been run, the mode can be moved to
`UPGRADING` and then `NONE` through `config.cassandra-yaml.storage_compatibility_mode` to unlock the new storage
features, each change being applied by a rolling restart.

### Prometheus metrics

The Prometheus exporter embedded in the Cassandra pods, [cassandra-exporter](https://github.com/instaclustr/cassandra-exporter),
is compiled against Cassandra internals and only has releases compatible with 3.x (0.9.8) and 4.1 (0.9.14). On 4.0 and
5.0 nodes the agent is not loaded and port 9500 does not serve metrics
([cassandra-exporter#126](https://github.com/instaclustr/cassandra-exporter/issues/126),
[cassandra-exporter#127](https://github.com/instaclustr/cassandra-exporter/issues/127)). Jolokia, used by CassKop and
by the probes, is not affected.

### Upgrading a cluster

Changing `Spec.cassandraImage` triggers a rolling update of the pods, one rack at a time. Cassandra supports
upgrading from 3.11 to 4.x and from 4.x to 5.0; the versions must be upgraded one major at a time and
`nodetool upgradesstables` should be run once every node runs the new version.

## Configuration override using configMap

CassKop allows you to customize the configuration of Apache Cassandra nodes by specifying a dedicated `ConfigMap`
name in the `CassandraCluster.spec.configMapName` containing configuration files to be overwritten above the default
configuration files of the dedicated Docker Image.

We have a specific Cassandra Docker image startup script that will overwrite each file in the directory
`/etc/cassandra`, from the one specified in the configMap if they exist.

>You can surcharge any files in the docker image `/etc/cassandra` with the ConfigMap.

Typical overwriting files may be :

- specifying a pre_run.sh and/or post_run.sh script

See the example below:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cassandra-configmap-v1
data:

  pre_run.sh: |-
    echo "  ** this is pre_run.sh script executed before run.sh **"
    #Examples:
    echo "Change default Authenticator & Authorizer"
    sed -ri 's/(authenticator:).*/\1 PasswordAuthenticator/' /etc/cassandra/cassandra.yaml
    sed -ri 's/(authorizer:).*/\1 CassandraAuthorizer/' /etc/cassandra/cassandra.yaml
    #test "$(hostname)" == 'cassandra-demo-dc1-rack2-0' && echo "update param" && sed -i 's/windows_timer_interval: 1/windows_timer_interval: 2/' /etc/cassandra/cassandra.yaml
    #test "$(hostname)" == 'cassandra-demo-dc1-rack3-0' && echo "-Dcassandra.replace_address_first_boot=172.31.183.209" > /etc/cassandra/jvm.options
    #test "$(hostname)" == 'cassandra-demo-dc2-rack1-0' && echo "-Dcassandra.override_decommission=true" > /etc/cassandra/jvm.options
    echo "  ** end of pre_run.sh script, continue with run.sh **"
  post_run.sh: |-
    echo "Check Configured seeds by bootstrap"
    grep "seeds:" /etc/cassandra/cassandra.yaml
```

You can create the ConfigMap above from the repo:

```bash
kubectl apply -f config/cassandra-configmap-v1.yaml
configmap "cassandra-configmap-v1" created
```

Now you can add the `configMapName: cassandra-configmap-v1` to the Spec section of your CassandraCluster definition
[example](https://github.com/cscetbon/casskop/tree/master/config/samples/cassandracluster.yaml)

If you edit the ConfigMap it won't be detected neither by CassKop nor by the statefulsets/pods (unless you reboot the
pods).
It is recommended for configuration changes, to version the configmap and to create apply new configmap in the CRD, this
will trigger a rollingRestart of the whole cluster applying the new configuration.

:::important
each time you specify a new configMap CassKop will start a `rollingUpdate` of all nodes
in the cluster. more info on [UpdateConfigMap](/casskop/docs/operations/cluster_operations#updateconfigmap)
:::

:::important
At this time CassKop won't allow you to specify only excerpt of the configurations files, your
ConfigMap **MUST** contain valid and complete configuration files
:::

### Configuration pre_run.sh script

In case you need to make some specific actions on a particular node, such as make use of the **CASSANDRA_REPLACE_NODE**
variable, you can use the pre_run.sh script in the ConfigMap. If present, the cassandra docker will execute this script
prior to the `run.sh` script from the docker image.

Example of a configMap with the pre_run.sh script :

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cassandra-configmap-pre-run
data:
  pre_run.sh: |-
    echo "** this is a pre-scrip for run.sh that can be edit with configmap"
    test "$(hostname)" == 'cassandra-demo-dc1-rack1-0' && export CASSANDRA_REPLACE_NODE=10.233.93.174
    echo "** end of pre_run.sh script, continue with run.sh"
```

:::important
In the case you use the configmap for one-time specific action, don't forget to edit again to remove
the specific treatment once it is no more needed.
:::

## JVM options

### Memory

Apache Cassandra is running inside a Java Virtual Machine (JVM). JVM has many configuration options to optimize the
performance for different platforms and architectures.

CassKop allows configuring these values by adding a `jvm.options` in the user `ConfigMap`.

The default value used for `-Xmx` depends on whether there is a memory request configured for the container :

- If there is a memory request, the JVM's maximum memory must be set to a value corresponding to the limit.
- When there is no memory request, CassKop will limit it to "2048M".

- set the memory request and the memory limit to the same value, so that the pod is in guarantee mode

> CassKop will automatically compute the initial and max heap size from 1/4 of the available defined resources.

## Authentication and authorizations

CassKop uses Jolokia from the cassandra-image to communicate. We can add
authentication on Jolokia by defining a secret :

Example:

```console
apiVersion: v1
kind: Secret
metadata:
  name: jolokia-auth
type: Opaque
data:
  password: TTBucDQ1NXcwcmQ=
  username: am9sb2tpYS11c2Vy
```

and in the CRD you will define `spec.imageJolokiaSecret`

```console
...
  imageJolokiaSecret:
    name: jolokia-auth
...
```

CassKop will propagate the secrets in Cassandra so that it can configure Jolokia and use it to connect.
