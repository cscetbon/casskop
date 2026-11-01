package envVars

import (
	"encoding/json"
	"testing"

	api "github.com/cscetbon/casskop/api/v2"
	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	jvmServer := json.RawMessage(`{"jvm-server-options": {"max_heap_size": "1G"}}`)
	jvm17 := json.RawMessage(`{"jvm17-server-options": {"additional-jvm-opts": ["-Dfoo=bar"]}}`)

	tests := []struct {
		name    string
		image   string
		config  json.RawMessage
		wantErr string
	}{
		{"no config", "cassandra:3.11.19", nil, ""},
		{"legacy jvm-options on 3.11", "cassandra:3.11.19", json.RawMessage(`{"jvm-options": {"log_gc": "true"}}`), ""},
		{"legacy jvm-options on 4.1 are moved", "cassandra:4.1.12", json.RawMessage(`{"jvm-options": {"log_gc": "true"}}`), ""},
		{"jvm-server-options on 3.11", "cassandra:3.11.19", jvmServer,
			"spec.config: section jvm-server-options needs Cassandra 4.0 or later, the cluster runs 3.11.19"},
		{"jvm-server-options on 4.0", "cassandra:4.0.21", jvmServer, ""},
		{"jvm11-server-options on 3.11", "cassandra:3.11.19", json.RawMessage(`{"jvm11-server-options": {}}`),
			"section jvm11-server-options needs Cassandra 4.0 or later"},
		{"jvm17-server-options on 4.1", "cassandra:4.1.12", jvm17,
			"section jvm17-server-options needs Cassandra 5.0 or later, the cluster runs 4.1.12"},
		{"jvm17-server-options on 5.0", "cassandra:5.0.9", jvm17, ""},
		{"not an object", "cassandra:5.0.9", json.RawMessage(`[1, 2]`), "spec.config is not a JSON object"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &api.CassandraCluster{Spec: api.CassandraClusterSpec{CassandraImage: tt.image, Config: tt.config}}
			err := ValidateConfig(cc)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_dcAndRackSections(t *testing.T) {
	cc := &api.CassandraCluster{Spec: api.CassandraClusterSpec{
		CassandraImage: "cassandra:4.0.21",
		Topology: api.Topology{DC: []api.DC{{
			Name:   "dc1",
			Config: json.RawMessage(`{"jvm17-server-options": {}}`),
			Rack: []api.Rack{
				{Name: "rack1", Config: json.RawMessage(`{"jvm-server-options": {}}`)},
				{Name: "rack2", Config: json.RawMessage(`{"jvm17-server-options": {}}`)},
			},
		}}},
	}}
	err := ValidateConfig(cc)
	assert.ErrorContains(t, err, "dc dc1 config: section jvm17-server-options needs Cassandra 5.0 or later")
	assert.ErrorContains(t, err, "dc dc1 rack rack2 config: section jvm17-server-options")
	assert.NotContains(t, err.Error(), "rack1")
}
