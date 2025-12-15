package envVars

import (
	"testing"

	api "github.com/cscetbon/casskop/api/v2"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-cluster",
		},
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
