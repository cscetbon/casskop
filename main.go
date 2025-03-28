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

package main

import (
	"context"
	"flag"
	"fmt"
	gozap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"strconv"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	api "github.com/cscetbon/casskop/api/v2"
	"github.com/cscetbon/casskop/controllers/cassandrabackup"
	"github.com/cscetbon/casskop/controllers/cassandracluster"
	"github.com/cscetbon/casskop/controllers/cassandrarestore"
	"github.com/operator-framework/operator-lib/leader"
	"github.com/sirupsen/logrus"
	"github.com/zput/zxcTool/ztLog/zt_formatter"
)

// Change below variables to serve metrics on different host or port.
var (
	scheme   = apimachineryruntime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	logLevelEnvVar     = "LOG_LEVEL"
	resyncPeriodEnvVar = "RESYNC_PERIOD"
)

// to be set by compilator with -ldflags "-X main.compileDate=`date -u +.%Y%m%d.%H%M%S`"
var compileDate string

// to be set by compilator with -ldflags "-X main.version=Major.Minor.Patch"
var version string

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("casskop Version: %v", version)
	logrus.Infof("casskop Compilation Date: %s", compileDate)
	logrus.Infof("casskop LogLevel: %v", getLogLevel())
	logrus.Infof("casskop ResyncPeriod: %v", getResyncPeriod())
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

func zapLogLevel(level logrus.Level) zapcore.Level {
	switch level {
	case logrus.DebugLevel:
		return zapcore.DebugLevel
	case logrus.WarnLevel:
		return zapcore.WarnLevel
	case logrus.ErrorLevel:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func getResyncPeriod() int {
	var resyncPeriod int
	var err error
	resync, found := os.LookupEnv(resyncPeriodEnvVar)
	if !found {
		resyncPeriod = api.DefaultResyncPeriod
	} else {
		resyncPeriod, err = strconv.Atoi(resync)
		if err != nil {
			resyncPeriod = api.DefaultResyncPeriod
		}
	}
	return resyncPeriod
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(api.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	logLevel := getLogLevel()
	logrus.SetLevel(logLevel)
	var metricsAddr string
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	opts := zap.Options{
		Level:       gozap.NewAtomicLevelAt(zapLogLevel(logLevel)),
		TimeEncoder: zapcore.TimeEncoderOfLayout(time.RFC3339),
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	if logLevel == logrus.DebugLevel {
		ztFormatter := &zt_formatter.ZtFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := path.Base(f.File)
				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
			},
		}
		logrus.SetReportCaller(true)
		logrus.SetFormatter(ztFormatter)
		opts.Development = true
	}
	if logType, _ := os.LookupEnv("LOG_TYPE"); logType == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	printVersion()

	namespace, err := getWatchNamespace()
	if err != nil {
		setupLog.Error(err, "unable to get WatchNamespace, "+
			"the manager will watch and manage resources in all namespaces")
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctx := context.TODO()

	// Become the leader before proceeding
	err = leader.Become(ctx, "casskop-lock")
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                server.Options{BindAddress: metricsAddr},
		Cache:                  cache.Options{DefaultNamespaces: map[string]cache.Config{namespace: {}}},
		HealthProbeBindAddress: probeAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&cassandrabackup.CassandraBackupReconciler{
		Client:    mgr.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("CassandraBackup"),
		Scheme:    mgr.GetScheme(),
		Recorder:  mgr.GetEventRecorderFor("cassandrabackup-controller"),
		Scheduler: cassandrabackup.NewScheduler(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CassandraBackup")
		os.Exit(1)
	}
	if err = (&cassandracluster.CassandraClusterReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("CassandraCluster"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CassandraCluster")
		os.Exit(1)
	}
	if err = (&cassandrarestore.CassandraRestoreReconciler{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("controllers").WithName("CassandraRestore"),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("cassandrabackup-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CassandraRestore")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	logrus.Info("Starting the Cmd.")

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	// Start the Cmd
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
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
