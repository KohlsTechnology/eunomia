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

package e2e

import (
	goctx "context"
	"fmt"
	gitopsv1alpha1 "gitops-operator/pkg/apis/eunomia/v1alpha1"
	test "gitops-operator/test"
	"testing"
	"time"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
	name                 = "gitops"
)

func TestOCPTemplate(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	//test.Initialize()
	test.AddToFrameworkSchemeForTests(t, ctx)
	if err := ocpTemplateTestDeploy(t, framework.Global, ctx); err != nil {
		t.Fatal(err)
	}
}

func ocpTemplateTestDeploy(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	gitops := &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops",
			Namespace: namespace,
		},
		Spec: gitopsv1alpha1.GitOpsConfigSpec{
			TemplateSource: gitopsv1alpha1.GitConfig{
				URI:        "https://github.com/KohlsTechnology/eunomia",
				Ref:        "master",
				ContextDir: "example/templates",
			},
			ParameterSource: gitopsv1alpha1.GitConfig{
				URI:        "https://github.com/KohlsTechnology/eunomia",
				Ref:        "master",
				ContextDir: "example/parameters",
			},
			Triggers: []gitopsv1alpha1.GitOpsTrigger{
				{
					Type: "Change",
				},
			},
			ResourceDeletionMode:   "Delete",
			TemplateProcessorImage: "quay.io/kohlstechnology/eunomia-ocp-templates:v0.0.1",
			ResourceHandlingMode:   "CreateOrMerge",
			ServiceAccountRef:      "gitops-operator",
		},
	}
	gitops.Annotations = map[string]string{"gitopsconfig.eunomia.kohls.io/initialized": "true"}

	err = f.Client.Create(goctx.TODO(), gitops, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		return err
	}

	// Check if the CRD has been created
	crd := &gitopsv1alpha1.GitOpsConfig{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "gitops", Namespace: namespace}, crd)
	fmt.Printf("test %v+", crd)

	return WaitForPod(t, f, ctx, namespace, "hello-openshift", retryInterval, timeout)
}
