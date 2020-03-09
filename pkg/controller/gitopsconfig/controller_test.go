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

package gitopsconfig

import (
	"context"
	"testing"

	"golang.org/x/xerrors"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	"github.com/KohlsTechnology/eunomia/pkg/util"
)

const (
	namespace = "gitops"
)

func defaultGitOpsConfig() *gitopsv1alpha1.GitOpsConfig {
	return &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "gitops-operator",
			Namespace:   namespace,
			Finalizers:  []string{tagFinalizer},
			Annotations: map[string]string{tagInitialized: "true"},
		},
		Spec: gitopsv1alpha1.GitOpsConfigSpec{
			TemplateSource: gitopsv1alpha1.GitConfig{
				URI:        "https://github.com/KohlsTechnology/eunomia",
				Ref:        "master",
				HTTPProxy:  "http://proxy.com:8080",
				HTTPSProxy: "http://proxy.com:8080",
				NOProxy:    "mygit.com",
				ContextDir: "test/deploy",
				SecretRef:  "pio",
			},
			ParameterSource: gitopsv1alpha1.GitConfig{
				URI:        "https://github.com/URI1/URI2",
				Ref:        "master",
				HTTPProxy:  "http://proxy.com:8080",
				HTTPSProxy: "http://proxy.com:8080",
				NOProxy:    "mygit.com",
				ContextDir: "ciaoContext",
				SecretRef:  "pio",
			},
			Triggers: []gitopsv1alpha1.GitOpsTrigger{
				{
					Type: "Periodic",
					Cron: "0 * * * *",
				},
			},
			ServiceAccountRef:      "mysvcaccount",
			ResourceDeletionMode:   "Delete",
			TemplateProcessorImage: "myimage",
			ResourceHandlingMode:   "Apply",
		},
	}
}

func defaultNamespace() *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespace,
			Namespace: namespace,
		},
	}
}

func TestCRDInitialization(t *testing.T) {
	gitops := defaultGitOpsConfig()

	// Initialize fake client with objects it should track
	cl := fake.NewFakeClient(gitops)
	r := &Reconciler{client: cl, scheme: scheme.Scheme}

	r.Reconcile(reconcile.Request{
		NamespacedName: util.GetNN(gitops),
	})

	// Check if the CRD has been created
	crd := &gitopsv1alpha1.GitOpsConfig{}
	err := cl.Get(context.Background(), util.GetNN(gitops), crd)
	if err != nil {
		t.Fatal(err)
	}

	// Check if the name matches what was deployed
	if crd.Name != gitops.Name {
		t.Errorf("expected name %q, got %q", gitops.Name, crd.Name)
	}
}

func TestPeriodicTrigger(t *testing.T) {
	gitops := defaultGitOpsConfig()

	// Initialize fake client with objects it should track
	cl := fake.NewFakeClient(gitops)
	r := &Reconciler{client: cl, scheme: scheme.Scheme}

	r.Reconcile(reconcile.Request{
		NamespacedName: util.GetNN(gitops),
	})

	// Check if the CronJob has been created
	cron := &batchv1beta1.CronJob{}
	err := cl.Get(context.Background(), util.NN{Name: "gitopsconfig-gitops-operator", Namespace: namespace}, cron)
	if err != nil {
		t.Fatal(err)
	}

	// Check if the name matches what was deployed
	wantName := "gitopsconfig-gitops-operator"
	if cron.Name != wantName {
		t.Errorf("expected name %q, got %q", wantName, cron.Name)
	}
}

func TestChangeTrigger(t *testing.T) {
	gitops := defaultGitOpsConfig()
	// Set trigger type to Change
	gitops.Spec.Triggers = []gitopsv1alpha1.GitOpsTrigger{
		{
			Type: "Change",
		},
	}
	// Initialize fake client with objects it should track
	cl := fake.NewFakeClient(gitops)
	r := &Reconciler{client: cl, scheme: scheme.Scheme}

	r.Reconcile(reconcile.Request{
		NamespacedName: util.GetNN(gitops),
	})

	// Check if the CRD has been created
	job := &batchv1.Job{}
	err := cl.Get(context.Background(), util.NN{Namespace: namespace}, job)
	if err != nil {
		t.Fatal(err)
	}

	// Check if the name matches what was deployed
	wantKind := "Job"
	if job.Kind != wantKind {
		t.Errorf("expected Kind %q, got %q", wantKind, job.Kind)
	}
}

func TestWebhookTrigger(t *testing.T) {
	gitops := defaultGitOpsConfig()
	// Set trigger type to Webhook
	gitops.Spec.Triggers = []gitopsv1alpha1.GitOpsTrigger{
		{
			Type: "Webhook",
		},
	}
	// Initialize fake client with objects it should track
	cl := fake.NewFakeClient(gitops)
	r := &Reconciler{client: cl, scheme: scheme.Scheme}

	r.Reconcile(reconcile.Request{
		NamespacedName: util.GetNN(gitops),
	})

	// Check if the Job has been created
	job := &batchv1.Job{}
	err := cl.Get(context.Background(), util.NN{Namespace: namespace}, job)
	if err != nil {
		t.Fatal(err)
	}

	// Check if the name matches what was deployed
	wantKind := "Job"
	if job.Kind != wantKind {
		t.Errorf("expected Kind %q, got %q", wantKind, job.Kind)
	}
}

func TestDeleteRemovingFinalizer(t *testing.T) {
	gitops := defaultGitOpsConfig()
	gitops.Spec.Triggers = []gitopsv1alpha1.GitOpsTrigger{
		{
			Type: "Change",
		},
	}

	// Initialize fake client with objects it should track
	cl := fake.NewFakeClient(gitops)
	r := &Reconciler{client: cl, scheme: scheme.Scheme}

	// Create a namespace
	err := cl.Create(context.Background(), defaultNamespace())
	if err != nil {
		t.Fatal(err)
	}

	r.Reconcile(reconcile.Request{
		NamespacedName: util.GetNN(gitops),
	})

	// Get the CRD so that we can add the deletion timestamp
	crd := &gitopsv1alpha1.GitOpsConfig{}
	err = cl.Get(context.Background(), util.GetNN(gitops), crd)
	if err != nil {
		t.Fatal(err)
	}
	// Make sure the finalizer has been added
	if len(crd.ObjectMeta.Finalizers) == 0 {
		t.Fatal("finalizer was not added")
	}

	// Set deletion timestamp
	deleteTime := metav1.Now()
	crd.ObjectMeta.DeletionTimestamp = &deleteTime
	// Update the CRD with the new deletion timestamp
	err = cl.Update(context.Background(), crd)
	if err != nil {
		t.Fatal(err)
	}
	// Create the deleteJob
	var (
		dummyInt32 int32 = 1
		dummyBool  bool  = true
	)
	err = cl.Create(context.Background(), &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-operator-delete",
			Namespace: namespace,
			Labels:    map[string]string{"action": "delete"},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "eunomia.kohls.io/v1alpha1",
					Kind:               "GitOpsConfig",
					Name:               gitops.Name,
					Controller:         &dummyBool,
					BlockOwnerDeletion: &dummyBool,
				},
			},
		},
		Spec: batchv1.JobSpec{
			Parallelism:  &dummyInt32,
			Completions:  &dummyInt32,
			BackoffLimit: &dummyInt32,
		},
		Status: batchv1.JobStatus{
			Succeeded: 2,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Reconcile so that the controller can delete the finalizer
	r.Reconcile(reconcile.Request{
		NamespacedName: util.GetNN(gitops),
	})

	// Check the status
	crd = &gitopsv1alpha1.GitOpsConfig{}
	err = cl.Get(context.Background(), util.NN{Namespace: namespace}, crd)
	if err != nil {
		t.Fatal(err)
	}
	// The finalizer should have been removed
	if len(crd.ObjectMeta.Finalizers) != 0 {
		t.Errorf("expected empty finalizers, got: %v", crd.ObjectMeta.Finalizers)
	}
}

func TestCreatingDeleteJob(t *testing.T) {
	gitops := defaultGitOpsConfig()
	gitops.Spec.Triggers = []gitopsv1alpha1.GitOpsTrigger{
		{
			Type: "Change",
		},
	}

	// Initialize fake client with objects it should track
	cl := fake.NewFakeClient(gitops)
	r := &Reconciler{client: cl, scheme: scheme.Scheme}

	// Create a namespace
	err := cl.Create(context.Background(), defaultNamespace())
	if err != nil {
		t.Fatal(err)
	}

	r.Reconcile(reconcile.Request{
		NamespacedName: util.GetNN(gitops),
	})

	// Get the CRD so that we can add the deletion timestamp
	crd := &gitopsv1alpha1.GitOpsConfig{}
	err = cl.Get(context.Background(), util.GetNN(gitops), crd)
	if err != nil {
		t.Fatal(err)
	}
	// Make sure the finalizer has been added
	if len(crd.ObjectMeta.Finalizers) == 0 {
		t.Fatal("finalizer was not added")
	}

	// Make sure there's no delete job
	job, err := findDeleteJob(cl)
	if err != nil {
		t.Fatal(err)
	}
	if job.GetLabels()["action"] == "delete" {
		t.Fatalf("found unexpected delete job: %q", job.Name)
	}

	// Set deletion timestamp
	deleteTime := metav1.Now()
	crd.ObjectMeta.DeletionTimestamp = &deleteTime
	// Update the CRD with the new deletion timestamp
	err = cl.Update(context.Background(), crd)
	if err != nil {
		t.Error(err)
	}

	// Fakeclient is not updating the job status , inorder to create the new job we are
	// Updating the job status manually for the existing job created by Reconcile.
	job, err = findRunningJob(cl)
	if err != nil {
		t.Fatal(err)
	}
	job.Status.Active = 0
	job.Status.Succeeded = 1
	job.Status.Failed = 0
	job.Status.StartTime = &deleteTime
	// Update the job with the status
	err = cl.Update(context.Background(), &job)
	if err != nil {
		t.Fatal(err)
	}

	// There shouldn't be a delete job at this point, the reconciler should create one
	r.Reconcile(reconcile.Request{
		NamespacedName: util.GetNN(gitops),
	})

	// See if a delete job was created
	job, err = findDeleteJob(cl)
	if err != nil {
		t.Fatal(err)
	}
	if job.GetLabels()["action"] != "delete" {
		t.Error("delete job not found")
	}
}

func TestDeleteWhileNamespaceDeleting(t *testing.T) {
	gitops := defaultGitOpsConfig()
	gitops.Spec.Triggers = []gitopsv1alpha1.GitOpsTrigger{
		{
			Type: "Change",
		},
	}

	// Initialize fake client
	cl := fake.NewFakeClient(gitops)
	r := &Reconciler{client: cl, scheme: scheme.Scheme}

	// Create a namespace
	// Set deletion timestamp on the namespace
	deleteTime := metav1.Now()
	ns0 := defaultNamespace()
	ns0.ObjectMeta.DeletionTimestamp = &deleteTime
	err := cl.Create(context.Background(), ns0)
	if err != nil {
		t.Fatal(err)
	}

	r.Reconcile(reconcile.Request{
		NamespacedName: util.GetNN(gitops),
	})

	// Get the CRD so that we can add the deletion timestamp
	crd := &gitopsv1alpha1.GitOpsConfig{}
	err = cl.Get(context.Background(), util.GetNN(gitops), crd)
	if err != nil {
		t.Fatal(err)
	}
	// Make sure the finalizer has been added
	if len(crd.ObjectMeta.Finalizers) == 0 {
		t.Fatal("finalizer was not added")
	}

	// Set deletion timestamp
	deleteTime = metav1.Now()
	crd.ObjectMeta.DeletionTimestamp = &deleteTime
	// Update the CRD with the new deletion timestamp
	err = cl.Update(context.Background(), crd)
	if err != nil {
		t.Fatal(err)
	}

	// There shouldn't be a delete job at this point, the reconciler should create one
	r.Reconcile(reconcile.Request{
		NamespacedName: util.GetNN(gitops),
	})

	// Check the status
	crd = &gitopsv1alpha1.GitOpsConfig{}
	err = cl.Get(context.Background(), util.NN{Namespace: namespace}, crd)
	if err != nil {
		t.Fatal(err)
	}
	// The finalizer should have been removed
	if len(crd.ObjectMeta.Finalizers) != 0 {
		t.Errorf("expected empty finalizers, got: %v", crd.ObjectMeta.Finalizers)
	}
}

func findDeleteJob(cl client.Client) (batchv1.Job, error) {
	jobs, err := findJobList(cl)
	if err != nil {
		return batchv1.Job{}, xerrors.Errorf("unable to find delete job: %w", err)
	}
	// Return the first instance that is a delete job
	for _, job := range jobs {
		if job.GetLabels()["action"] == "delete" {
			return job, nil
		}
	}
	return batchv1.Job{}, nil
}

func TestCreateJob(t *testing.T) {
	gitops := defaultGitOpsConfig()
	// Set trigger type to Change
	gitops.Spec.Triggers = []gitopsv1alpha1.GitOpsTrigger{
		{
			Type: "Change",
		},
	}
	// Initialize fake client with objects it should track
	cl := fake.NewFakeClient(gitops)
	r := &Reconciler{client: cl, scheme: scheme.Scheme}

	r.Reconcile(reconcile.Request{
		NamespacedName: util.GetNN(gitops),
	})

	// Fakeclient is not updating the job status , inorder to test race condition between the jobs we are
	// Updating the job status manually for the existing job created by Reconcile.
	startTime := metav1.Now()
	job, err := findRunningJob(cl)
	if err != nil {
		t.Fatal(err)
	}
	job.Status.Active = 1
	job.Status.Succeeded = 0
	job.Status.Failed = 0
	job.Status.StartTime = &startTime

	err = cl.Update(context.Background(), &job)
	if err != nil {
		t.Fatal(err)
	}

	r.Reconcile(reconcile.Request{
		NamespacedName: util.GetNN(gitops),
	})

	jobs, err := findJobList(cl)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) > 1 {
		t.Error("Job was not postponed")
	}
}

func findJobList(cl client.Client) ([]batchv1.Job, error) {
	// Looking up all jobs
	jobs := batchv1.JobList{}
	err := cl.List(context.Background(), &client.ListOptions{
		Namespace: namespace,
	}, &jobs)
	if err != nil {
		return nil, xerrors.Errorf("unable to list the running jobs: %w", err)
	}
	return jobs.Items, nil
}

func findRunningJob(cl client.Client) (batchv1.Job, error) {
	jobs, err := findJobList(cl)
	if err != nil {
		return batchv1.Job{}, xerrors.Errorf("unable to find running job: %w", err)
	}
	if len(jobs) >= 2 {
		names := []string{}
		for _, j := range jobs {
			names = append(names, j.Name)
		}
		return batchv1.Job{}, xerrors.Errorf("found more than 1 job: %v", names)
	}
	if len(jobs) > 0 {
		return jobs[0], nil
	}
	return batchv1.Job{}, nil
}
