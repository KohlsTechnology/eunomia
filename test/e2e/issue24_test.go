package e2e

import (
	"context"
	"os"
	"testing"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/KohlsTechnology/eunomia/pkg/apis"
	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
)

func TestIssue24_RemovedResourceGetsDeleted(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("could not get namespace: %v", err)
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

	// Step 1: create initial CR, check that pods are started

	gitops := &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-issue24",
			Namespace: namespace,
		},
		Spec: gitopsv1alpha1.GitOpsConfigSpec{
			TemplateSource: gitopsv1alpha1.GitConfig{
				URI:        eunomiaURI,
				Ref:        eunomiaRef,
				ContextDir: "test/e2e/testdata/issue24/template1",
			},
			ParameterSource: gitopsv1alpha1.GitConfig{
				URI:        eunomiaURI,
				Ref:        eunomiaRef,
				ContextDir: "test/e2e/testdata/issue24/parameters",
			},
			Triggers: []gitopsv1alpha1.GitOpsTrigger{
				{Type: "Change"},
			},
			TemplateProcessorImage: "quay.io/kohlstechnology/eunomia-base:dev",
			ResourceHandlingMode:   "CreateOrMerge",
			ResourceDeletionMode:   "Delete",
			ServiceAccountRef:      "eunomia-operator",
		},
	}
	gitops.Annotations = map[string]string{"gitopsconfig.eunomia.kohls.io/initialized": "true"}

	err = framework.Global.Client.Create(context.TODO(), gitops, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	err = WaitForPodWithImage(t, framework.Global, namespace, "hello-world-issue24", "hello-app:1.0", retryInterval, timeout)
	if err != nil {
		t.Error(err)
	}
	err = WaitForPodWithImage(t, framework.Global, namespace, "now-only", "hello-app:2.0", retryInterval, timeout)
	if err != nil {
		t.Error(err)
	}

	// Step 2: change the CR to one with missing 'now-only' resource, then check that the pod gets deleted

	err = framework.Global.Client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: gitops.ObjectMeta.Name}, gitops)
	if err != nil {
		t.Fatal(err)
	}
	gitops.Spec.TemplateSource.ContextDir = "test/e2e/testdata/issue24/template2"
	err = framework.Global.Client.Update(context.Background(), gitops)
	if err != nil {
		t.Fatal(err)
	}
	// Verify that the main "hello-world-issue24" app gets upgraded
	err = WaitForPodWithImage(t, framework.Global, namespace, "hello-world-issue24", "hello-app:2.0", retryInterval, timeout)
	if err != nil {
		t.Error(err)
	}
	// Verify that the pod corresponding to the missing resource gets deleted
	err = WaitForPodAbsence(t, framework.Global, namespace, "now-only", "hello-app:2.0", retryInterval, timeout)
	if err != nil {
		t.Error(err)
	}
}
