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

package unittest

import (
	"flag"
	"os"
	"testing"

	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/KohlsTechnology/eunomia/pkg/apis"
	v1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	"github.com/KohlsTechnology/eunomia/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

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
	jt, found := os.LookupEnv("JOB_TEMPLATE")
	if !found {
		logf.Log.Info("Error: JOB_TEMPLATE must be set")
	}
	cjt, found := os.LookupEnv("CRONJOB_TEMPLATE")
	if !found {
		logf.Log.Info("Error: CRONJOB_TEMPLATE must be set")
	}
	util.InitializeTemplates(jt, cjt)
}

func AddToFrameworkSchemeForTests(t *testing.T, ctx *framework.TestCtx) {
	namespace, err := ctx.GetNamespace()
	assert.NoError(t, err)
	gitops := &v1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-operator",
			Namespace: namespace,
		},
		Spec: v1alpha1.GitOpsConfigSpec{},
	}

	assert.NoError(t, framework.AddToFrameworkScheme(apis.AddToScheme, gitops))
}
