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
	"context"
	"os"
	"testing"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/KohlsTechnology/eunomia/pkg/apis"
	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	"github.com/KohlsTechnology/eunomia/pkg/util"
)

func TestIssue216InvalidImageDeleted(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("could not get namespace: %v", err)
	}
	if err = SetupRbacInNamespace(namespace); err != nil {
		t.Error(err)
	}

	defer DumpJobsLogsOnError(t, framework.Global, namespace)
	err = framework.AddToFrameworkScheme(apis.AddToScheme, &gitopsv1alpha1.GitOpsConfigList{})
	if err != nil {
		t.Fatal(err)
	}

	eunomiaURI, found := os.LookupEnv("EUNOMIA_URI")
	if !found {
		eunomiaURI = "https://github.com/kohlstechnology/eunomia"
	}
	eunomiaRef, found := os.LookupEnv("EUNOMIA_REF")
	if !found {
		eunomiaRef = "master"
	}

	// Step 1: create a simple CR with an invalid template-processor URL

	gitops := &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-issue216",
			Namespace: namespace,
			Finalizers: []string{
				"gitopsconfig.eunomia.kohls.io/finalizer",
			},
		},
		Spec: gitopsv1alpha1.GitOpsConfigSpec{
			TemplateSource: gitopsv1alpha1.GitConfig{
				URI:        eunomiaURI,
				Ref:        eunomiaRef,
				ContextDir: "test/e2e/testdata/events/test-a",
			},
			ParameterSource: gitopsv1alpha1.GitConfig{
				URI:        eunomiaURI,
				Ref:        eunomiaRef,
				ContextDir: "test/e2e/testdata/empty-yaml",
			},
			Triggers: []gitopsv1alpha1.GitOpsTrigger{
				{Type: "Change"},
			},
			TemplateProcessorImage: "quay.io/kohlstechnology/invalid:bad",
			ResourceHandlingMode:   "Apply",
			ResourceDeletionMode:   "Delete",
			ServiceAccountRef:      "eunomia-operator",
		},
	}

	err = framework.Global.Client.Create(context.TODO(), gitops, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	// Step 2: Wait until Job exists (in incomplete state)

	err = wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		const name = "gitopsconfig-gitops-issue216-"
		pod, err := GetPod(namespace, name, "quay.io/kohlstechnology/invalid:bad", framework.Global.KubeClient)
		switch {
		case apierrors.IsNotFound(err):
			t.Logf("Waiting for availability of %s pod", name)
			return false, nil
		case err != nil:
			return false, err
		case pod != nil && pod.Status.Phase == "Pending":
			return true, nil
		case pod != nil:
			t.Logf("Waiting for error in pod %s; status: %s", name, debugJSON(pod.Status))
			return false, nil
		default:
			t.Logf("Waiting for error in pod %s", name)
			return false, nil
		}
	})
	if err != nil {
		t.Error(err)
	}

	// Step 3: Delete CR

	t.Logf("Deleting CR")
	err = framework.Global.Client.Delete(context.TODO(), gitops)
	if err != nil {
		t.Fatal(err)
	}

	// Step 4: Wait to verify that CR got successfully removed

	err = wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		found := gitopsv1alpha1.GitOpsConfig{}
		err = framework.Global.Client.Get(context.TODO(), util.GetNN(gitops), &found)
		switch {
		case apierrors.IsNotFound(err):
			t.Logf("Confirmed GitOpsConfig shutdown")
			return true, nil
		case err != nil:
			return false, err
		default:
			t.Logf("Waiting for shutdown of GitOpsConfig; status: %s", debugJSON(found.Status))
			return false, nil
		}
	})
	if err != nil {
		t.Error(err)
	}

	// Step 5: Wait until no Pods exist for this CR

	err = wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		const name = "gitopsconfig-gitops-issue216-"
		pod, err := GetPod(namespace, name, "", framework.Global.KubeClient)
		switch {
		case apierrors.IsNotFound(err):
			t.Logf("Confirmed no more pods found")
			return true, nil
		case err != nil:
			return false, err
		case pod == nil:
			t.Logf("Confirmed no more %s pods found", name)
			return true, nil
		default:
			t.Logf("Waiting for shutdown of %s pod; status: %s", name, debugJSON(pod.Status))
			return false, nil
		}
	})
	if err != nil {
		t.Error(err)
	}
}
