// +build e2e

/*
Copyright 2020 Kohl's Department Stores, Inc.

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
	"testing"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
)

func TestHierarchy(t *testing.T) {
	ctx, err := NewContext(t)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Cleanup()

	gitops := &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-hierarchy",
			Namespace: ctx.namespace,
		},
		Spec: gitopsv1alpha1.GitOpsConfigSpec{
			TemplateProcessorArgs: "--set namespace=" + ctx.namespace,
			TemplateSource: gitopsv1alpha1.GitConfig{
				URI:        ctx.eunomiaURI,
				Ref:        ctx.eunomiaRef,
				ContextDir: "test/e2e/testdata/helm/templates",
			},
			ParameterSource: gitopsv1alpha1.GitConfig{
				URI:        ctx.eunomiaURI,
				Ref:        ctx.eunomiaRef,
				ContextDir: "test/e2e/testdata/hierarchy/level4",
			},
			Triggers: []gitopsv1alpha1.GitOpsTrigger{
				{Type: "Change"},
			},
			ResourceDeletionMode:   "Delete",
			TemplateProcessorImage: "quay.io/kohlstechnology/eunomia-helm:dev",
			ResourceHandlingMode:   "Apply",
			ServiceAccountRef:      "eunomia-operator",
		},
	}
	gitops.Annotations = map[string]string{"gitopsconfig.eunomia.kohls.io/initialized": "true"}

	err = framework.Global.Client.Create(ctx, gitops, &framework.CleanupOptions{TestContext: ctx.TestCtx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	err = WaitForPodWithImage(t, framework.Global, ctx.namespace, "hello-world-hierarchy", "hello-app:1.0", retryInterval, timeout)
	if err != nil {
		t.Error(err)
	}
}
