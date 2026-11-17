package envVars

import (
	"encoding/json"
	"testing"

	"github.com/Jeffail/gabs"
	api "github.com/cscetbon/casskop/api/v2"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestCassandraEnvVar_givenEnvVars_thenSet(t *testing.T) {
	// given
	cc := &api.CassandraCluster{
		Spec: api.CassandraClusterSpec{
			EnvVars: []v1.EnvVar{
				{Name: "foo", Value: "test-env-var-value"},
				{Name: "bar", Value: "another-env-var-value"},
			},
		},
	}

	// when
	actualEnvVars := CassandraEnvVars(cc)

	// then
	fooEnvVars := findAllEnvVarsByName(actualEnvVars, "foo")
	barEnvVars := findAllEnvVarsByName(actualEnvVars, "bar")

	assert.Equal(t, 1, len(fooEnvVars))
	assert.Equal(t, 1, len(barEnvVars))

	assert.Equal(t, fooEnvVars[0].Value, "test-env-var-value")
	assert.Equal(t, barEnvVars[0].Value, "another-env-var-value")
}

func TestCassandraEnvVar_givenChangesToCommonEnvVars_thenOverride(t *testing.T) {
	// given
	cc := &api.CassandraCluster{
		Spec: api.CassandraClusterSpec{
			ImageJolokiaSecret: v1.LocalObjectReference{
				Name: "test-jolokia-secret",
			},
			EnvVars: []v1.EnvVar{
				{Name: podIpEnvVarName, Value: "override-pod-ip"},
				{Name: cassandraLogDirEnvVarName, Value: "override-cassandra-log-dir"},
				{Name: jolokiaUserEnvVarName, Value: "override-jolokia-user"},
				{Name: jolokiaPasswordEnvVarName, Value: "override-jolokia-password"},
				{Name: cassandraAuthJolokiaEnvVarName, Value: "override-cassandra-auth-jolokia"},
			},
		},
	}

	// when
	actualEnvVars := CassandraEnvVars(cc)

	// then
	podIpEnvVars := findAllEnvVarsByName(actualEnvVars, podIpEnvVarName)
	cassandraLogDirEnvVars := findAllEnvVarsByName(actualEnvVars, cassandraLogDirEnvVarName)
	jolokiaUserEnvVars := findAllEnvVarsByName(actualEnvVars, jolokiaUserEnvVarName)
	jolokiaPasswordEnvVars := findAllEnvVarsByName(actualEnvVars, jolokiaPasswordEnvVarName)
	cassandraAuthJolokiaEnvVars := findAllEnvVarsByName(actualEnvVars, cassandraAuthJolokiaEnvVarName)

	assert.Equal(t, 1, len(podIpEnvVars))
	assert.Equal(t, 1, len(cassandraLogDirEnvVars))
	assert.Equal(t, 1, len(jolokiaUserEnvVars))
	assert.Equal(t, 1, len(jolokiaPasswordEnvVars))
	assert.Equal(t, 1, len(cassandraAuthJolokiaEnvVars))

	assert.Equal(t, "override-pod-ip", podIpEnvVars[0].Value)
	assert.Equal(t, "override-cassandra-log-dir", cassandraLogDirEnvVars[0].Value)
	assert.Equal(t, "override-jolokia-user", jolokiaUserEnvVars[0].Value)
	assert.Equal(t, "override-jolokia-password", jolokiaPasswordEnvVars[0].Value)
	assert.Equal(t, "override-cassandra-auth-jolokia", cassandraAuthJolokiaEnvVars[0].Value)
}

func TestBootstrapContainerEnvVar_givenEnvVars_thenSet(t *testing.T) {
	// given
	cc := &api.CassandraCluster{
		Spec: api.CassandraClusterSpec{
			BackRestSidecar: &api.BackRestSidecar{
				EnvVars: []v1.EnvVar{
					{Name: "foo", Value: "test-env-var-value"},
					{Name: "bar", Value: "another-env-var-value"},
				},
			},
		},
	}
	status := &api.CassandraClusterStatus{}

	// when
	actualEnvVars := BootstrapContainerEnvVar(cc, status)

	// then
	fooEnvVars := findAllEnvVarsByName(actualEnvVars, "foo")
	barEnvVars := findAllEnvVarsByName(actualEnvVars, "bar")

	assert.Equal(t, 1, len(fooEnvVars))
	assert.Equal(t, 1, len(barEnvVars))

	assert.Equal(t, fooEnvVars[0].Value, "test-env-var-value")
	assert.Equal(t, barEnvVars[0].Value, "another-env-var-value")
}

func TestBootstrapContainerEnvVar_givenChangesToCoreEnvVars_thenOverride(t *testing.T) {
	// given
	cc := &api.CassandraCluster{
		Name: "test-cluster",
		Spec: api.CassandraClusterSpec{
			ImageJolokiaSecret: v1.LocalObjectReference{
				Name: "test-jolokia-secret",
			},
			BackRestSidecar: &api.BackRestSidecar{
				EnvVars: []v1.EnvVar{
					{Name: cassandraClusterNameEnvVarName, Value: "override-cluster-name"},
					{Name: cassandraSeedsEnvVarName, Value: "override-seeds"},
					{Name: cassandraDcEnvVarName, Value: "override-cassandra-dc"},
					{Name: cassandraRackEnvVarName, Value: "override-cassandra-rack"},
					{Name: podIpEnvVarName, Value: "override-pod-ip"},
					{Name: cassandraLogDirEnvVarName, Value: "override-cassandra-log-dir"},
					{Name: jolokiaUserEnvVarName, Value: "override-jolokia-user"},
					{Name: jolokiaPasswordEnvVarName, Value: "override-jolokia-password"},
					{Name: cassandraAuthJolokiaEnvVarName, Value: "override-cassandra-auth-jolokia"},
				},
			},
		},
	}
	status := &api.CassandraClusterStatus{
		SeedList: []string{"seed1", "seed2", "seed3"},
	}

	// when
	actualEnvVars := BootstrapContainerEnvVar(cc, status)

	// then
	nameEnvVars := findAllEnvVarsByName(actualEnvVars, cassandraClusterNameEnvVarName)
	seedsEnvVars := findAllEnvVarsByName(actualEnvVars, cassandraSeedsEnvVarName)
	dcEnvVars := findAllEnvVarsByName(actualEnvVars, cassandraDcEnvVarName)
	rackEnvVars := findAllEnvVarsByName(actualEnvVars, cassandraRackEnvVarName)
	podIpEnvVars := findAllEnvVarsByName(actualEnvVars, podIpEnvVarName)
	cassandraLogDirEnvVars := findAllEnvVarsByName(actualEnvVars, cassandraLogDirEnvVarName)
	jolokiaUserEnvVars := findAllEnvVarsByName(actualEnvVars, jolokiaUserEnvVarName)
	jolokiaPasswordEnvVars := findAllEnvVarsByName(actualEnvVars, jolokiaPasswordEnvVarName)
	cassandraAuthJolokiaEnvVars := findAllEnvVarsByName(actualEnvVars, cassandraAuthJolokiaEnvVarName)

	assert.Equal(t, 1, len(nameEnvVars))
	assert.Equal(t, 1, len(seedsEnvVars))
	assert.Equal(t, 1, len(dcEnvVars))
	assert.Equal(t, 1, len(rackEnvVars))
	assert.Equal(t, 1, len(podIpEnvVars))
	assert.Equal(t, 1, len(cassandraLogDirEnvVars))
	assert.Equal(t, 1, len(jolokiaUserEnvVars))
	assert.Equal(t, 1, len(jolokiaPasswordEnvVars))
	assert.Equal(t, 1, len(cassandraAuthJolokiaEnvVars))

	assert.Equal(t, "override-cluster-name", nameEnvVars[0].Value)
	assert.Equal(t, "override-seeds", seedsEnvVars[0].Value)
	assert.Equal(t, "override-cassandra-dc", dcEnvVars[0].Value)
	assert.Equal(t, "override-cassandra-rack", rackEnvVars[0].Value)
	assert.Equal(t, "override-pod-ip", podIpEnvVars[0].Value)
	assert.Equal(t, "override-cassandra-log-dir", cassandraLogDirEnvVars[0].Value)
	assert.Equal(t, "override-jolokia-user", jolokiaUserEnvVars[0].Value)
	assert.Equal(t, "override-jolokia-password", jolokiaPasswordEnvVars[0].Value)
	assert.Equal(t, "override-cassandra-auth-jolokia", cassandraAuthJolokiaEnvVars[0].Value)
}

func findAllEnvVarsByName(envVars []v1.EnvVar, name string) []*v1.EnvVar {
	var result []*v1.EnvVar
	for _, envVar := range envVars {
		if envVar.Name == name {
			result = append(result, &envVar)
		}
	}
	return result
}

func TestTranslateLegacyCassandraYaml(t *testing.T) {
	section := map[string]any{
		"read_request_timeout_in_ms":                  float64(5000),
		"write_request_timeout":                       "3000ms",
		"write_request_timeout_in_ms":                 float64(2000),
		"compaction_throughput_mb_per_sec":            float64(64),
		"stream_throughput_outbound_megabits_per_sec": float64(24),
		"index_summary_resize_interval_in_minutes":    float64(-1),
		"cache_load_timeout_seconds":                  float64(-1),
		"commitlog_total_space_in_mb":                 float64(-1),
		"enable_user_defined_functions":               true,
		"commitlog_sync_group_window_in_ms":           15.5,
		"num_tokens":                                  float64(16),
		"key_cache_size_in_mb":                        "100MiB",
		"key_cache_save_period":                       float64(14400),
		"row_cache_save_period":                       "30m",
		"counter_cache_save_period":                   float64(0),
	}
	expected := map[string]any{
		"read_request_timeout":       "5000ms",
		"write_request_timeout":      "3000ms",
		"compaction_throughput":      "64MiB/s",
		"stream_throughput_outbound": "3000000B/s",
		// negative values: null when the setting is enabled by default, 0 when null is not accepted,
		// left out when the default is already null
		"index_summary_resize_interval":  nil,
		"cache_load_timeout":             "0s",
		"user_defined_functions_enabled": true,
		"commitlog_sync_group_window":    "15.5ms",
		"num_tokens":                     float64(16),
		"key_cache_size":                 "100MiB",
		// same name in 4.1, only the unit is added
		"key_cache_save_period":     "14400s",
		"row_cache_save_period":     "30m",
		"counter_cache_save_period": "0s",
	}
	assert.Equal(t, expected, translateLegacyCassandraYaml(section, false))
}

// The partition size and tombstone thresholds only got their guardrail names in 5.0, a 4.1 node
// refuses them
func TestTranslateLegacyCassandraYaml_renamedIn50(t *testing.T) {
	section := map[string]any{
		"compaction_large_partition_warning_threshold_mb": float64(200),
		"compaction_tombstone_warning_threshold":          float64(50000),
	}
	assert.Equal(t, map[string]any{
		"compaction_large_partition_warning_threshold": "200MiB",
		"compaction_tombstone_warning_threshold":       float64(50000),
	}, translateLegacyCassandraYaml(section, false))
	assert.Equal(t, map[string]any{
		"partition_size_warn_threshold":       "200MiB",
		"partition_tombstones_warn_threshold": float64(50000),
	}, translateLegacyCassandraYaml(section, true))
	assert.Equal(t, map[string]any{"partition_size_warn_threshold": "300MiB"},
		translateLegacyCassandraYaml(map[string]any{"compaction_large_partition_warning_threshold": "300MiB"}, true))
}

// A legacy key set on a rack must win over the new style key set on the cluster
func TestMergeConfig_translatesEachLevel(t *testing.T) {
	merged := gabs.New()
	translate := func(config *gabs.Container) { translateLegacyCassandraYamlSection(config, false) }
	mergeConfig(json.RawMessage(`{"cassandra-yaml": {"read_request_timeout": "1000ms"}}`), merged, jvmServerOptionsKey, translate)
	mergeConfig(json.RawMessage(`{"cassandra-yaml": {"read_request_timeout_in_ms": 2000}}`), merged, jvmServerOptionsKey, translate)
	assert.Equal(t, map[string]any{"read_request_timeout": "2000ms"}, merged.Path(cassandraYamlKey).Data())
}
