package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"
	test "github.com/KohlsTechnology/eunomia/test"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestHelmTemplate(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	test.AddToFrameworkSchemeForTests(t, ctx)
	if err := helmTestDeploy(t, framework.Global, ctx); err != nil {
		t.Fatal(err)
	}
}

func helmTestDeploy(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	gitops := &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-helm",
			Namespace: namespace,
		},
		Spec: gitopsv1alpha1.GitOpsConfigSpec{
			TemplateSource: gitopsv1alpha1.GitConfig{
				URI:        "https://github.com/KohlsTechnology/eunomia/examples",
				Ref:        "master",
				ContextDir: "simple/helm",
			},
			ParameterSource: gitopsv1alpha1.GitConfig{
				URI:        "https://github.com/KohlsTechnology/eunomia/examples",
				Ref:        "master",
				ContextDir: "simple/helm",
			},
			Triggers: []gitopsv1alpha1.GitOpsTrigger{
				{
					Type: "Change",
				},
			},
			ResourceDeletionMode:   "Delete",
			TemplateProcessorImage: "quay.io/KohlsTechnology/eunomia-helm:v0.0.1",
			ResourceHandlingMode:   "CreateOrMerge",
			ServiceAccountRef: "gitops-operator",
		},
	}
	gitops.Annotations = map[string]string{"gitopsconfig.eunomia.kohls.io/initialized": "true"}

	err = f.Client.Create(goctx.TODO(), gitops, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	if err != nil {
		return err
	}

	return WaitForPodWithImage(t, f, ctx, namespace, "helloworld-helm","quay.io/KohlsTechnology/helloworld:v0.1.0", retryInterval, timeout)
}