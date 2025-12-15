package envVars

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs"
	api "github.com/cscetbon/casskop/api/v2"
	v1 "k8s.io/api/core/v1"
)

type NodeConfig map[string]map[string]interface{}

/*JvmMemory sets the maximium size of the heap*/
type JvmMemory struct {
	maxHeapSize     string
	initialHeapSize string
}

// JMXConfigurationMap
// Create a JMX Configuration map to convert values from CR to how they look like as env vars
var JMXConfigurationMap = map[string]string{
	"JMXRemote":             "-Dcom.sun.management.jmxremote=",
	"JMXRemotePort":         "-Dcom.sun.management.jmxremote.port=",
	"JMXRemoteRmiPort":      "-Dcom.sun.management.jmxremote.rmi.port=",
	"JXMRemoteSSL":          "-Dcom.sun.management.jmxremote.ssl=",
	"JMXRemoteAuthenticate": "-Dcom.sun.management.jmxremote.authenticate=",
}

const (
	cassandraClusterNameEnvVarName = "CASSANDRA_CLUSTER_NAME"
	cassandraSeedsEnvVarName       = "CASSANDRA_SEEDS"
	cassandraDcEnvVarName          = "CASSANDRA_DC"
	cassandraRackEnvVarName        = "CASSANDRA_RACK"

	podIpEnvVarName           = "POD_IP"
	cassandraLogDirEnvVarName = "CASSANDRA_LOG_DIR"

	jolokiaUserEnvVarName          = "JOLOKIA_USER"
	jolokiaPasswordEnvVarName      = "JOLOKIA_PASSWORD"
	cassandraAuthJolokiaEnvVarName = "CASSANDRA_AUTH_JOLOKIA"

	javaToolOptionsEnvVarName = "JAVA_TOOL_OPTIONS"

	jvmOptsEnvVarName = "JVM_OPTS"

	defaultJvmMaxHeap  = "2048M"
	defaultJvmInitHeap = "512M"
)

func BootstrapContainerEnvVar(cc *api.CassandraCluster, status *api.CassandraClusterStatus) []v1.EnvVar {
	baseBootstrapEnvVars := baseCassandraEnvVars(cc, status)
	commonBootstrapEnvVars := commonEnvVars(cc)
	customEnvVars := cc.Spec.BackRestSidecar.EnvVars
	return deduplicateMerge(baseBootstrapEnvVars, commonBootstrapEnvVars, customEnvVars)
}

// deduplicateMerge merges multiple slices of EnvVars, removing duplicates. When resolving duplicates the first argument has the lowest priority, the last argument has the highest priority.
func deduplicateMerge(envVarSlices ...[]v1.EnvVar) []v1.EnvVar {
	result := make(map[string]v1.EnvVar)

	for _, slice := range envVarSlices {
		envVarMap := toMap(slice)
		for envName, env := range envVarMap {
			result[envName] = env
		}
	}

	envVars := slices.Collect(maps.Values(result))
	// sort to ensure stable output and avoid unnecessary restarts
	slices.SortFunc(envVars, func(a, b v1.EnvVar) int {
		return strings.Compare(a.Name, b.Name)
	})
	return envVars
}

func toMap(slice []v1.EnvVar) map[string]v1.EnvVar {
	result := make(map[string]v1.EnvVar)
	for _, env := range slice {
		result[env.Name] = env
	}
	return result
}

func baseCassandraEnvVars(cc *api.CassandraCluster, status *api.CassandraClusterStatus) []v1.EnvVar {
	return []v1.EnvVar{
		{
			Name:  cassandraClusterNameEnvVarName,
			Value: cc.GetName(),
		},
		{
			Name:  cassandraSeedsEnvVarName,
			Value: cc.SeedList(&status.SeedList),
		},
		{
			Name: cassandraDcEnvVarName,
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.labels['cassandraclusters.db.orange.com.dc']",
				},
			},
		},
		{
			Name: cassandraRackEnvVarName,
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.labels['cassandraclusters.db.orange.com.rack']",
				},
			},
		},
	}
}

func CassandraEnvVars(cc *api.CassandraCluster) []v1.EnvVar {
	commonCassandraEnvVars := commonEnvVars(cc)
	// This option required for nodetool correct execution
	commonCassandraEnvVars = append(commonCassandraEnvVars, v1.EnvVar{
		Name:  javaToolOptionsEnvVarName,
		Value: "-Dcom.sun.jndi.rmiURLParsing=legacy",
	})

	if cc.Spec.JMXConfiguration != nil {
		jmxEnvVariable := GenerateJMXConfigEnvVar(*cc.Spec.JMXConfiguration)
		if jmxEnvVariable.Value != "" {
			commonCassandraEnvVars = append(commonCassandraEnvVars, jmxEnvVariable)
		}
	}

	customEnvVars := cc.Spec.EnvVars

	return deduplicateMerge(commonCassandraEnvVars, customEnvVars)
}

func GenerateJMXConfigEnvVar(jmxConf api.JMXConfiguration) v1.EnvVar {
	var jmxEnvVar v1.EnvVar
	var jmxParam string
	if jmxConf.JMXRemote != nil {
		jmxParam += JMXConfigurationMap["JMXRemote"] + strconv.FormatBool(*jmxConf.JMXRemote) + " "
	}
	if jmxConf.JXMRemoteSSL != nil {
		jmxParam += JMXConfigurationMap["JXMRemoteSSL"] + strconv.FormatBool(*jmxConf.JXMRemoteSSL) + " "
	}
	if jmxConf.JMXRemoteAuthenticate != nil {
		jmxParam += JMXConfigurationMap["JMXRemoteAuthenticate"] + strconv.FormatBool(*jmxConf.JMXRemoteAuthenticate) + " "
	}
	if jmxConf.JMXRemotePort != 0 {
		jmxParam += JMXConfigurationMap["JMXRemotePort"] + strconv.Itoa(jmxConf.JMXRemotePort) + " "
	}
	if jmxConf.JMXRemoteRmiPort != 0 {
		jmxParam += JMXConfigurationMap["JMXRemoteRmiPort"] + strconv.Itoa(jmxConf.JMXRemoteRmiPort) + " "
	}
	jmxEnvVar = v1.EnvVar{Name: jvmOptsEnvVarName, Value: jmxParam}
	return jmxEnvVar
}

func InitContainerEnvVar(cc *api.CassandraCluster, status *api.CassandraClusterStatus,
	resources v1.ResourceRequirements, dcRackName string) []v1.EnvVar {
	seedList := cc.SeedList(&status.SeedList)

	image := strings.Split(cc.Spec.CassandraImage, ":")
	serverVersion := cc.Spec.ServerVersion
	if serverVersion == "" {
		if len(image) >= 2 {
			version := strings.Split(image[len(image)-1], "-")
			serverVersion = version[0]
			if len(version) != 1 {
				serverVersion += ".0"
			}
		}
	}

	serverType := cc.Spec.ServerType
	if serverType == "" {
		if strings.Contains(image[0], "dse") {
			serverType = "dse"
		} else {
			serverType = "cassandra"
		}
	}

	defaultConfig := NodeConfig{
		"cassandra-yaml": {
			"read_request_timeout_in_ms":          5000,
			"write_request_timeout_in_ms":         5000,
			"counter_write_request_timeout_in_ms": 5000,
		},
		"logback-xml": {
			"debuglog-enabled": false,
		},
	}

	dcName := cc.GetDCNameFromDCRackName(dcRackName)

	config := NodeConfig{
		"cluster-info": {
			"name":  cc.GetName(),
			"seeds": seedList,
		},
		"datacenter-info": {
			"name": dcName,
		},
	}

	parsedConfig := parseConfig(config)
	dc := cc.GetDCFromDCRackName(dcRackName)
	rack := cc.GetRackFromDCRackName(dcRackName)

	mergeConfig(cc.Spec.Config, parsedConfig, serverVersion)
	mergeConfig(dc.Config, parsedConfig, serverVersion)
	mergeConfig(rack.Config, parsedConfig, serverVersion)

	defaultConfig[jvmOptionName(cc)] = map[string]interface{}{
		"initial_heap_size":       defineJvmMemory(resources).initialHeapSize,
		"max_heap_size":           defineJvmMemory(resources).maxHeapSize,
		"cassandra_ring_delay_ms": 30000,
		"jmx-connection-type":     "remote-no-auth",
	}

	for key, value := range defaultConfig {
		for subkey, subvalue := range value {
			keyPath := fmt.Sprintf("%s.%s", key, subkey)
			if parsedConfig.Path(keyPath).Data() == nil {
				parsedConfig.SetP(subvalue, keyPath)
			}
		}
	}

	return []v1.EnvVar{
		{
			Name:  "CONFIG_FILE_DATA",
			Value: parsedConfig.String(),
		},
		{
			Name:  "CONFIG_OUTPUT_DIRECTORY",
			Value: "/bootstrap",
		},
		{
			Name: "RACK_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.labels['cassandraclusters.db.orange.com.rack']",
				},
			},
		},
		{
			Name:  "PRODUCT_NAME",
			Value: serverType,
		},
		{
			Name:  "PRODUCT_VERSION",
			Value: serverVersion,
		},
		{
			Name: "POD_IP",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		},
		{
			Name: "HOST_IP",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.hostIP",
				},
			},
		},
	}
}

func commonEnvVars(cc *api.CassandraCluster) []v1.EnvVar {
	envVars := []v1.EnvVar{
		{
			Name: podIpEnvVarName,
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		},
		{
			Name:  cassandraLogDirEnvVarName,
			Value: "/var/log/cassandra",
		},
	}
	if (cc.Spec.ImageJolokiaSecret != v1.LocalObjectReference{}) {
		jolokiaEnvVars := []v1.EnvVar{
			{
				Name: jolokiaUserEnvVarName,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: cc.Spec.ImageJolokiaSecret,
						Key:                  "username",
					},
				},
			},
			{
				Name: jolokiaPasswordEnvVarName,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: cc.Spec.ImageJolokiaSecret,
						Key:                  "password",
					},
				},
			},
			{
				Name:  cassandraAuthJolokiaEnvVarName,
				Value: "true",
			},
		}
		envVars = append(envVars, jolokiaEnvVars...)
	}
	return envVars
}

func jvmOptionName(cc *api.CassandraCluster) (jvmOption string) {
	jvmOption = "jvm-options"
	if strings.HasPrefix(cc.Spec.ServerVersion, "4") {
		jvmOption = "jvm-server-options"
	}
	return
}

func mergeConfig(config json.RawMessage, currentParsedConfig *gabs.Container, serverVersion string) {
	if config != nil {
		parsedConfig, _ := gabs.ParseJSON(config)
		if strings.HasPrefix(serverVersion, "4") && parsedConfig.Path("jvm-options") != nil {
			parsedConfig.SetP(parsedConfig.Path("jvm-options").Data(), "jvm-server-options")
			parsedConfig.DeleteP("jvm-options")
		}
		currentParsedConfig.MergeFn(parsedConfig,
			func(dest, source interface{}) interface{} { return source })
	}
}

func parseConfig(config NodeConfig) *gabs.Container {
	generatedConfig, _ := json.Marshal(config)
	parsedConfig, _ := gabs.ParseJSON(generatedConfig)
	return parsedConfig
}

func defineJvmMemory(resources v1.ResourceRequirements) JvmMemory {

	var maxHeapSize, initialHeapSize string

	if !resources.Limits.Memory().IsZero() {
		mhsInBytes := float64(resources.Limits.Memory().Value()) / 4
		mhsInMB := int(mhsInBytes / float64(1024*1024))
		ihs := mhsInMB / 4 // Newheapsize = (container Mem)/8
		maxHeapSize = strings.Join([]string{strconv.Itoa(mhsInMB), "M"}, "")
		initialHeapSize = strings.Join([]string{strconv.Itoa(ihs), "M"}, "")
	} else {
		maxHeapSize = defaultJvmMaxHeap
		initialHeapSize = defaultJvmInitHeap
	}

	return JvmMemory{
		maxHeapSize:     maxHeapSize,
		initialHeapSize: initialHeapSize,
	}
}
