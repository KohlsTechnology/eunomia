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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	"github.com/KohlsTechnology/eunomia/test"
)

const (
	name      = "gitops-operator"
	namespace = "gitops"
)

func TestMain(m *testing.M) {
	// Initialize the environment
	test.Initialize()

	code := m.Run()
	os.Exit(code)
}

func defaultGitOpsConfig() *gitopsv1alpha1.GitOpsConfig {
	return &gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
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
			ResourceDeletionMode:   "Cascade",
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
	// This flag is needed to let the reconciler know that the CRD has been initialized
	gitops.Annotations = map[string]string{"gitopsconfig.eunomia.kohls.io/initialized": "true"}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		gitops,
	}
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(gitopsv1alpha1.SchemeGroupVersion, gitops)
	// Initialize fake client
	cl := fake.NewFakeClient(objs...)
	r := &Reconciler{client: cl, scheme: s}

	nsn := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	req := reconcile.Request{
		NamespacedName: nsn,
	}

	r.Reconcile(req)

	// Check if the CRD has been created
	crd := &gitopsv1alpha1.GitOpsConfig{}
	err := cl.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, crd)
	if err != nil {
		log.Error(err, "Get CRD", "Failed retrieving CRD type of GitOpsConfig")
	}

	// Check if the name matches what was deployed
	assert.Equal(t, crd.Name, nsn.Name)
	// Make sure no errors happened when getting the resource
	assert.NoError(t, err)
}

func TestPeriodicTrigger(t *testing.T) {
	gitops := defaultGitOpsConfig()
	// This flag is needed to let the reconciler know that the CRD has been initialized
	gitops.Annotations = map[string]string{"gitopsconfig.eunomia.kohls.io/initialized": "true"}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		gitops,
	}
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(gitopsv1alpha1.SchemeGroupVersion, gitops)
	// Initialize fake client
	cl := fake.NewFakeClient(objs...)
	r := &Reconciler{client: cl, scheme: s}

	nsn := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	req := reconcile.Request{
		NamespacedName: nsn,
	}

	r.Reconcile(req)

	// Check if the CRD has been created
	cron := &batchv1beta1.CronJob{}
	err := cl.Get(context.TODO(), types.NamespacedName{Name: "gitopsconfig-gitops-operator", Namespace: namespace}, cron)

	if err != nil {
		log.Error(err, "Get Cron", "Failed retrieving type of CronJob")
	}

	// Check if the name matches what was deployed
	assert.Equal(t, cron.Name, "gitopsconfig-gitops-operator")
	// Make sure no errors happened when getting the resource
	assert.NoError(t, err)
}

func TestChangeTrigger(t *testing.T) {
	gitops := defaultGitOpsConfig()
	// This flag is needed to let the reconciler know that the CRD has been initialized
	gitops.Annotations = map[string]string{"gitopsconfig.eunomia.kohls.io/initialized": "true"}
	// Set trigger type to Change
	gitops.Spec.Triggers = []gitopsv1alpha1.GitOpsTrigger{
		{
			Type: "Change",
		},
	}
	// Objects to track in the fake client.
	objs := []runtime.Object{
		gitops,
	}
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(gitopsv1alpha1.SchemeGroupVersion, gitops)
	// Initialize fake client
	cl := fake.NewFakeClient(objs...)
	r := &Reconciler{client: cl, scheme: s}

	nsn := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	req := reconcile.Request{
		NamespacedName: nsn,
	}

	r.Reconcile(req)

	// Check if the CRD has been created
	job := &batchv1.Job{}
	err := cl.Get(context.TODO(), types.NamespacedName{Namespace: namespace}, job)

	if err != nil {
		log.Error(err, "Get Job", "Failed retrieving type of Job")
	}

	// Check if the name matches what was deployed
	assert.Equal(t, job.Kind, "Job")
	// Make sure no errors happened when getting the resource
	assert.NoError(t, err)
}

func TestWebhookTrigger(t *testing.T) {
	gitops := defaultGitOpsConfig()
	// This flag is needed to let the reconciler know that the CRD has been initialized
	gitops.Annotations = map[string]string{"gitopsconfig.eunomia.kohls.io/initialized": "true"}
	// Set trigger type to Webhook
	gitops.Spec.Triggers = []gitopsv1alpha1.GitOpsTrigger{
		{
			Type: "Webhook",
		},
	}
	// Objects to track in the fake client.
	objs := []runtime.Object{
		gitops,
	}
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(gitopsv1alpha1.SchemeGroupVersion, gitops)
	// Initialize fake client
	cl := fake.NewFakeClient(objs...)
	r := &Reconciler{client: cl, scheme: s}

	nsn := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	req := reconcile.Request{
		NamespacedName: nsn,
	}

	r.Reconcile(req)

	// Check if the CRD has been created
	job := &batchv1.Job{}
	err := cl.Get(context.TODO(), types.NamespacedName{Namespace: namespace}, job)

	if err != nil {
		log.Error(err, "Get Job", "Failed retrieving type of Job")
	}

	// Check if the name matches what was deployed
	assert.Equal(t, job.Kind, "Job")
	// Make sure no errors happened when getting the resource
	assert.NoError(t, err)
}

func TestDeleteRemovingFinalizer(t *testing.T) {
	gitops := defaultGitOpsConfig()
	// This flag is needed to let the reconciler know that the CRD has been initialized
	gitops.Annotations = map[string]string{"gitopsconfig.eunomia.kohls.io/initialized": "true"}
	gitops.Spec.Triggers = []gitopsv1alpha1.GitOpsTrigger{
		{
			Type: "Change",
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		gitops,
	}
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(gitopsv1alpha1.SchemeGroupVersion, gitops)
	// Initialize fake client
	cl := fake.NewFakeClient(objs...)
	r := &Reconciler{client: cl, scheme: s}

	// Create a namespace
	err := cl.Create(context.TODO(), defaultNamespace())
	if err != nil {
		log.Error(err, "Namespace", "Failed to create namespace")
	}

	nsn := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	req := reconcile.Request{
		NamespacedName: nsn,
	}

	r.Reconcile(req)

	// Add a finalizer to the CRD
	gitops.ObjectMeta.Finalizers = append(gitops.ObjectMeta.Finalizers, "gitopsconfig.eunomia.kohls.io/finalizer")
	err = cl.Update(context.Background(), gitops)
	if err != nil {
		log.Error(err, "Add Finalizer", "Failed adding finalizer to CRD")
	}

	// Get the CRD so that we can add the deletion timestamp
	crd := &gitopsv1alpha1.GitOpsConfig{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, crd)
	if err != nil {
		log.Error(err, "Get CRD", "Failed retrieving CRD type of GitOpsConfig")
	}
	// Make sure the finalizer has been added
	assert.NotEmpty(t, crd.ObjectMeta.Finalizers)

	// Set deletion timestamp
	deleteTime := metav1.Now()
	crd.ObjectMeta.DeletionTimestamp = &deleteTime
	// Update the CRD with the new deletion timestamp
	err = cl.Update(context.TODO(), crd)
	if err != nil {
		log.Error(err, "Update CRD", "Failed Updating CRD type of GitOpsConfig")
	}
	// Create the deleteJob
	var (
		dummyInt32 int32 = 1
		dummyBool  bool  = true
	)
	err = cl.Create(context.TODO(), &batchv1.Job{
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
					Name:               name,
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
		log.Error(err, "Create Job", "Failed creating Job type action of Delete")
	}
	// Reconcile so that the controller can delete the finalizer
	r.Reconcile(req)

	// Check the status
	crd = &gitopsv1alpha1.GitOpsConfig{}
	err = cl.Get(context.TODO(), types.NamespacedName{Namespace: namespace}, crd)
	// The finalizer should have been removed
	assert.Empty(t, crd.ObjectMeta.Finalizers)
	assert.NoError(t, err)
}

func TestCreatingDeleteJob(t *testing.T) {
	gitops := defaultGitOpsConfig()
	// This flag is needed to let the reconciler know that the CRD has been initialized
	gitops.Annotations = map[string]string{"gitopsconfig.eunomia.kohls.io/initialized": "true"}
	gitops.Spec.Triggers = []gitopsv1alpha1.GitOpsTrigger{
		{
			Type: "Change",
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		gitops,
	}
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(gitopsv1alpha1.SchemeGroupVersion, gitops)
	// Initialize fake client
	cl := fake.NewFakeClient(objs...)
	r := &Reconciler{client: cl, scheme: s}

	// Create a namespace
	err := cl.Create(context.TODO(), defaultNamespace())
	if err != nil {
		log.Error(err, "Namespace", "Failed to create namespace")
	}

	nsn := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	req := reconcile.Request{
		NamespacedName: nsn,
	}

	r.Reconcile(req)

	ns := &corev1.Namespace{}
	err = cl.Get(context.TODO(), types.NamespacedName{
		Name: namespace,
	}, ns)
	if err != nil {
		log.Error(err, "unable to lookup instance's namespace")
	}

	// Add a finalizer to the CRD
	gitops.ObjectMeta.Finalizers = append(gitops.ObjectMeta.Finalizers, "gitopsconfig.eunomia.kohls.io/finalizer")
	err = cl.Update(context.Background(), gitops)
	if err != nil {
		log.Error(err, "Add Finalizer", "Failed adding finalizer to CRD")
	}

	// Get the CRD so that we can add the deletion timestamp
	crd := &gitopsv1alpha1.GitOpsConfig{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, crd)
	if err != nil {
		log.Error(err, "Get CRD", "Failed retrieving CRD type of GitOpsConfig")
	}
	// Make sure the finalizer has been added
	assert.NotEmpty(t, crd.ObjectMeta.Finalizers)

	// Make sure there's no delete job
	job := findDeleteJob(cl)
	assert.NotEqual(t, "delete", job.GetLabels()["action"])

	// Set deletion timestamp
	deleteTime := metav1.Now()
	crd.ObjectMeta.DeletionTimestamp = &deleteTime
	// Update the CRD with the new deletion timestamp
	err = cl.Update(context.TODO(), crd)
	if err != nil {
		log.Error(err, "Update CRD", "Failed Updating CRD type of GitOpsConfig")
	}

	// Fakeclient is not updating the job status , inorder to create the new job we are
	// Updating the job status manually for the existing job created by Reconcile.
	job = findRunningJob(cl)
	job.Status.Active = 0
	job.Status.Succeeded = 1
	job.Status.Failed = 0
	job.Status.StartTime = &deleteTime
	// Update the job with the status
	err = cl.Update(context.TODO(), &job)
	if err != nil {
		log.Error(err, "Update job", "Failed to Updating the job status")
	}

	// There shouldn't be a delete job at this point, the reconciler should create one
	r.Reconcile(req)

	// See if a delete job was created
	job = findDeleteJob(cl)
	assert.Equal(t, "delete", job.GetLabels()["action"])
}

func TestDeleteWhileNamespaceDeleting(t *testing.T) {
	gitops := defaultGitOpsConfig()
	// This flag is needed to let the reconciler know that the CRD has been initialized
	gitops.Annotations = map[string]string{"gitopsconfig.eunomia.kohls.io/initialized": "true"}
	gitops.Spec.Triggers = []gitopsv1alpha1.GitOpsTrigger{
		{
			Type: "Change",
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		gitops,
	}
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(gitopsv1alpha1.SchemeGroupVersion, gitops)
	// Initialize fake client
	cl := fake.NewFakeClient(objs...)
	r := &Reconciler{client: cl, scheme: s}

	// Create a namespace
	// Set deletion timestamp on the namespace
	deleteTime := metav1.Now()
	ns0 := defaultNamespace()
	ns0.ObjectMeta.DeletionTimestamp = &deleteTime
	err := cl.Create(context.TODO(), ns0)
	if err != nil {
		log.Error(err, "Namespace", "Failed to create namespace")
	}

	nsn := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	req := reconcile.Request{
		NamespacedName: nsn,
	}

	r.Reconcile(req)

	ns := &corev1.Namespace{}
	err = cl.Get(context.TODO(), types.NamespacedName{
		Name: namespace,
	}, ns)
	if err != nil {
		log.Error(err, "unable to lookup instance's namespace")
	}

	// Add a finalizer to the CRD
	gitops.ObjectMeta.Finalizers = append(gitops.ObjectMeta.Finalizers, "gitopsconfig.eunomia.kohls.io/finalizer")
	err = cl.Update(context.Background(), gitops)
	if err != nil {
		log.Error(err, "Add Finalizer", "Failed adding finalizer to CRD")
	}

	// Get the CRD so that we can add the deletion timestamp
	crd := &gitopsv1alpha1.GitOpsConfig{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, crd)
	if err != nil {
		log.Error(err, "Get CRD", "Failed retrieving CRD type of GitOpsConfig")
	}
	// Make sure the finalizer has been added
	assert.NotEmpty(t, crd.ObjectMeta.Finalizers)

	// Set deletion timestamp
	deleteTime = metav1.Now()
	crd.ObjectMeta.DeletionTimestamp = &deleteTime
	// Update the CRD with the new deletion timestamp
	err = cl.Update(context.TODO(), crd)
	if err != nil {
		log.Error(err, "Update CRD", "Failed Updating CRD type of GitOpsConfig")
	}

	// There shouldn't be a delete job at this point, the reconciler should create one
	r.Reconcile(req)

	// Check the status
	crd = &gitopsv1alpha1.GitOpsConfig{}
	err = cl.Get(context.TODO(), types.NamespacedName{Namespace: namespace}, crd)
	// The finalizer should have been removed
	assert.Empty(t, crd.ObjectMeta.Finalizers)
	assert.NoError(t, err)
}

func findDeleteJob(cl client.Client) batchv1.Job {
	// At times other jobs can exist
	jobList := &batchv1.JobList{}
	// Looking up all jobs
	err := cl.List(context.TODO(), &client.ListOptions{
		Namespace: namespace,
	}, jobList)
	if err != nil {
		log.Error(err, "unable to list jobs")
		return batchv1.Job{}
	}
	// Return the first instance that is a delete job
	for _, job := range jobList.Items {
		if job.GetLabels()["action"] == "delete" {
			return job
		}
	}
	return batchv1.Job{}
}

func TestCreateJob(t *testing.T) {
	gitops := defaultGitOpsConfig()
	// This flag is needed to let the reconciler know that the CRD has been initialized
	gitops.Annotations = map[string]string{"gitopsconfig.eunomia.kohls.io/initialized": "true"}
	// Set trigger type to Change
	gitops.Spec.Triggers = []gitopsv1alpha1.GitOpsTrigger{
		{
			Type: "Change",
		},
	}
	// Objects to track in the fake client.
	objs := []runtime.Object{
		gitops,
	}
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(gitopsv1alpha1.SchemeGroupVersion, gitops)
	// Initialize fake client
	cl := fake.NewFakeClient(objs...)
	r := &Reconciler{client: cl, scheme: s}

	nsn := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	req := reconcile.Request{
		NamespacedName: nsn,
	}

	r.Reconcile(req)

	// Fakeclient is not updating the job status , inorder to test race condition between the jobs we are
	// Updating the job status manually for the existing job created by Reconcile.
	startTime := metav1.Now()
	job := findRunningJob(cl)
	job.Status.Active = 1
	job.Status.Succeeded = 0
	job.Status.Failed = 0
	job.Status.StartTime = &startTime

	err := cl.Update(context.TODO(), &job)
	if err != nil {
		log.Error(err, "Update job", "Failed to Updating the job status")
	}
	r.Reconcile(req)
	jobCount, err := findJobList(cl)
	if err != nil {
		log.Error(err, "Job list", "Failed to fetch job list")
	}
	if jobCount > 1 {
		t.Error("Job was not postponed")
	}
}

func findJobList(cl client.Client) (int, error) {
	// At times other jobs can exist
	jobList := &batchv1.JobList{}
	// Looking up all jobs
	err := cl.List(context.TODO(), &client.ListOptions{
		Namespace: namespace,
	}, jobList)
	if err != nil {
		log.Error(err, "unable to list the running jobs")
		return 0, err
	}
	return len(jobList.Items), nil
}

func findRunningJob(cl client.Client) batchv1.Job {
	// At times other jobs can exist
	jobList := &batchv1.JobList{}
	// Looking up all jobs
	err := cl.List(context.TODO(), &client.ListOptions{
		Namespace: namespace,
	}, jobList)
	if err != nil {
		log.Error(err, "unable to list jobs")
		return batchv1.Job{}
	}
	// Returning the jobs
	if len(jobList.Items) > 0 {
		return jobList.Items[0]
	}
	return batchv1.Job{}
}
