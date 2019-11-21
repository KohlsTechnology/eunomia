package e2e

import (
	goctx "context"
	"fmt"
	"os"
	"testing"

	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	test "github.com/KohlsTechnology/eunomia/test"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHierarchy(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	test.AddToFrameworkSchemeForTests(t, ctx)
	if err := hierarchyTestDeploy(t, framework.Global, ctx); err != nil {
		t.Fatal(err)
	}
}

func hierarchyTestDeploy(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	eunomiaURI, found := os.LookupEnv("EUNOMIA_URI")
	if !found {
		eunomiaURI = "https://github.com/kohlstechnology/eunomia"
	}

	eunomiaRef, found := os.LookupEnv("EUNOMIA_REF")
	if !found {
		eunomiaRef = "master"
	}

	gitops := &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-hierarchy",
			Namespace: namespace,
		},
		Spec: gitopsv1alpha1.GitOpsConfigSpec{
			TemplateSource: gitopsv1alpha1.GitConfig{
				URI:        eunomiaURI,
				Ref:        eunomiaRef,
				ContextDir: "test/e2e/configs/helm/templates",
			},
			ParameterSource: gitopsv1alpha1.GitConfig{
				URI:        eunomiaURI,
				Ref:        eunomiaRef,
				ContextDir: "test/e2e/configs/hierarchy/level4",
			},
			Triggers: []gitopsv1alpha1.GitOpsTrigger{
				{
					Type: "Change",
				},
			},
			ResourceDeletionMode:   "Delete",
			TemplateProcessorImage: "quay.io/kohlstechnology/eunomia-helm:dev",
			ResourceHandlingMode:   "CreateOrMerge",
			ServiceAccountRef:      "eunomia-operator",
		},
	}
	gitops.Annotations = map[string]string{"gitopsconfig.eunomia.kohls.io/initialized": "true"}

	err = f.Client.Create(goctx.TODO(), gitops, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		return err
	}

	return WaitForPodWithImage(t, f, namespace, "hello-world-hierarchy", "hello-app:1.0", retryInterval, timeout)
}
