/**
 * Copyright (c) 2017-present, Facebook, Inc.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

module.exports = {
    "docs":
        {
            "Concepts": [
                "concepts/introduction",
                "concepts/design_principes",
            ],
            "Setup": [
                "setup/getting_started",
                "setup/install_plugin",
                "setup/multi_casskop",
                {
                    "type" : "category",
                    "label": "Platform Setup",
                    "items"  : [
                        "setup/platform_setup/gke",
                        "setup/platform_setup/minikube",
                    ]
                },
                "setup/upgrade_v1_to_v2",
            ],
            "Advanced Configuration": [
                "configuration_deployment/customizable_install_with_helm",
                "configuration_deployment/cassandra_cluster",
                "configuration_deployment/storage",
                "configuration_deployment/cluster_topology",
                "configuration_deployment/cassandra_configuration",
                "configuration_deployment/sidecars",
                "configuration_deployment/advanced_configuration",
                "configuration_deployment/nodes_management",
                "configuration_deployment/cassandra_cluster_status",
            ],
            "Operations" : [
                "operations/implementation_architecture",
                "operations/cluster_operations",
                "operations/pods_operations",
                "operations/multi_casskop",
                "operations/backup_restore",
                "operations/upgrade_operator",
                "operations/uninstall_casskop",
            ],
            "Reference": [
                "references/cassandra_cluster",
                "references/topology",
                "references/cassandra_cluster_status",
                "references/multicasskop",
                "references/cassandra_backup",
                "references/cassandra_restore",
            ],
            "Troubleshooting" : [
                "troubleshooting/operations_issues",
                "troubleshooting/gke_issues",
            ],
            "Contributing" : [
                "contributing/developer_guide",
                "contributing/release_guide",
                "contributing/reporting_bugs",
                "contributing/credits",
            ]
        }
};
