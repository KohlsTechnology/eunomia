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
	"k8s.io/client-go/util/retry"

	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	"github.com/KohlsTechnology/eunomia/pkg/util"
)

func TestIssue24RemovedResourceGetsDeleted(t *testing.T) {
	ctx, err := NewContext(t)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Cleanup()

	// Step 1: create initial CR, check that pods are started

	gitops := &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-issue24",
			Namespace: ctx.namespace,
		},
		Spec: gitopsv1alpha1.GitOpsConfigSpec{
			TemplateSource: gitopsv1alpha1.GitConfig{
				URI:        ctx.eunomiaURI,
				Ref:        ctx.eunomiaRef,
				ContextDir: "test/e2e/testdata/issue24/template1",
			},
			ParameterSource: gitopsv1alpha1.GitConfig{
				URI:        ctx.eunomiaURI,
				Ref:        ctx.eunomiaRef,
				ContextDir: "test/e2e/testdata/empty-yaml",
			},
			Triggers: []gitopsv1alpha1.GitOpsTrigger{
				{Type: "Change"},
			},
			TemplateProcessorImage: "quay.io/kohlstechnology/eunomia-base:dev",
			ResourceHandlingMode:   "Apply",
			ResourceDeletionMode:   "Delete",
			ServiceAccountRef:      "eunomia-operator",
		},
	}
	gitops.Annotations = map[string]string{"gitopsconfig.eunomia.kohls.io/initialized": "true"}

	err = framework.Global.Client.Create(ctx, gitops, &framework.CleanupOptions{TestContext: ctx.TestCtx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	err = WaitForPodWithImage(t, framework.Global, ctx.namespace, "hello-world-issue24", "hello-app:1.0", retryInterval, timeout)
	if err != nil {
		t.Error(err)
	}
	err = WaitForPodWithImage(t, framework.Global, ctx.namespace, "now-only", "hello-app:2.0", retryInterval, timeout)
	if err != nil {
		t.Error(err)
	}

	// Step 2: change the CR to one with missing 'now-only' resource, then check that the pod gets deleted

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err := framework.Global.Client.Get(ctx, util.GetNN(gitops), gitops)
		if err != nil {
			t.Fatal(err)
		}
		gitops.Spec.TemplateSource.ContextDir = "test/e2e/testdata/issue24/template2"
		err = framework.Global.Client.Update(ctx, gitops)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the main "hello-world-issue24" app gets upgraded
	err = WaitForPodWithImage(t, framework.Global, ctx.namespace, "hello-world-issue24", "hello-app:2.0", retryInterval, timeout)
	if err != nil {
		t.Error(err)
	}
	// Verify that the pod corresponding to the missing resource gets deleted
	err = WaitForPodAbsence(t, framework.Global, ctx.namespace, "now-only", "hello-app:2.0", retryInterval, timeout)
	if err != nil {
		t.Error(err)
	}
}
