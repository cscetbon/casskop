package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	var expects multiFlag
	var absents multiFlag
	flag.Var(&expects, "expect", "container:VAR_NAME=VALUE;container:VAR_NAME=VALUE;...")
	flag.Var(&absents, "absent", "container:VAR_NAME;container:VAR_NAME;...")
	flag.Parse()

	expectedEnvVars := parseEnvVars(expects)
	absentEnvVars := parseEnvVars(absents)

	ss := getSts()

	for containerName, envVars := range expectedEnvVars {
		for _, c := range ss.Spec.Template.Spec.Containers {
			if c.Name == containerName {
				envsString := c.Env.String()
				for _, envVar := range envVars {
					fmt.Println("kbannach: checking expected env var:", "all", envsString, "expect", envVar.String())
					if !strings.Contains(envsString, envVar.String()) {
						fmt.Printf("EnvVars expected to be present on container %s are not found! \nExpected present: %s\nAll: %s\n", containerName, envVars, c.Env)
						os.Exit(1)
					}
				}
			}
		}
	}

	for containerName, envVars := range absentEnvVars {
		for _, c := range ss.Spec.Template.Spec.Containers {
			if c.Name == containerName {
				envsString := c.Env.String()
				for _, envVar := range envVars {
					fmt.Println("kbannach: checking absent env var:", "all", envsString, "absent", envVar.String())
					if strings.Contains(envsString, envVar.String()) {
						fmt.Printf("EnvVars expected to be absent on container %s found! \nExpected absent: %s\nAll: %s\n", containerName, envVars, c.Env)
						os.Exit(1)
					}
				}
			}
		}
	}
}
