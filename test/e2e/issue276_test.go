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
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/KohlsTechnology/eunomia/pkg/apis"
	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
)

// TestIssue276NoTemplatesDir verifies that job's pod fails in such a
// way that it prints a custom error message to output when user passes
// nonexistent TemplateSource ContextDir. Then, the test also verifies
// that the Custom Recource is successfully deleted
func TestIssue276NoTemplatesDir(t *testing.T) {
	if testing.Short() {
		// FIXME: as of writing this test, "backoffLimit" in job.yaml is set to 4,
		// which means eunomia will wait to launch deletion Job until 5 Pod retries
		// fail, eventually triggering the origninal Job's failure; the back-off
		// time between the runs is unfortunately exponential and non-configurable,
		// which makes this test awfully long. Try to at least make it possible to
		// run in parallel with other tests.
		t.Skip("This test currently takes minutes to run, because of exponential backoff in kubernetes")
	}

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

	// Step 1: create a CR with a nonexistent TemplateSource ContextDir

	const noDir = "test/e2e/testdata/no-directory"
	gitops := &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-issue276",
			Namespace: namespace,
			Finalizers: []string{
				"gitopsconfig.eunomia.kohls.io/finalizer",
			},
		},
		Spec: gitopsv1alpha1.GitOpsConfigSpec{
			TemplateSource: gitopsv1alpha1.GitConfig{
				URI:        eunomiaURI,
				Ref:        eunomiaRef,
				ContextDir: noDir,
			},
			ParameterSource: gitopsv1alpha1.GitConfig{
				URI:        eunomiaURI,
				Ref:        eunomiaRef,
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

	err = framework.Global.Client.Create(context.TODO(), gitops, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	// Step 2: Wait until Job's pod fails and check if it printed a clear error message

	const name = "gitopsconfig-gitops-issue276-"
	err = wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		pod, err := GetPod(namespace, name, "quay.io/kohlstechnology/eunomia-base:dev", framework.Global.KubeClient)
		switch {
		case apierrors.IsNotFound(err):
			t.Logf("Waiting for availability of %s pod", name)
			return false, nil
		case err != nil:
			return false, err
		case pod != nil && pod.Status.Phase == "Failed":
			logs, err := GetPodLogs(pod, framework.Global.KubeClient)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(logs, fmt.Sprintf("ERROR - directory %s does not exist in the remote repository", noDir)) {
				t.Fatalf("Pod %s failed in an unexpected way; logs:\n%s", pod.Name, logs)
			}
			return true, nil
		case pod != nil:
			t.Logf("Waiting for error in pod %s; status: %s", pod.Name, debugJSON(pod.Status))
			return false, nil
		default:
			t.Logf("Waiting for error in pod %s", pod.Name)
			return false, nil
		}
	})
	if err != nil {
		t.Error(err)
	}

	// Step 3: Delete GitOpsConfig and make sure that the deletion succeeded

	t.Logf("Deleting CR")
	err = framework.Global.Client.Delete(context.TODO(), gitops)
	if err != nil {
		t.Fatal(err)
	}

	// Three minutes timeout.
	err = WaitForPodAbsence(t, framework.Global, namespace, name, "quay.io/kohlstechnology/eunomia-base:dev", retryInterval, 3*time.Minute)
	if err != nil {
		t.Error(err)
	}
}

// TestIssue276EmptyTemplatesDir verifies that job succeeds when user
// passes a TemplateSource directory containing no resource files, but
// only a single ".gitkeep" file
func TestIssue276EmptyTemplatesDir(t *testing.T) {
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

	// Step 1: create a CR with a TemplateSource ContextDir containing only a ".gitkeep" file

	gitops := &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-issue276",
			Namespace: namespace,
			Finalizers: []string{
				"gitopsconfig.eunomia.kohls.io/finalizer",
			},
		},
		Spec: gitopsv1alpha1.GitOpsConfigSpec{
			TemplateSource: gitopsv1alpha1.GitConfig{
				URI:        eunomiaURI,
				Ref:        eunomiaRef,
				ContextDir: "test/e2e/testdata/empty-directory",
			},
			ParameterSource: gitopsv1alpha1.GitConfig{
				URI:        eunomiaURI,
				Ref:        eunomiaRef,
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

	err = framework.Global.Client.Create(context.TODO(), gitops, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	// Step 2: Wait until Job's pod succeeds

	err = wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		const name = "gitopsconfig-gitops-issue276-"
		pod, err := GetPod(namespace, name, "quay.io/kohlstechnology/eunomia-base:dev", framework.Global.KubeClient)
		switch {
		case apierrors.IsNotFound(err):
			t.Logf("Waiting for availability of %s pod", name)
			return false, nil
		case err != nil:
			return false, err
		case pod != nil && pod.Status.Phase == "Succeeded":
			return true, nil
		case pod != nil:
			t.Logf("Waiting for pod %s to succeed; status: %s", pod.Name, debugJSON(pod.Status))
			return false, nil
		default:
			t.Logf("Waiting for pod %s", name)
			return false, nil
		}
	})
	if err != nil {
		t.Error(err)
	}
}
