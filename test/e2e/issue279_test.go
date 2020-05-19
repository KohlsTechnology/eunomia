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
	"time"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/KohlsTechnology/eunomia/pkg/apis"
	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	"github.com/KohlsTechnology/eunomia/pkg/util"
)

// TestIssue279CronJobDeletion verifies that after removing Periodic trigger from
// GitOpsConfig, CronJob gets also deleted
func TestIssue279CronJobDeletion(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	namespace, err := ctx.GetOperatorNamespace()
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

	// Step 1: create a CR with a Periodic trigger

	gitops := &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-issue279",
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
				{
					Type: "Periodic",
					Cron: "*/1 * * * *",
				},
			},
			TemplateProcessorImage: "quay.io/kohlstechnology/eunomia-base:dev",
			ResourceHandlingMode:   "Apply",
			ResourceDeletionMode:   "Delete",
			ServiceAccountRef:      "eunomia-operator",
		},
	}
	gitops.Annotations = map[string]string{"gitopsconfig.eunomia.kohls.io/initialized": "true"}

	err = framework.Global.Client.Create(context.TODO(), gitops, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	// Step 2: Wait until CronJob creates a Job, whose pod succeeds

	// Two minutes timeout to give cronjob enought time to kick off
	err = wait.Poll(retryInterval, time.Second*120, func() (done bool, err error) {
		const name = "gitopsconfig-gitops-issue279-"
		pod, err := GetPod(t, namespace, name, "quay.io/kohlstechnology/eunomia-base:dev", framework.Global.KubeClient)
		switch {
		case apierrors.IsNotFound(err):
			t.Logf("Waiting for availability of %s pod", name)
			return false, nil
		case err != nil:
			return false, err
		case pod != nil && pod.Status.Phase == "Succeeded":
			return true, nil
		case pod != nil:
			t.Logf("Waiting for pod %s to succeed", pod.Name)
			return false, nil
		default:
			t.Logf("Waiting for pod %s", name)
			return false, nil
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	// Step 3: Change Periodic trigger to Webhook one in GitOpsConfig, and apply it

	err = framework.Global.Client.Get(context.TODO(), util.NN{Name: "gitops-issue279", Namespace: namespace}, gitops)
	if err != nil {
		t.Fatal(err)
	}
	gitops.Spec.Triggers = []gitopsv1alpha1.GitOpsTrigger{{Type: "Webhook"}}
	err = framework.Global.Client.Update(context.TODO(), gitops)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("GitOpsConfig successfully updated")

	// Step 4: Wait for CronJob to be deleted

	err = wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		const name = "gitopsconfig-gitops-issue279-"
		cronJob, err := GetCronJob(namespace, name, framework.Global.KubeClient)
		switch {
		case apierrors.IsNotFound(err) || cronJob == nil:
			t.Logf("CronJob %s successfully deleted", name)
			return true, nil
		case err != nil:
			return false, err
		case cronJob != nil:
			t.Logf("Waiting for cronJob %s to be deleted", cronJob.Name)
			return false, nil
		default:
			t.Logf("Waiting for cronJob %s to be deleted", name)
			return false, nil
		}
	})
	if err != nil {
		t.Fatal(err)
	}

}
