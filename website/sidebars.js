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
                "concepts/1_introduction",
                "concepts/2_design_principes",
            ],
            "Setup": [
                "setup/1_getting_started",
                "setup/2_install_plugin",
                "setup/3_multi_casskop",
                {
                    "type" : "category",
                    "label": "Platform Setup",
                    "items"  : [
                        "setup/platform_setup/1_gke",
                        "setup/platform_setup/2_minikube",
                    ]
                },
                "setup/5_upgrade_v1_to_v2",
            ],
            "Advanced Configuration": [
                "configuration_deployment/1_customizable_install_with_helm",
                "configuration_deployment/2_cassandra_cluster",
                "configuration_deployment/3_storage",
                "configuration_deployment/4_cluster_topology",
                "configuration_deployment/2_cassandra_configuration",
                "configuration_deployment/5_sidecars",
                "configuration_deployment/9_advanced_configuration",
                "configuration_deployment/10_nodes_management",
                "configuration_deployment/11_cassandra_cluster_status",
            ],
            "Operations" : [
                "operations/0_implementation_architecture",
                "operations/1_cluster_operations",
                "operations/2_pods_operations",
                "operations/3_multi_casskop",
                "operations/3_5_backup_restore",
                "operations/4_upgrade_operator",
                "operations/5_uninstall_casskop",
            ],
            "Reference": [
                "references/1_cassandra_cluster",
                "references/2_topology",
                "references/3_cassandra_cluster_status",
                "references/4_multicasskop",
                "references/5_cassandra_backup",
                "references/6_cassandra_restore",
            ],
            "Troubleshooting" : [
                "troubleshooting/1_operations_issues",
                "troubleshooting/2_gke_issues",
            ],
            "Contributing" : [
                "contributing/1_developer_guide",
                "contributing/2_release_guide",
                "contributing/3_reporting_bugs",
                "contributing/4_credits",
            ]
        }
};
