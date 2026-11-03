package envVars

import (
	"encoding/json"
	"errors"
	"fmt"

	api "github.com/cscetbon/casskop/api/v2"
)

// configSectionMinVersions lists the config sections the config builder only knows from a given
// Cassandra version. Passing one of them to an older version makes every init container fail with
// "Invalid config keys", so the operator refuses such a configuration up front. The legacy
// jvm-options section is not listed as it is moved to jvm-server-options on 4.0 and later.
var configSectionMinVersions = []struct {
	section      string
	major, minor int
}{
	{jvmServerOptionsKey, 4, 0},
	{"jvm8-server-options", 4, 0},
	{"jvm11-server-options", 4, 0},
	{"jvm17-server-options", 5, 0},
}

// ValidateConfig checks the config of the cluster, its DCs and its racks against the Cassandra
// version the cluster runs and returns every problem found
func ValidateConfig(cc *api.CassandraCluster) error {
	errs := []error{validateConfigSections(cc, cc.Spec.Config, "spec.config")}
	for _, dc := range cc.Spec.Topology.DC {
		errs = append(errs, validateConfigSections(cc, dc.Config, fmt.Sprintf("dc %s config", dc.Name)))
		for _, rack := range dc.Rack {
			errs = append(errs, validateConfigSections(cc, rack.Config, fmt.Sprintf("dc %s rack %s config", dc.Name, rack.Name)))
		}
	}
	return errors.Join(errs...)
}

func validateConfigSections(cc *api.CassandraCluster, config json.RawMessage, location string) error {
	if len(config) == 0 {
		return nil
	}
	var sections map[string]json.RawMessage
	if err := json.Unmarshal(config, &sections); err != nil {
		return fmt.Errorf("%s is not a JSON object: %w", location, err)
	}
	for _, s := range configSectionMinVersions {
		if _, found := sections[s.section]; found && !cc.ServerVersionAtLeast(s.major, s.minor) {
			return fmt.Errorf("%s: section %s needs Cassandra %d.%d or later, the cluster runs %s",
				location, s.section, s.major, s.minor, cc.ServerVersion())
		}
	}
	return nil
}
