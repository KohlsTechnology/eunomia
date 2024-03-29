/*
Copyright 2019 Kohl's Department Stores, Inc.

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
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/KohlsTechnology/eunomia/pkg/apis"
	"github.com/KohlsTechnology/eunomia/pkg/controller"
	"github.com/KohlsTechnology/eunomia/pkg/controller/gitopsconfig"
	"github.com/KohlsTechnology/eunomia/pkg/handler"
	"github.com/KohlsTechnology/eunomia/pkg/util"
	"github.com/KohlsTechnology/eunomia/version"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost       = "0.0.0.0"
	metricsPort int32 = 8383
)
var log = logf.Log.WithName("cmd").WithValues("filename", "main.go")

func printVersion() {
	log.Info(fmt.Sprintf("Eunomia version: %s (build date: %s, branch: %s, git SHA1: %s)", version.Version, version.BuildDate, version.Branch, version.GitSHA1))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
}

func main() {
	// Add the zap logger flag set to the CLI. The flag set must
	// be added before calling pflag.Parse().
	pflag.CommandLine.AddFlagSet(zap.FlagSet())

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	versionFlag := pflag.Bool("version", false, "print version information and exit")

	pflag.Parse()

	// Use a zap logr.Logger implementation. If none of the zap
	// flags are configured (or if the zap flag set is not being
	// used), this defaults to a production zap logger.
	//
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(zap.Logger())

	printVersion()
	if *versionFlag {
		os.Exit(0)
	}

	// Register a Prometheus-formatted metric displaying app version & other useful build info.
	buildInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "eunomia",
			Name:      "build_info",
			Help: "A metric with a constant '1' value labeled by version from " +
				"which eunomia was built, and other useful build information.",
		},
		[]string{
			"version", "builddate", "branch", "gitsha1",
			"goversion", "operatorsdk",
		},
	)
	buildInfo.WithLabelValues(
		version.Version, version.BuildDate, version.Branch, version.GitSHA1,
		runtime.Version(), sdkVersion.Version).Set(1)
	metrics.Registry.MustRegister(buildInfo)

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.Error(err, "Failed to get watch namespace")
		os.Exit(1)
	}

	// initialize the templates
	jt, found := os.LookupEnv("JOB_TEMPLATE")
	if !found {
		log.Info("JOB_TEMPLATE not set. Using default job template.")
		jt = "/default-job-templates/job.yaml"
	}
	cjt, found := os.LookupEnv("CRONJOB_TEMPLATE")
	if !found {
		log.Info("CRONJOB_TEMPLATE not set. Using default job template.")
		cjt = "/default-job-templates/cronjob.yaml"
	}
	util.InitializeTemplates(jt, cjt) //nolint:errcheck
	log.Info("Templates initialized correctly")

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	ctx := context.TODO()

	// Become the leader before proceeding
	err = leader.Become(ctx, "eunomia-lock")
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          namespace,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create Service object to expose the metrics port.
	// commented because service is generated via a manifest at deploy time.
	// servicePorts := []v1.ServicePort{
	// 	{Port: metricsPort, Name: metrics.OperatorPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: metricsPort}},
	// 	{Port: operatorMetricsPort, Name: metrics.CRPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: operatorMetricsPort}},
	// }
	//
	// _, err = metrics.CreateMetricsService(ctx, servicePorts)

	// Set up WebHook listener and healthz endpoint

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/", func(w http.ResponseWriter, r *http.Request) {
		handler.WebhookHandler(w, r, gitopsconfig.NewReconciler(mgr))
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok")) //nolint:errcheck
	})

	// Get namespaces (as there for sure will be some) directly from Kubernetes to ensure that
	// there is access to Kubernetes API, and there are no errors when calling it.
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		namespaces := corev1.NamespaceList{}
		err = mgr.GetClient().List(ctx, &namespaces, []client.ListOption{}...)
		if err != nil {
			log.Error(err, "namespaces listing for /readyz endpoint failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok")) //nolint:errcheck
	})

	log.Info("Starting the Web Server")
	go http.ListenAndServe(":8080", mux) //nolint:errcheck

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}
