/*
Copyright 2018 The Multicluster-Controller Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"admiralty.io/multicluster-controller/pkg/cluster"
	"admiralty.io/multicluster-controller/pkg/manager"
	"admiralty.io/multicluster-service-account/pkg/config"
	"github.com/cscetbon/casskop/multi-casskop/controllers"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/sample-controller/pkg/signals"
	kconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	logLevelEnvVar = "LOG_LEVEL"
)

// Change below variables to serve metrics on different host or port.
var (
	scheme = apimachineryruntime.NewScheme()
)

// to be set by compilator with -ldflags "-X main.compileDate=`date -u +.%Y%m%d.%H%M%S`"
var compileDate string

// to be set by compilator with -ldflags "-X main.version=Major.Minor.Patch"
var version string

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("multi-casskop Version: %v", version)
	logrus.Infof("multi-casskop Compilation Date: %s", compileDate)
}

func getLogLevel() logrus.Level {
	logLevel, found := os.LookupEnv(logLevelEnvVar)
	if !found {
		return logrus.InfoLevel
	}
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		return logrus.DebugLevel
	case "INFO":
		return logrus.InfoLevel
	case "ERROR":
		return logrus.ErrorLevel
	case "WARN":
		return logrus.WarnLevel
	}
	return logrus.InfoLevel
}

type Options struct {
	// Slice of bool will append 'true' each time the option
	// is encountered (can be set multiple times, like -vvv)
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`

	Local string `short:"l" description:"The local k8s cluster name"`

	// Example of a slice of strings
	Remotes []string `short:"r" description:"The remotes k8S cluster name list"`
}

var opts Options

var parser = flags.NewParser(&opts, flags.Default)

func main() {

	logType, found := os.LookupEnv("LOG_TYPE")
	if found && logType == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	logrus.SetLevel(getLogLevel())

	printVersion()

	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	var clusters controllers.Clusters

	namespace, err := getWatchNamespace()
	if err != nil {
		logrus.Error(err, "unable to get WatchNamespace, "+
			"the manager will watch and manage resources in all namespaces")
	}

	// Configuring Local cluster
	logrus.Infof("Configuring Client for local cluster %s. using local k8s api access",
		opts.Local)

	cfg, err := kconfig.GetConfig()
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	// Set up local cluster
	clusters.Local = &controllers.Cluster{Name: opts.Local, Cluster: cluster.New(opts.Local, cfg,
		cluster.Options{CacheOptions: cluster.CacheOptions{Namespace: namespace}})}

	// Configuring Remotes clusters
	for i, remote := range opts.Remotes {

		var cfg *rest.Config
		var err error

		logrus.Infof("Configuring Client %d for remote cluster %s. using imported secret of same name", i+1,
			remote)
		cfg, _, err = config.NamedConfigAndNamespace(remote)
		if err != nil {
			log.Fatal(err)
		}

		// Set up Remotes clusters
		clusters.Remotes = append(clusters.Remotes,
			&controllers.Cluster{Name: remote, Cluster: cluster.New(remote, cfg,
				cluster.Options{CacheOptions: cluster.CacheOptions{Namespace: namespace}})})
	}

	logrus.Info("Creating Controller")
	co, err := controllers.NewController(clusters, namespace)
	if err != nil {
		log.Fatalf("creating Cassandra Multi Cluster controller: %v", err)
	}

	m := manager.New()
	m.AddController(co)
	// TODO: create more meaningful checks
	http.HandleFunc("/readyz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})
	http.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})
	go func() {
		err := http.ListenAndServe(":8081", nil)
		if err != nil {
			log.Fatalf("failed to run readyness/liveness checks")
		}
	}()

	logrus.Info("Starting Manager.")
	if err := m.Start(signals.SetupSignalHandler()); err != nil {
		log.Fatalf("while or after starting manager: %v", err)
	}
}

// getWatchNamespace returns the Namespace the operator should be watching for changes
func getWatchNamespace() (string, error) {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	var watchNamespaceEnvVar = "WATCH_NAMESPACE"

	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}
	return ns, nil
}
