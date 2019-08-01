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
	"testing"

	"github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"

	test "github.com/KohlsTechnology/eunomia/test"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestNone(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	test.AddToFrameworkSchemeForTests(t, ctx)
	simpleTestDeploy(t, framework.Global, ctx)
}

func noneTestDeploy(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {
	namespace, err := ctx.GetNamespace()
	assert.NoError(t, err)

	// Check if the CRD has been created
	crd := &gitopsv1alpha1.GitOpsConfig{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "gitops-simple", Namespace: namespace}, crd)
	assert.Error(t, err)

	gitops := &v1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-none",
			Namespace: namespace,
		},
		Spec: gitopsv1alpha1.GitOpsConfigSpec{
			TemplateSource: gitopsv1alpha1.GitConfig{
				URI:        "https://",
				Ref:        "master",
				ContextDir: "/",
			},
			ParameterSource: gitopsv1alpha1.GitConfig{
				URI:        "https://",
				Ref:        "master",
				ContextDir: "/",
			},
			Triggers: []gitopsv1alpha1.GitOpsTrigger{
				{
					Type: "Change",
				},
			},
			ResourceDeletionMode: "None",
			ResourceHandlingMode: "None",
			ServiceAccountRef:    "eunomia-operator",
		},
	}

	err = f.Client.Create(goctx.TODO(), gitops, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	assert.NoError(t, err)

	// Check if the CRD has been created
	crd = &gitopsv1alpha1.GitOpsConfig{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "gitops-none", Namespace: namespace}, crd)
	assert.NoError(t, err)
}
