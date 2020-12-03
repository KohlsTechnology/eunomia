// +build e2e

package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"

	framework "github.com/operator-framework/operator-sdk/pkg/test"

	"github.com/KohlsTechnology/eunomia/pkg/apis"
	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
)

// Context contains various data commonly used in e2e tests. The struct also
// embeds and implements context.Context interface.
type Context struct {
	context.Context
	*framework.TestCtx
	namespace              string
	eunomiaURI, eunomiaRef string
}

// NewContext creates a new Context object with all fields filled with values
// appropriate for the current test function. It is expected that the following
// preamble will be used at the beginning of typical e2e test functions:
//
//	ctx, err := NewContext(t)
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer ctx.Cleanup()
func NewContext(t *testing.T) (*Context, error) {
	testCtx := framework.NewTestCtx(t)
	ns, err := testCtx.GetOperatorNamespace()
	if err != nil {
		return nil, fmt.Errorf("e2e new test context: getting namespace: %w", err)
	}
	err = SetupRbacInNamespace(ns)
	if err != nil {
		return nil, fmt.Errorf("e2e new test context: setting up RBAC: %w", err)
	}

	err = framework.AddToFrameworkScheme(apis.AddToScheme, &gitopsv1alpha1.GitOpsConfigList{})
	if err != nil {
		return nil, fmt.Errorf("e2e new test context: setting up GitOpsConfig scheme: %w", err)
	}

	eunomiaURI, found := os.LookupEnv("EUNOMIA_URI")
	if !found {
		eunomiaURI = "https://github.com/kohlstechnology/eunomia"
	}
	eunomiaRef, found := os.LookupEnv("EUNOMIA_REF")
	if !found {
		eunomiaRef = "master"
	}

	testCtx.AddCleanupFn(func() error {
		DumpJobsLogsOnError(t, framework.Global, ns)
		return nil
	})

	return &Context{
		Context:    context.Background(),
		TestCtx:    testCtx,
		namespace:  ns,
		eunomiaURI: eunomiaURI,
		eunomiaRef: eunomiaRef,
	}, nil
}
