// Copyright 2019 Orange
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// 	You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// 	See the License for the specific language governing permissions and
// limitations under the License.

package cassandracluster

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/Jeffail/gabs"
	"github.com/cscetbon/casskop/controllers/common"
	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	api "github.com/cscetbon/casskop/api/v2"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cscetbon/casskop/pkg/k8s"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func helperInitCluster(t *testing.T, name string) (*CassandraClusterReconciler, *api.CassandraCluster) {
	var cc api.CassandraCluster
	err := yaml.Unmarshal(common.HelperLoadBytes(t, name), &cc)
	if err != nil {
		log.Error(err, "error: helpInitCluster")
		os.Exit(-1)
	}

	ccList := api.CassandraClusterList{}
	//Create Fake client
	//Objects to track in the Fake client
	objs := []runtime.Object{&cc}
	// Register operator types with the runtime scheme.
	fakeClientScheme := scheme.Scheme
	fakeClientScheme.AddKnownTypes(api.GroupVersion, &cc)
	fakeClientScheme.AddKnownTypes(api.GroupVersion, &ccList)
	cl := fake.NewClientBuilder().WithScheme(fakeClientScheme).WithRuntimeObjects(objs...).WithStatusSubresource(&cc).Build()
	// Create a CassandraClusterReconciler object with the scheme and fake client.
	rcc := CassandraClusterReconciler{Client: cl, Scheme: fakeClientScheme}

	cc.InitCassandraRackList()
	return &rcc, &cc
}

func GetEnvVarByName(envVars []v1.EnvVar, name string) *v1.EnvVar {
	for _, envVar := range envVars {
		if envVar.Name == name {
			return &envVar
		}
	}
	return nil
}

func TestCreateNodeAffinity(t *testing.T) {
	assert := assert.New(t)

	nodeAffinity := createNodeAffinity(map[string]string{
		"A": "value1",
		"B": "value2",
		"C": "value3",
		"D": "value4",
		"E": "value5",
	})

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key, "A")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values[0], "value1")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[1].Key, "B")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[1].Values[0], "value2")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[2].Key, "C")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[2].Values[0], "value3")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[3].Key, "D")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[3].Values[0], "value4")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[4].Key, "E")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[4].Values[0], "value5")
}

func TestCreateNodeAffinitySort(t *testing.T) {
	assert := assert.New(t)

	//unsort labels gives sorted result
	nodeAffinity := createNodeAffinity(map[string]string{
		"B": "value2",
		"A": "value1",
		"D": "value4",
		"E": "value5",
		"C": "value3",
	})

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key, "A")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values[0], "value1")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[1].Key, "B")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[1].Values[0], "value2")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[2].Key, "C")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[2].Values[0], "value3")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[3].Key, "D")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[3].Values[0], "value4")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[4].Key, "E")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[4].Values[0], "value5")
}

func TestCreatePodAntiAffinityHard(t *testing.T) {
	assert := assert.New(t)

	labels := map[string]string{
		"label1": "value1",
		"label2": "value2",
		"label3": "value3",
	}
	podAntiAffinityHard := createPodAntiAffinity(true, labels)

	assert.Equal(podAntiAffinityHard.RequiredDuringSchedulingIgnoredDuringExecution[0].TopologyKey, hostnameTopologyKey)
	assert.Equal(podAntiAffinityHard.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels, labels)
}

func TestVolumeMounts(t *testing.T) {
	_, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

	volumeMounts := generateContainerVolumeMount(cc, initContainer)
	assert.Equal(t, 1, len(volumeMounts))
	assert.Equal(t, "/bootstrap", volumeMounts[getPos(volumeMounts, "bootstrap")].MountPath)

	volumeMounts = generateContainerVolumeMount(cc, bootstrapContainer)
	assert.Equal(t, 3, len(volumeMounts))
	assert.Equal(t, "/etc/cassandra", volumeMounts[getPos(volumeMounts, "bootstrap")].MountPath)
	assert.Equal(t, "/extra-lib", volumeMounts[getPos(volumeMounts, "extra-lib")].MountPath)
	assert.Equal(t, "/opt/bin", volumeMounts[getPos(volumeMounts, "tools")].MountPath)

	volumeMounts = generateContainerVolumeMount(cc, cassandraContainer)
	assert.Equal(t, 6, len(volumeMounts))
	assert.Equal(t, "/etc/cassandra", volumeMounts[getPos(volumeMounts, "bootstrap")].MountPath)
	assert.Equal(t, "/extra-lib", volumeMounts[getPos(volumeMounts, "extra-lib")].MountPath)
	assert.Equal(t, "/opt/bin", volumeMounts[getPos(volumeMounts, "tools")].MountPath)
	assert.Equal(t, "/tmp", volumeMounts[getPos(volumeMounts, "tmp")].MountPath)
	assert.Equal(t, "/var/lib/cassandra", volumeMounts[getPos(volumeMounts, "data")].MountPath)
	assert.Equal(t, "/var/log/cassandra", volumeMounts[getPos(volumeMounts, "log")].MountPath)
}

func TestGenerateCassandraService(t *testing.T) {
	assert := assert.New(t)

	_, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")
	selector := k8s.LabelsForCassandra(cc)
	svc := generateCassandraService(cc, selector, nil)

	assert.Equal(map[string]string{
		"app":              "cassandracluster",
		"cassandracluster": "cassandra-demo",
		"cluster":          "k8s.pic"},
		svc.Labels)
	assert.Equal(map[string]string{"external-dns.alpha.kubernetes.io/hostname": "my.custom.domain.com."},
		svc.Annotations)
}

func TestInitContainerConfiguration(t *testing.T) {
	dcName := "dc1"
	rackName := "rack1"
	dcRackName := fmt.Sprintf("%s-%s", dcName, rackName)

	_, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")
	cc.Spec.ServerVersion = "4.0.1"
	cc.Spec.Config, _ = json.Marshal(map[string]map[string]interface{}{
		"jvm-options": {
			"initial_heap_size": "800M",
			"max_heap_size":     "1600M",
		},
	})
	cassieResources := cc.Spec.Resources

	assert := assert.New(t)
	initEnvVar := initContainerEnvVar(cc, &cc.Status, cassieResources, dcRackName)
	bootstrapEnvVar := bootstrapContainerEnvVar(cc, &cc.Status)

	assert.Equal(6, len(bootstrapEnvVar))
	assert.Equal(7, len(initEnvVar))

	configFileData, _ := gabs.ParseJSON([]byte(`{
		"cassandra-yaml": {
			"counter_write_request_timeout_in_ms": 5000,
			"num_tokens":32,
			"read_request_timeout_in_ms": 5000,
			"write_request_timeout_in_ms": 5000
		},
		"cluster-info": {
			"name": "cassandra-demo",
			"seeds": ""
		},
		"datacenter-info": {
			"name": "dc1"
		},
		"jvm-server-options": {
			"cassandra_ring_delay_ms": 30000,
			"initial_heap_size": "800M",
			"jmx-connection-type": "remote-no-auth",
			"max_heap_size": "1600M"
		},
		"logback-xml": {
			"debuglog-enabled": false
		}
	}`))

	vars := map[string]interface{}{
		"CONFIG_FILE_DATA": configFileData.String(),
		"PRODUCT_NAME":     "cassandra",
		"PRODUCT_VERSION":  "4.0.1",
	}

	checkInitContainerVarEnv(t, initEnvVar, vars)

	cc.Spec.Config, _ = json.Marshal(map[string]map[string]interface{}{
		"cassandra-yaml": {
			"read_request_timeout_in_ms": 10000,
		},
		"jvm-options": {
			"cassandra_ring_delay_ms": 10000,
			"initial_heap_size":       "800M",
			"max_heap_size":           "4G",
		},
	})

	cc.Spec.Topology.DC[0].Config, _ = json.Marshal(map[string]map[string]interface{}{
		"jvm-options": {
			"resize_tlb":        "true",
			"initial_heap_size": "1024M",
		},
	})

	cc.Spec.Topology.DC[0].Rack[0].Config, _ = json.Marshal(map[string]map[string]interface{}{
		"jvm-options": {
			"print_tenuring_distribution": "true",
		},
	})

	cc.Spec.ServerVersion = "3.11.19"

	initEnvVar = initContainerEnvVar(cc, &cc.Status, cassieResources, dcRackName)

	assert.Equal(7, len(initEnvVar))

	configFileData.SetP(10000, "cassandra-yaml.read_request_timeout_in_ms")
	configFileData.DeleteP("cassandra-yaml.num_tokens")
	configFileData.DeleteP("jvm-server-options")
	configFileData.SetP(10000, "jvm-options.cassandra_ring_delay_ms")
	configFileData.SetP("1024M", "jvm-options.initial_heap_size")
	configFileData.SetP("4G", "jvm-options.max_heap_size")
	configFileData.SetP("true", "jvm-options.resize_tlb")
	configFileData.SetP("true", "jvm-options.print_tenuring_distribution")
	configFileData.SetP("remote-no-auth", "jvm-options.jmx-connection-type")

	vars = map[string]interface{}{
		"CONFIG_FILE_DATA": configFileData.String(),
		"PRODUCT_NAME":     "cassandra",
		"PRODUCT_VERSION":  cc.Spec.ServerVersion,
	}

	checkInitContainerVarEnv(t, initEnvVar, vars)
}

func TestInitContainerConfigFileData(t *testing.T) {
	_, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")
	cc.Spec.ServerVersion = "3.11.7"
	cassieResources := cc.Spec.Resources
	dcName, rackName := "dc1", "rack1"
	dcRackName := fmt.Sprintf("%s-%s", dcName, rackName)

	initEnvVar := initContainerEnvVar(cc, &cc.Status, cassieResources, dcRackName)

	assert := assert.New(t)

	assert.Equal(7, len(initEnvVar))

	configFileData, _ := gabs.ParseJSON([]byte(`{
		"cassandra-yaml": {
			"counter_write_request_timeout_in_ms": 5000,
			"num_tokens": 32,
			"read_request_timeout_in_ms": 5000,
			"write_request_timeout_in_ms": 5000
		},
		"cluster-info": {
			"name": "cassandra-demo",
			"seeds": ""
		},
		"datacenter-info": {
			"name": "dc1"
		},
		"jvm-options": {
			"cassandra_ring_delay_ms": 30000,
			"initial_heap_size": "128M",
			"jmx-connection-type": "remote-no-auth",
			"max_heap_size": "512M"
		},
		"logback-xml": {
			"debuglog-enabled": false
		}
	}`))

	assert.Equal(configFileData.String(), initEnvVar[0].Value)

	dcRackName = fmt.Sprintf("%s-%s", dcName, "rack2")
	initEnvVar = initContainerEnvVar(cc, &cc.Status, cc.Spec.Resources, dcRackName)
	configFileData.SetP(16, "cassandra-yaml.num_tokens")
	assert.Equal(configFileData.String(), initEnvVar[0].Value)

	dcName = "dc2"
	dcRackName = fmt.Sprintf("%s-%s", dcName, "rack1")
	initEnvVar = initContainerEnvVar(cc, &cc.Status, cc.Spec.Resources, dcRackName)
	configFileData.SetP(64, "cassandra-yaml.num_tokens")
	configFileData.SetP("dc2", "datacenter-info.name")
	assert.Equal(configFileData.String(), initEnvVar[0].Value)
}

func TestInitContainerServerVersionDetect(t *testing.T) {
	// Check of Cassandra version detection in case of different image formats
	dcName := "dc1"
	rackName := "rack1"
	dcRackName := fmt.Sprintf("%s-%s", dcName, rackName)

	_, cc := helperInitCluster(t, "cassandracluster-1DC.yaml")
	imageTag := "1.0.0"
	images := []string{
		fmt.Sprintf("artifactory-name/repo/cassandra:%s", imageTag),
		fmt.Sprintf("artifactory-name:44301/repo/cassandra:%s", imageTag),
	}
	assert := assert.New(t)
	for _, image := range images {
		cc.Spec.CassandraImage = image
		cassieResources := cc.Spec.Resources
		initEnvVar := initContainerEnvVar(cc, &cc.Status, cassieResources, dcRackName)
		prodVer := GetEnvVarByName(initEnvVar, "PRODUCT_VERSION")
		assert.Equal(imageTag, prodVer.Value)
	}
}

func TestGenerateCassandraStatefulSet(t *testing.T) {
	assert := assert.New(t)
	dcName := "dc1"
	rackName := "rack1"
	dcRackName := fmt.Sprintf("%s-%s", dcName, rackName)

	_, cc := HelperInitCluster(t, "cassandracluster-2DC.yaml")
	ccDefault := cc.DeepCopy()
	cc.CheckDefaults()
	labels, nodeSelector := k8s.DCRackLabelsAndNodeSelectorForStatefulSet(cc, 0, 0)
	sts, _ := generateCassandraStatefulSet(cc, &cc.Status, dcName, dcRackName, labels, nodeSelector, nil)

	assert.Equal(map[string]string{
		"app":                                  "cassandracluster",
		"cassandracluster":                     "cassandra-demo",
		"cassandraclusters.db.orange.com.dc":   "dc1",
		"cassandraclusters.db.orange.com.rack": "rack1",
		"dc-rack":                              "dc1-rack1",
		"cluster":                              "k8s.pic",
	}, sts.Labels)

	assert.Equal("my.custom.annotation", sts.Spec.Template.Annotations["exemple.com/test"])
	assert.Equal([]v1.Toleration{
		{
			Key:      "my_custom_taint",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
	}, sts.Spec.Template.Spec.Tolerations)

	assert.Equal(int64(1001), *sts.Spec.Template.Spec.SecurityContext.RunAsUser)
	assert.Equal(int64(1002), *sts.Spec.Template.Spec.SecurityContext.FSGroup)

	checkVolumeClaimTemplates(t, labels, sts.Spec.VolumeClaimTemplates, "10Gi", "test-storage")
	checkLiveAndReadiNessProbe(t, sts.Spec.Template.Spec.Containers,
		1010, 201, 32, 7, 9, 1205, 151, 17, 50, 30)
	checkVolumeMount(t, sts.Spec.Template.Spec.Containers)
	checkVarEnv(t, sts.Spec.Template.Spec.Containers, cc, dcRackName)
	checkDefaultInitContainerResources(t, sts.Spec.Template.Spec.InitContainers)
	checkBackRestSidecar(t, sts.Spec.Template.Spec.Containers,
		"ghcr.io/cscetbon/instaclustr-icarus:1.1.3",
		v1.PullAlways,
		v1.ResourceRequirements{
			Requests: generateResourceList("1", "1Gi"),
			Limits:   generateResourceList("2", "3Gi"),
		})
	checkResourcesConfiguration(t, sts.Spec.Template.Spec.Containers, "3", "3Gi")

	cc.Spec.StorageConfigs[0].PVCSpec = nil
	_, err := generateCassandraStatefulSet(cc, &cc.Status, dcName, dcRackName, labels, nodeSelector, nil)
	assert.NotEqual(t, err, nil)

	// Test default setup
	dcNameDefault := "dc2"
	rackNameDefault := "rack1"
	dcRackNameDefault := fmt.Sprintf("%s-%s", dcNameDefault, rackNameDefault)
	setupForDefaultTest(ccDefault)

	ccDefault.CheckDefaults()
	labelsDefault, nodeSelectorDefault := k8s.DCRackLabelsAndNodeSelectorForStatefulSet(ccDefault, 0, 0)
	stsDefault, _ := generateCassandraStatefulSet(ccDefault, &ccDefault.Status, dcNameDefault, dcRackNameDefault,
		labelsDefault, nodeSelectorDefault, nil)

	checkVolumeClaimTemplates(t, labels, stsDefault.Spec.VolumeClaimTemplates, "3Gi", "local-storage")
	checkLiveAndReadiNessProbe(t, stsDefault.Spec.Template.Spec.Containers,
		60, 10, 10, 0, 0, 120, 20, 10, 0, 0)
	checkDefaultInitContainerResources(t, stsDefault.Spec.Template.Spec.InitContainers)
	resources := generateResourceList(defaultBackRestContainerRequestsCPU, defaultBackRestContainerRequestsMemory)
	checkBackRestSidecar(t, stsDefault.Spec.Template.Spec.Containers,
		api.DefaultBackRestImage,
		"",
		v1.ResourceRequirements{
			Requests: resources,
			Limits:   resources,
		})
	checkResourcesConfiguration(t, stsDefault.Spec.Template.Spec.Containers, "1", "2Gi")

	ccDefault.Spec.BackRestSidecar.Image = ""
	ccDefault.CheckDefaults()
	assert.Equal(ccDefault.Spec.BackRestSidecar.Image, api.DefaultBackRestImage)
}

func TestGenerateCassandraStatefulSetBackrestVolumes(t *testing.T) {
	assert := assert.New(t)
	dcName := "dc1"
	rackName := "rack1"
	dcRackName := fmt.Sprintf("%s-%s", dcName, rackName)

	_, cc := HelperInitCluster(t, "cassandracluster-backrest-vol.yaml")
	cc.CheckDefaults()
	extraVol := v1.VolumeMount{Name: "data", MountPath: "/data/backups"}
	labels, nodeSelector := k8s.DCRackLabelsAndNodeSelectorForStatefulSet(cc, 0, 0)
	sts, _ := generateCassandraStatefulSet(cc, &cc.Status, dcName, dcRackName, labels, nodeSelector, nil)
	for _, c := range sts.Spec.Template.Spec.Containers {
		if c.Name == "backrest-sidecar" {
			assert.Equal(5, len(c.VolumeMounts))
			assert.Contains(c.VolumeMounts, extraVol)
		}
	}

}

func TestCassandraStatefulSetHasNoDuplicateVolumes(t *testing.T) {
	dcName := "dc1"
	dcRackName := fmt.Sprintf("%s-rack1", dcName)

	_, cc := HelperInitCluster(t, "cassandracluster-2DC.yaml")
	labels, nodeSelector := k8s.DCRackLabelsAndNodeSelectorForStatefulSet(cc, 0, 0)
	sts, _ := generateCassandraStatefulSet(cc, &cc.Status, dcName, dcRackName, labels, nodeSelector, nil)

	assert := assert.New(t)
	cassandraContainer := sts.Spec.Template.Spec.Containers[2]
	assert.Equal(cassandraContainer.Name, cassandraContainerName)
	cassandraLogVolumeMounts := 0
	for _, vol := range cassandraContainer.VolumeMounts {
		if vol.MountPath == "/var/log/cassandra" {
			cassandraLogVolumeMounts++
		}
	}
	assert.Equal(cassandraLogVolumeMounts, 1, "Duplicate volume mount found in Cassandra container")
	assert.Equal(len(sts.Spec.Template.Spec.Volumes), 4, "Volume defined when it is a VolumeClaim")
}

func checkResourcesConfiguration(t *testing.T, containers []v1.Container, cpu string, memory string) {
	for _, c := range containers {
		if c.Name == "cassandra" {
			assert.Equal(t, resource.MustParse(cpu), *c.Resources.Requests.Cpu())
			assert.Equal(t, resource.MustParse(memory), *c.Resources.Requests.Memory())
			assert.Equal(t, resource.MustParse(cpu), *c.Resources.Limits.Cpu())
			assert.Equal(t, resource.MustParse(memory), *c.Resources.Limits.Memory())
		}
	}
}

func setupForDefaultTest(cc *api.CassandraCluster) {
	cc.Spec.LivenessFailureThreshold = nil
	cc.Spec.LivenessSuccessThreshold = nil
	cc.Spec.LivenessHealthCheckPeriod = nil
	cc.Spec.LivenessHealthCheckTimeout = nil
	cc.Spec.LivenessInitialDelaySeconds = nil
	cc.Spec.ReadinessHealthCheckPeriod = nil
	cc.Spec.ReadinessHealthCheckTimeout = nil
	cc.Spec.ReadinessInitialDelaySeconds = nil
	cc.Spec.ReadinessFailureThreshold = nil
	cc.Spec.ReadinessSuccessThreshold = nil
	cc.Spec.BackRestSidecar = nil
}

func checkBackRestSidecar(t *testing.T, containers []v1.Container,
	image string,
	imagePullPolicy v1.PullPolicy,
	resources v1.ResourceRequirements,
) {
	for _, c := range containers {
		if c.Name == "backrest-sidecar" {
			assert.Equal(t, image, c.Image)
			assert.Equal(t, imagePullPolicy, c.ImagePullPolicy)
			assert.Equal(t, resources, c.Resources)
		}
	}
}

func checkLiveAndReadiNessProbe(t *testing.T, containers []v1.Container,
	readinessInitialDelaySecond,
	readinessTimeoutSeconds,
	readinessPeriodSeconds,
	readinessFailureThreshold,
	readinessSuccessThreshold,
	livenessInitialDelaySecond,
	livenessTimeoutSeconds,
	livenessPeriodSeconds,
	livenessFailureThreshold,
	livenessSuccessThreshold int32) {
	for _, c := range containers {
		if c.Name == cassandraContainerName {
			// Readiness Config check
			assert.Equal(t, readinessInitialDelaySecond, c.ReadinessProbe.InitialDelaySeconds)
			assert.Equal(t, readinessTimeoutSeconds, c.ReadinessProbe.TimeoutSeconds)
			assert.Equal(t, readinessPeriodSeconds, c.ReadinessProbe.PeriodSeconds)
			assert.Equal(t, readinessFailureThreshold, c.ReadinessProbe.FailureThreshold)
			assert.Equal(t, readinessSuccessThreshold, c.ReadinessProbe.SuccessThreshold)

			// Liveness Config check
			assert.Equal(t, livenessInitialDelaySecond, c.LivenessProbe.InitialDelaySeconds)
			assert.Equal(t, livenessTimeoutSeconds, c.LivenessProbe.TimeoutSeconds)
			assert.Equal(t, livenessPeriodSeconds, c.LivenessProbe.PeriodSeconds)
			assert.Equal(t, livenessFailureThreshold, c.LivenessProbe.FailureThreshold)
			assert.Equal(t, livenessSuccessThreshold, c.LivenessProbe.SuccessThreshold)
		}
	}
}

func checkVolumeClaimTemplates(t *testing.T, expectedlabels map[string]string, pvcs []v1.PersistentVolumeClaim,
	dataCapacity, dataClassStorage string) {
	assert.Equal(t, 2, len(pvcs))
	for _, pvc := range pvcs {
		switch pvc.Name {
		case "data":
			assert.Equal(t, generateExpectedDataStoragePVC(expectedlabels, dataCapacity, dataClassStorage), pvc)
		case "cassandra-logs":
			assert.Equal(t, generateExpectedCassandraLogsStoragePVC(expectedlabels), pvc)
		default:
			t.Errorf("unexpected pvc name: %s.", pvc.Name)
		}
	}
}

func generateExpectedDataStoragePVC(expectedlabels map[string]string, dataCapacity, dataClassStorage string) v1.PersistentVolumeClaim {

	expectedDataStorageQuantity, _ := resource.ParseQuantity(dataCapacity)

	return v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "data",
			Labels: expectedlabels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},

			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					"storage": expectedDataStorageQuantity,
				},
			},
			StorageClassName: &dataClassStorage,
		},
	}
}

func generateExpectedCassandraLogsStoragePVC(expectedlabels map[string]string) v1.PersistentVolumeClaim {

	expectedDataStorageQuantity, _ := resource.ParseQuantity("10Gi")
	expectedDataStorageClassName := "standard-wait"

	return v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "cassandra-logs",
			Labels: expectedlabels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},

			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					"storage": expectedDataStorageQuantity,
				},
			},
			StorageClassName: &expectedDataStorageClassName,
		},
	}
}

func checkVolumeMount(t *testing.T, containers []v1.Container) {
	assert.Equal(t, 4, len(containers))
	for _, container := range containers {
		switch container.Name {
		case "cassandra":
			assert.Equal(t, 6, len(container.VolumeMounts))
		case "gc-logs":
			assert.Equal(t, 1, len(container.VolumeMounts))
		case "cassandra-logs":
			assert.Equal(t, 1, len(container.VolumeMounts))
		case "backrest-sidecar":
			assert.Equal(t, 4, len(container.VolumeMounts))
		default:
			t.Errorf("unexpected container: %s.", container.Name)
		}

		_, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

		for _, volumeMount := range container.VolumeMounts {
			switch container.Name {
			case "cassandra":
				assert.True(t, volumesContains(append(generateContainerVolumeMount(cc, cassandraContainer),
					generateCassandraStorageConfigVolumeMounts()...), volumeMount))
			case "gc-logs", "cassandra-logs":
				assert.True(t, volumesContains([]v1.VolumeMount{{Name: "cassandra-logs", MountPath: "/var/log/cassandra"}}, volumeMount))
			case "backrest-sidecar":
				assert.True(t, volumesContains(generateContainerVolumeMount(cc, backrestContainer), volumeMount))
			default:
				t.Errorf("unexpected container: %s.", container.Name)
			}
		}
	}
}

func checkDefaultInitContainerResources(t *testing.T, containers []v1.Container) {

	resourcesRequirements := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			"cpu":    resource.MustParse(defaultInitContainerLimitsCPU),
			"memory": resource.MustParse(defaultInitContainerLimitsMemory),
		},
		Requests: v1.ResourceList{
			"cpu":    resource.MustParse(defaultInitContainerRequestsCPU),
			"memory": resource.MustParse(defaultInitContainerRequestsMemory),
		},
	}

	for _, container := range containers {
		switch container.Name {
		case "bootstrap":
			assert.Equal(t, container.Resources, resourcesRequirements)
		case "init-config":
			assert.Equal(t, container.Resources, resourcesRequirements)
		default:
		}
	}
}

func volumesContains(vms []v1.VolumeMount, mount v1.VolumeMount) bool {
	for _, vm := range vms {
		if mount == vm {
			return true
		}
	}
	return false
}

func generateCassandraStorageConfigVolumeMounts() []v1.VolumeMount {
	var vms []v1.VolumeMount
	vms = append(vms, v1.VolumeMount{Name: "cassandra-logs", MountPath: "/var/log/cassandra"})

	return vms
}

func checkVarEnv(t *testing.T, containers []v1.Container, cc *api.CassandraCluster, dcRackName string) {
	cassieResources := cc.Spec.Resources
	initContainerEnvVar := initContainerEnvVar(cc, &cc.Status, cassieResources, dcRackName)
	bootstrapContainerEnvVar := bootstrapContainerEnvVar(cc, &cc.Status)
	jmxEnvVar := generateJMXConfiguration(*cc.Spec.JMXConfiguration)

	assert := assert.New(t)

	assert.Equal(6, len(bootstrapContainerEnvVar))
	assert.Equal(4, len(containers))
	assert.Equal(7, len(initContainerEnvVar))

	configFileData, _ := gabs.ParseJSON([]byte(`{
		"cassandra-yaml": {
			"counter_write_request_timeout_in_ms":5000,
			"num_tokens":32,
			"read_request_timeout_in_ms":5000,
			"write_request_timeout_in_ms":5000
		},
		"cluster-info": {
			"name":  "cassandra-demo",
			"seeds": ""
		},
		"datacenter-info": {
			"name": "dc1"
		},
		"jvm-options": {
			"cassandra_ring_delay_ms":30000,
			"initial_heap_size":"128M",
			"jmx-connection-type":"remote-no-auth",
			"max_heap_size":"512M"
		},
		"logback-xml": {
			"debuglog-enabled": false
		}
	}`))

	vars := map[string]interface{}{
		"CONFIG_FILE_DATA":       configFileData.String(),
		"PRODUCT_NAME":           "cassandra",
		"PRODUCT_VERSION":        "3.11.7",
		"CASSANDRA_SEEDS":        "",
		"CASSANDRA_DC":           "",
		"CASSANDRA_RACK":         "",
		"CASSANDRA_LOG_DIR":      "/var/log/cassandra",
		"CASSANDRA_CLUSTER_NAME": "cassandra-demo",
	}

	for _, container := range containers {
		if container.Name != cassandraContainerName {
			for _, env := range container.Env {
				if env.Name == "POD_IP" {
					continue
				}
				assert.Contains(vars, env.Name)
				assert.Equal(vars[env.Name], env.Value)
			}
		} else {
			// Check cassandra container env vars
			podIP := v1.EnvVar{
				Name: "POD_IP",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "status.podIP",
					},
				},
			}
			assert.Contains(container.Env, podIP)
			assert.Contains(container.Env, jmxEnvVar)

			checkInitContainerVarEnv(t, initContainerEnvVar, vars)
		}
	}
}

func checkInitContainerVarEnv(t *testing.T, initContainerEnvVar []v1.EnvVar, vars map[string]interface{}) {
	assert := assert.New(t)
	for _, env := range initContainerEnvVar {
		if value, ok := vars[env.Name]; ok {
			assert.Equal(value, env.Value)
		}
	}
}

func TestDisableBackRest(t *testing.T) {
	// Check of Cassandra version detection in case of different image formats
	dcName := "dc1"
	rackName := "rack1"
	dcRackName := fmt.Sprintf("%s-%s", dcName, rackName)
	_, cc := helperInitCluster(t, "cassandracluster-disable-backrest.yaml")
	cc.CheckDefaults()
	labels, nodeSelector := k8s.DCRackLabelsAndNodeSelectorForStatefulSet(cc, 0, 0)
	sts, _ := generateCassandraStatefulSet(cc, &cc.Status, dcName, dcRackName, labels, nodeSelector, nil)
	assert := assert.New(t)
	assert.Equal(1, len(sts.Spec.Template.Spec.Containers))
}
