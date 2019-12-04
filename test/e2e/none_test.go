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

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/KohlsTechnology/eunomia/pkg/apis"
	"github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
)

func TestNone(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("could not get namespace: %v", err)
	}
	err = framework.AddToFrameworkScheme(apis.AddToScheme, &gitopsv1alpha1.GitOpsConfigList{})
	if err != nil {
		t.Fatal(err)
	}

	// Check if the CRD has been created
	err = framework.Global.Client.Get(
		goctx.TODO(),
		types.NamespacedName{Name: "gitops-simple", Namespace: namespace},
		&gitopsv1alpha1.GitOpsConfig{})
	if err == nil {
		t.Error("expected error, got nil")
	}

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
				{Type: "Change"},
			},
			ResourceDeletionMode: "None",
			ResourceHandlingMode: "None",
			ServiceAccountRef:    "eunomia-operator",
		},
	}

	err = framework.Global.Client.Create(
		goctx.TODO(),
		gitops,
		&framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	// Check if the CRD has been created
	err = framework.Global.Client.Get(
		goctx.TODO(),
		types.NamespacedName{Name: "gitops-none", Namespace: namespace},
		&gitopsv1alpha1.GitOpsConfig{})
	if err != nil {
		t.Error(err)
	}
}
