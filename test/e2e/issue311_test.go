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

// The error that this tests for is the case where Deletion mode is set to "None" but changes need to me made.
// Prior to the fix, when the deletion mode was set to "None" nothing would get created.
func TestIssue311DeleteModeNone(t *testing.T) {
	ctx, err := NewContext(t)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Cleanup()

	// Create initial CR with "ResourceDeletionMode" mode set to "None", check that pods are started
	// prior to this fix, the pods would not get created.

	gitops := &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-modes",
			Namespace: ctx.namespace,
		},
		Spec: gitopsv1alpha1.GitOpsConfigSpec{
			TemplateSource: gitopsv1alpha1.GitConfig{
				URI:        ctx.eunomiaURI,
				Ref:        ctx.eunomiaRef,
				ContextDir: "test/e2e/testdata/modes/template1",
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
			ResourceHandlingMode:   "Create",
			ResourceDeletionMode:   "None",
			ServiceAccountRef:      "eunomia-operator",
		},
	}

	err = framework.Global.Client.Create(ctx, gitops, &framework.CleanupOptions{TestContext: ctx.TestCtx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	err = WaitForPodWithImage(t, framework.Global, ctx.namespace, "hello-world-modes", "hello-app:1.0", retryInterval, timeout)
	if err != nil {
		t.Error(err)
	}
}
