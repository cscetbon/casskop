package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (e EnvVar) String() string {
	return fmt.Sprintf("%s=%s", e.Name, e.Value)
}

type EnvVarSlice []EnvVar

func (e EnvVarSlice) String() string {
	var parts []string
	for _, envVar := range e {
		parts = append(parts, envVar.String())
	}
	return strings.Join(parts, ",")
}

type Container struct {
	Name string      `json:"name"`
	Env  EnvVarSlice `json:"env"`
}

type PodSpec struct {
	Containers []Container `json:"containers"`
}

type PodTemplate struct {
	Spec PodSpec `json:"spec"`
}

type StatefulSetSpec struct {
	Template PodTemplate `json:"template"`
}

type StatefulSet struct {
	Spec StatefulSetSpec `json:"spec"`
}

// multiFlag allows repeated flags
type multiFlag []string

var _ flag.Value = &multiFlag{}

func (m *multiFlag) String() string {
	return fmt.Sprintf("%v", *m)
}

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func getSts() StatefulSet {
	out, err := exec.Command("kubectl", "get", "statefulset", "cassandra-e2e-dc1-rack1", "-o", "json").Output()
	if err != nil {
		fmt.Println("Failed to get StatefulSet:", err)
		os.Exit(1)
	}

	var sts StatefulSet
	if err := json.Unmarshal(out, &sts); err != nil {
		fmt.Println("Failed to parse StatefulSet JSON:", err)
		os.Exit(1)
	}

	return sts
}

const (
	containerNameSeparator = ":"
	envVarSeparator        = "="
)

// parseEnvVars parses arguments in the format `container:VAR_NAME=VALUE;container:VAR_NAME=VALUE`
// or `container:VAR_NAME;container:VAR_NAME` into a map of container names to EnvVar slices.
//
// Example:
//
//	args := multiFlag{"cassandra:MY_VAR=value1", "sidecar:OTHER_VAR=value2"}
//	result := parseEnvVars(args)
//	// result["cassandra"] = []EnvVar{{Name: "MY_VAR", Value: "value1"}}
func parseEnvVars(args multiFlag) map[string][]EnvVar {
	result := make(map[string][]EnvVar)

	for _, singleArg := range args {
		split := strings.SplitN(singleArg, containerNameSeparator, 2)
		validateTwoPartArgument(split, singleArg)

		containerName := split[0]
		envVarParts := strings.SplitN(split[1], envVarSeparator, 2)

		envVar := EnvVar{}
		envVar.Name = envVarParts[0]
		if len(envVarParts) == 2 {
			envVar.Value = envVarParts[1]
		}
		result[containerName] = append(result[containerName], envVar)
	}

	return result
}

func validateTwoPartArgument(argument []string, inputString string) {
	if len(argument) != 2 {
		fmt.Printf("Invalid argument format: %s\n", inputString)
		os.Exit(1)
	}
}

func contains(envVars []EnvVar) func(EnvVar) bool {
	return func(expected EnvVar) bool {
		for _, e := range envVars {
			if expected.Name != e.Name && expected.Value != e.Value {
				return false
			}
		}
		return true
	}
}
