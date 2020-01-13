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

package test

import (
	"flag"
	"path/filepath"
	"runtime"

	"github.com/KohlsTechnology/eunomia/pkg/util"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/spf13/pflag"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // Make linter tmp happy
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// Initialize infrastructure for eunomia unit tests
func Initialize() {

	logf.Log.Info("Initializing Test")

	// Add the zap logger flag set to the CLI. The flag set must
	// be added before calling pflag.Parse().
	pflag.CommandLine.AddFlagSet(zap.FlagSet())

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

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

	// initialize the templates
	_, initFilename, _, _ := runtime.Caller(0)
	eunomiaRoot := filepath.Join(filepath.Dir(initFilename), "..")
	util.InitializeTemplates(
		filepath.Join(eunomiaRoot, "./build/job-templates/job.yaml"),
		filepath.Join(eunomiaRoot, "./build/job-templates/cronjob.yaml"),
	)
}
