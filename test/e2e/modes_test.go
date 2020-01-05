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

func TestModes_CreateReplaceDelete(t *testing.T) {
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

	// Step 1: create initial CR with "Create" mode, check that pods are started

	gitops := &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-modes",
			Namespace: namespace,
		},
		Spec: gitopsv1alpha1.GitOpsConfigSpec{
			TemplateSource: gitopsv1alpha1.GitConfig{
				URI:        eunomiaURI,
				Ref:        eunomiaRef,
				ContextDir: "test/e2e/testdata/modes/template1",
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
			ResourceHandlingMode:   "Create",
			ResourceDeletionMode:   "Delete",
			ServiceAccountRef:      "eunomia-operator",
		},
	}
	gitops.Annotations = map[string]string{"gitopsconfig.eunomia.kohls.io/initialized": "true"}

	err = framework.Global.Client.Create(context.TODO(), gitops, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatal(err)
	}

	err = WaitForPodWithImage(t, framework.Global, namespace, "hello-world-modes", "hello-app:1.0", retryInterval, timeout)
	if err != nil {
		t.Error(err)
	}

	// Step 2: change the CR to a different version of image, using "Replace" mode, then verify pod change

	err = framework.Global.Client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: gitops.ObjectMeta.Name}, gitops)
	if err != nil {
		t.Fatal(err)
	}
	gitops.Spec.TemplateSource.ContextDir = "test/e2e/testdata/modes/template2"
	gitops.Spec.ResourceHandlingMode = "Replace"
	err = framework.Global.Client.Update(context.Background(), gitops)
	if err != nil {
		t.Fatal(err)
	}
	// Verify that the main "hello-world-modes" app gets upgraded
	err = WaitForPodWithImage(t, framework.Global, namespace, "hello-world-modes", "hello-app:2.0", retryInterval, timeout)
	if err != nil {
		t.Error(err)
	}

	// Step 2: change the CR to "Delete" mode, then verify that the Pod is deleted

	err = framework.Global.Client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: gitops.ObjectMeta.Name}, gitops)
	if err != nil {
		t.Fatal(err)
	}
	gitops.Spec.ResourceHandlingMode = "Delete"
	err = framework.Global.Client.Update(context.Background(), gitops)
	if err != nil {
		t.Fatal(err)
	}
	// Verify that the pod corresponding to the missing resource gets deleted
	err = WaitForPodAbsence(t, framework.Global, namespace, "hello-world-modes", "hello-app:2.0", retryInterval, timeout)
	if err != nil {
		t.Error(err)
	}
}
