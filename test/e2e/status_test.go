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
	"time"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	"github.com/KohlsTechnology/eunomia/pkg/util"
)

func TestStatusSuccess(t *testing.T) {
	ctx, err := NewContext(t)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Cleanup()

	// Step 1: create a simple CR with a single Pod

	gitops := &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-status-hello-success",
			Namespace: ctx.namespace,
		},
		Spec: gitopsv1alpha1.GitOpsConfigSpec{
			TemplateSource: gitopsv1alpha1.GitConfig{
				URI:        ctx.eunomiaURI,
				Ref:        ctx.eunomiaRef,
				ContextDir: "test/e2e/testdata/hello-a",
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

	start := time.Now().Truncate(time.Second) // Note: kubernetes returns times with only 1s precision, so truncate for comparisons
	err = framework.Global.Client.Create(ctx, gitops, &framework.CleanupOptions{TestContext: ctx.TestCtx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	// Step 2: watch Status till Success & verify Status fields

	err = wait.Poll(retryInterval, 25*time.Second, func() (done bool, err error) {
		fresh := gitopsv1alpha1.GitOpsConfig{}
		err = framework.Global.Client.Get(ctx, util.GetNN(gitops), &fresh)
		if err != nil {
			return false, err
		}
		switch fresh.Status.State {
		case "InProgress":
			now, s := time.Now(), fresh.Status
			if s.CompletionTime != nil {
				t.Errorf("want CompletionTime==nil, got: %#v", s)
			}
			if s.StartTime == nil || s.StartTime.Time.Before(start) || now.Before(s.StartTime.Time) {
				t.Errorf("want %v <= StartTime <= %v, got: %v", start, now, debugJSON(s))
			}
			return false, nil
		case "Success":
			now, s := time.Now(), fresh.Status
			if s.StartTime == nil || s.CompletionTime == nil ||
				s.StartTime.Time.Before(start) ||
				s.CompletionTime.Before(s.StartTime) ||
				now.Before(s.CompletionTime.Time) {
				t.Errorf("want %v <= StartTime <= CompletionTime <= %v, got: %v", start, now, debugJSON(s))
			}
			return true, nil
		default:
			t.Errorf("Unexpected State: %v", debugJSON(fresh.Status))
			return false, nil
		}
	})
	if err != nil {
		t.Error(err)
	}

	// Step 3: verify that the pod exists

	pod, err := GetPod(ctx.namespace, "hello-test-a", "hello-app:1.0", framework.Global.KubeClient)
	if err != nil {
		t.Fatal(err)
	}
	if pod == nil || pod.Status.Phase != "Running" {
		t.Fatalf("unexpected state of Pod: %v", pod)
	}
}

func TestStatusPeriodicJobSuccess(t *testing.T) {
	ctx, err := NewContext(t)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Cleanup()

	// Step 1: create a simple CR with a single Pod

	gitops := &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-status-periodic-success",
			Namespace: ctx.namespace,
		},
		Spec: gitopsv1alpha1.GitOpsConfigSpec{
			TemplateSource: gitopsv1alpha1.GitConfig{
				URI:        ctx.eunomiaURI,
				Ref:        ctx.eunomiaRef,
				ContextDir: "test/e2e/testdata/hello-b",
			},
			ParameterSource: gitopsv1alpha1.GitConfig{
				URI:        ctx.eunomiaURI,
				Ref:        ctx.eunomiaRef,
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

	start := time.Now().Truncate(time.Second) // Note: kubernetes returns times with only 1s precision, so truncate for comparisons
	err = framework.Global.Client.Create(ctx, gitops, &framework.CleanupOptions{TestContext: ctx.TestCtx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	// Step 2: skip initial empty Status, before any Job was started

	err = wait.Poll(retryInterval, 2*time.Minute, func() (done bool, err error) {
		fresh := gitopsv1alpha1.GitOpsConfig{}
		err = framework.Global.Client.Get(ctx, util.GetNN(gitops), &fresh)
		if err != nil {
			return false, err
		}
		return fresh.Status.State != "", nil
	})
	if err != nil {
		t.Error(err)
	}

	// Step 3: watch Status till Success & verify Status fields

	err = wait.Poll(retryInterval, 2*time.Minute, func() (done bool, err error) {
		fresh := gitopsv1alpha1.GitOpsConfig{}
		err = framework.Global.Client.Get(ctx, util.GetNN(gitops), &fresh)
		if err != nil {
			return false, err
		}
		switch fresh.Status.State {
		case "InProgress":
			now, s := time.Now(), fresh.Status
			if s.CompletionTime != nil {
				t.Errorf("want CompletionTime==nil, got: %#v", s)
			}
			if s.StartTime == nil || s.StartTime.Time.Before(start) || now.Before(s.StartTime.Time) {
				t.Errorf("want %v <= StartTime <= %v, got: %v", start, now, debugJSON(s))
			}
			return false, nil
		case "Success":
			now, s := time.Now(), fresh.Status
			if s.StartTime == nil || s.CompletionTime == nil ||
				s.StartTime.Time.Before(start) ||
				s.CompletionTime.Before(s.StartTime) ||
				now.Before(s.CompletionTime.Time) {
				t.Errorf("want %v <= StartTime <= CompletionTime <= %v, got: %v", start, now, debugJSON(s))
			}
			return true, nil
		default:
			t.Errorf("Unexpected State: %v", debugJSON(fresh.Status))
			return false, nil
		}
	})
	if err != nil {
		t.Error(err)
	}

	// Step 3: verify that the pod exists

	pod, err := GetPod(ctx.namespace, "hello-test-b", "hello-app:1.0", framework.Global.KubeClient)
	if err != nil {
		t.Fatal(err)
	}
	if pod == nil || pod.Status.Phase != "Running" {
		t.Fatalf("unexpected state of Pod: %v", pod)
	}
}

func TestStatusFailure(t *testing.T) {
	if testing.Short() {
		// FIXME: as of writing this test, "backoffLimit" in job.yaml is set to 4,
		// which means we need to wait until 5 Pod retries fail, eventually
		// triggering a Job failure; the back-off time between the runs is
		// unfortunately exponential and non-configurable, which makes this test
		// awfully long. Try to at least make it possible to run in parallel with
		// other tests.
		t.Skip("This test currently takes minutes to run, because of exponential backoff in kubernetes")
	}

	ctx, err := NewContext(t)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Cleanup()

	// Step 1: create a CR with an invalid URI

	gitops := &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-status-hello-failed",
			Namespace: ctx.namespace,
		},
		Spec: gitopsv1alpha1.GitOpsConfigSpec{
			TemplateSource: gitopsv1alpha1.GitConfig{
				URI:        "https://INVALID!!!",
				Ref:        ctx.eunomiaRef,
				ContextDir: "URI is already invalid so this value should be irrelevant",
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

	start := time.Now().Truncate(time.Second) // Note: kubernetes returns times with only 1s precision, so truncate for comparisons
	err = framework.Global.Client.Create(ctx, gitops, &framework.CleanupOptions{TestContext: ctx.TestCtx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	// Step 2: watch Status till Failure & verify Status fields

	err = wait.Poll(retryInterval, 3*time.Minute, func() (done bool, err error) {
		fresh := gitopsv1alpha1.GitOpsConfig{}
		err = framework.Global.Client.Get(ctx, util.GetNN(gitops), &fresh)
		if err != nil {
			return false, err
		}
		switch fresh.Status.State {
		case "InProgress":
			now, s := time.Now(), fresh.Status
			if s.CompletionTime != nil {
				t.Errorf("want CompletionTime==nil, got: %#v", s)
			}
			if s.StartTime == nil || s.StartTime.Time.Before(start) || now.Before(s.StartTime.Time) {
				t.Errorf("want %v <= StartTime <= %v, got: %v", start, now, debugJSON(s))
			}
			return false, nil
		case "Failure":
			now, s := time.Now(), fresh.Status
			if s.CompletionTime != nil {
				t.Errorf("want CompletionTime==nil, got: %v", debugJSON(s))
			}
			if s.StartTime == nil || s.StartTime.Time.Before(start) || now.Before(s.StartTime.Time) {
				t.Errorf("want %v <= StartTime <= %v, got: %v", start, now, debugJSON(s))
			}
			return true, nil
		default:
			t.Errorf("Unexpected State: %#v", debugJSON(fresh.Status))
			return false, nil
		}
	})
	if err != nil {
		t.Error(err)
	}
}
