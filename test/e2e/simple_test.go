/*
Copyright 2019 Kohl's Department Stores, Inc.

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
	goctx "context"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/stretchr/testify/assert"
	"gitops-operator/pkg/apis/eunomia/v1alpha1"
	gitopsv1alpha1 "gitops-operator/pkg/apis/eunomia/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestSimple(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	simpleTestDeploy(t, framework.Global, ctx)
}

func simpleTestDeploy(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {
	namespace, err := ctx.GetNamespace()
	assert.NoError(t, err)

	// Check if the CRD has been created
	crd := &gitopsv1alpha1.GitOpsConfig{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "gitops", Namespace: namespace}, crd)
	assert.Error(t, err)

	gitops := &v1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops",
			Namespace: namespace,
		},
		Spec: v1alpha1.GitOpsConfigSpec{},
	}

	err = f.Client.Create(goctx.TODO(), gitops, &framework.CleanupOptions{TestContext: ctx, Timeout: timeout, RetryInterval: retryInterval})
	assert.NoError(t, err)

	// Check if the CRD has been created
	crd = &gitopsv1alpha1.GitOpsConfig{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "gitops", Namespace: namespace}, crd)
	assert.NoError(t, err)
}
