<img src="static/casskop.png" alt="Logo" width="150"/>

# CassKop - Cassandra Kubernetes operator

## Overview

CassKop, the Cassandra Kubernetes operator makes it easy to run Apache Cassandra on Kubernetes. Apache Cassandra is a popular, free, open-source, distributed wide column store, NoSQL database management system.
The operator allows to easily create and manage racks and data centers aware Cassandra clusters.

CassKop is based on CoreOS
[operator-sdk](https://github.com/operator-framework/operator-sdk) tools and APIs.


CassKop creates/configures/manages Cassandra clusters atop Kubernetes and is by default **space-scoped** which means
that :
- CassKop is able to manage X Cassandra clusters in one Kubernetes namespace.
- You need X instances of CassKop to manage Y Cassandra clusters in X different namespaces (1 instance of CassKop
  per namespace).

> This adds security between namespaces with a better isolation, and less work for each operator.

## Installation

For detailed installation instructions, see the [Casskop Documentation Page](https://orange-opensource.github.io/casskop/docs/2_setup/1_getting_started).

## Documentation

The documentation of the Casskop operator project is available at the [Casskop Documentation Page](https://orange-opensource.github.io/casskop/docs/1_concepts/1_introduction).

## Cassandra operator

The Cassandra operator image is automatically built and stored on [Docker Hub](https://cloud.docker.com/u/orangeopensource/repository/docker/orangeopensource/casskop)

Casskop uses standard Cassandra image (tested up to Version 3.11)

### Operator SDK

CassKop is build using operator SDK:

- [operator-sdk](https://github.com/operator-framework/operator-sdk)
- [operator-lifecycle-manager](https://github.com/operator-framework/operator-lifecycle-manager)

### Build pipelines

We uses Github Action as our CI tool to build and test the operator.

#### Build image

To accelerate build phases we have created a custom [build-image](docker/ci/Dockerfile) used by the CI pipeline:

https://cloud.docker.com/u/orangeopensource/repository/docker/orangeopensource/casskop-build

You can find more info in the [developer Section](documentation/development.md)

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches and the contribution workflow.

### For developers

Operator SDK is part of the operator framework provided by RedHat & CoreOS. The goal 
is to provide high-level abstractions that simplifies creating Kubernetes operators.

The quick start guide walks through the process of building the Cassandra operator 
using the SDK CLI, setting up the RBAC, deploying the operator and creating a 
Cassandra cluster.

You can find this in the [Developer section](/casskop/docs/8_contributing/1_developer_guide)

# Contacts

You can contact the team with our mailing-list prj.casskop.support@list.orangeportails.net and join our slack https://casskop.slack.com (request sent to that ML)

# License

CassKop is under Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
