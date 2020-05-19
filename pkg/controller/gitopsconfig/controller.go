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
	"errors"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	"github.com/KohlsTechnology/eunomia/pkg/util"
)

var log = logf.Log.WithName(controllerName)

const (
	tagInitialized string = "gitopsconfig.eunomia.kohls.io/initialized"
	tagFinalizer   string = "gitopsconfig.eunomia.kohls.io/finalizer"
	tagJobOwner    string = "gitopsconfig.eunomia.kohls.io/jobOwner"
	controllerName string = "gitopsconfig-controller"
)

// PushEvents channel on which we get the github webhook push events
var PushEvents = make(chan event.GenericEvent)

// Add creates a new GitOpsConfig Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// NewReconciler creates a new git ops reconciler
func NewReconciler(mgr manager.Manager) Reconciler {
	return Reconciler{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &Reconciler{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return fmt.Errorf("failed creation of controller %q: %w", controllerName, err)
	}

	// Watch for changes to primary resource GitOpsConfig
	err = c.Watch(
		&source.Kind{Type: &gitopsv1alpha1.GitOpsConfig{}},
		&handler.EnqueueRequestForObject{},
		// TODO: once we update to sigs.k8s.io/controller-runtime >=0.2.0, use their
		// .../pkg/predicate.GenerationChangedPredicate instead of rewriting it on our own
		predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				if e.MetaOld == nil {
					log.Error(nil, "Update event has no old metadata", "event", e)
					return false
				}
				if e.MetaNew == nil {
					log.Error(nil, "Update event has no new metadata", "event", e)
					return false
				}
				// If there's a status update, .metadata.Generation field isn't changed - ignore such event
				return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
			},
		},
	)
	if err != nil {
		return fmt.Errorf("controller watch for changes to primary resource GitOpsConfig failed: %w", err)
	}

	err = c.Watch(
		&source.Channel{Source: PushEvents},
		&handler.EnqueueRequestForObject{},
	)
	if err != nil {
		return fmt.Errorf("controller watch for PushEvents failed: %w", err)
	}

	// TODO: we should somehow detect when Reconciler is stopped, and run the
	// stop func returned by addJobWatch, to not leak resources (though if it's
	// done only once, it's not such a big problem)
	_, err = addJobWatch(mgr.GetConfig(), &jobCompletionEmitter{
		client:        mgr.GetClient(),
		eventRecorder: mgr.GetEventRecorderFor(controllerName),
	})
	if err != nil {
		return fmt.Errorf("cannot create watch job for jobCompletionEmitter handler: %w", err)
	}

	// TODO: detect when Reconciler stops and run stop func returned by addJobWatch
	// TODO: make this and above addJobWatch share a single cache.SharedInformer
	_, err = addJobWatch(mgr.GetConfig(), &statusUpdater{
		client: mgr.GetClient(),
	})
	if err != nil {
		return fmt.Errorf("cannot create watch job for statusUpdater handler: %w", err)
	}

	return nil
}

var _ reconcile.Reconciler = &Reconciler{}

// Reconciler reconciles a GitOpsConfig object
type Reconciler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a GitOpsConfig object and makes changes based on the state read
// and what is in the GitOpsConfig.Spec
//
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// TODO general algorithm:
	// if delete and delete cascade allocate the delete pod
	// if periodic and cronjob does not exist, create cronjob
	// if change create job

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling GitOpsConfig")
	// Fetch the GitOpsConfig instance
	instance := &gitopsv1alpha1.GitOpsConfig{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, fmt.Errorf("reconciler failed to read GitOpsConfig from kubernetes: %w", err)
	}
	reqLogger.Info("found instance", "instance", instance.GetName())

	//object is being deleted
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.manageDeletion(instance)
	}

	if _, ok := instance.GetAnnotations()[tagInitialized]; !ok {
		reqLogger.Info("Instance needs to be initialized", "instance", instance.GetName())
		return reconcile.Result{Requeue: true}, r.initialize(instance)
	}

	if syncFinalizer(instance) {
		err := r.client.Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "syncing finalizer")
			return reconcile.Result{RequeueAfter: 5 * time.Second}, fmt.Errorf("syncing finalizer on %s.%s: %w", request.Namespace, request.Name, err)
		}
		return reconcile.Result{Requeue: true}, nil
	}

	reqLogger.Info("Instance is initialized", "instance", instance.GetName())

	if ContainsTrigger(instance, "Periodic") {
		reqLogger.Info("Instance has a periodic trigger, creating/updating cronjob", "instance", instance.GetName())
		err = r.createCronJob(instance)
		if err != nil {
			reqLogger.Error(err, "error creating the cronjob, continuing...")
		}
	} else {
		// if there are some leftover cronjobs after removing the Periodic trigger, delete them
		cronJobs, err := ownedCronJobs(context.TODO(), r.client, instance)
		if err != nil {
			reqLogger.Error(err, "unable to list cronjobs", "namespace", instance.Namespace)
			return reconcile.Result{}, fmt.Errorf("unable to list cronjobs while checking if there are any left after updating GitOpsConfig: %w", err)
		}
		for _, cronJob := range cronJobs {
			err = r.client.Delete(context.TODO(), &cronJob, client.PropagationPolicy(metav1.DeletePropagationBackground))
			if err != nil {
				log.Error(err, "Unable to delete leftover cronjob", "instance", instance.Name, "cronjob", cronJob.Name)
				return reconcile.Result{}, fmt.Errorf("Unable to delete leftover cronjob %q for %q: %w", cronJob.Name, instance.Name, err)
			}
			log.Info("Deleted leftover cronjob", "instance", instance.Name, "cronjob", cronJob.Name)
		}
	}

	if ContainsTrigger(instance, "Change") || ContainsTrigger(instance, "Webhook") {
		reqLogger.Info("Instance has a change or Webhook trigger, creating job", "instance", instance.GetName())
		reconcileResult, err := r.createJob("create", instance)
		if err != nil {
			reqLogger.Error(err, "reconciler failed to create a job, continuing...", "instance", instance.GetName())
			return reconcileResult, fmt.Errorf("reconciler failed to create a job for GitOpsConfig instance %q: %w", instance.GetName(), err)
		}
		return reconcileResult, nil
	}

	return reconcile.Result{}, err
}

// ContainsTrigger returns true if the passed instance contains the given trigger
func ContainsTrigger(instance *gitopsv1alpha1.GitOpsConfig, triggeType string) bool {
	for _, trigger := range instance.Spec.Triggers {
		if trigger.Type == triggeType {
			return true
		}
	}
	return false
}

// createJob creates a new gitops job for the passed instance
func (r *Reconciler) createJob(jobtype string, instance *gitopsv1alpha1.GitOpsConfig) (reconcile.Result, error) {
	// looking up for running jobs, to avoid creating duplicate one
	jobs, err := ownedJobs(context.TODO(), r.client, instance)
	if err != nil {
		log.Error(err, "unable to list the jobs", "namespace", instance.Namespace)
		return reconcile.Result{}, fmt.Errorf("unable to list owned jobs when trying to create new one: %w", err)
	}
	for _, j := range jobs {
		if j.Status.Active != 0 || j.Status.StartTime.IsZero() {
			log.Info("Job is already running for this instance, postponing new job creation", "instance", instance.Name, "job", j.Name)
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Second * 5,
			}, nil
		}
	}

	mergedata := util.JobMergeData{
		Config: *instance,
		Action: jobtype,
	}
	job, err := util.CreateJob(mergedata)
	if err != nil {
		log.Error(err, "unable to create job manifest from merge data", "mergedata", mergedata)
		return reconcile.Result{}, fmt.Errorf("unable to create job manifest from merge data: %w", err)
	}
	err = controllerutil.SetControllerReference(instance, &job, r.scheme)
	if err != nil {
		log.Error(err, "unable to set GitOpsConfig instance as Controller OwnerReference on owned job", "instanceName", instance.Name, "job", job)
		return reconcile.Result{}, fmt.Errorf("unable to set GitOpsConfig instance %q as Controller OwnerReference on owned job %q: %w", instance.Name, job.Name, err)
	}

	log.Info("Creating a new Job", "job.Namespace", job.Namespace, "job.Name", job.Name)
	err = r.client.Create(context.TODO(), &job)
	if err != nil {
		log.Error(err, "unable to create the job", "job", job, "namespace", job.Namespace)
		return reconcile.Result{}, fmt.Errorf("unable to create the job %q in namespace %q: %w", job.Name, job.Namespace, err)
	}
	return reconcile.Result{}, nil
}

func (r *Reconciler) createCronJob(instance *gitopsv1alpha1.GitOpsConfig) error {
	mergedata := util.JobMergeData{
		Config: *instance,
		Action: "create",
	}

	cronjob, err := util.CreateCronJob(mergedata)
	if err != nil {
		log.Error(err, "unable to create cronjob manifest from merge data", "mergedata", mergedata)
		return fmt.Errorf("unable to create cronjob manifest from merge data: %w", err)
	}

	err = r.client.Get(context.TODO(), util.GetNN(&cronjob), &batchv1beta1.CronJob{})
	update := true
	if err != nil {
		if apierrors.IsNotFound(err) {
			update = false
		} else {
			// Error reading the object - requeue the request.
			return fmt.Errorf("client failed to retrieve CronJob %q from namespace %q: %w", cronjob.GetName(), cronjob.GetNamespace(), err)
		}
	}

	err = controllerutil.SetControllerReference(instance, &cronjob, r.scheme)
	if err != nil {
		log.Error(err, "unable to set GitOpsConfig instance as Controller OwnerReference on owned cronjob", "instanceName", instance.Name, "cronjob", cronjob)
		return fmt.Errorf("unable to set GitOpsConfig instance %q as Controller OwnerReference on owned cronjob %q: %w", instance.Name, cronjob.Name, err)
	}
	log.Info("Creating/updating CronJob", "cronjob.Namespace", cronjob.Namespace, "cronjob.Name", cronjob.Name)
	if update {
		err = r.client.Update(context.TODO(), &cronjob)
	} else {
		err = r.client.Create(context.TODO(), &cronjob)
	}

	if err != nil {
		log.Error(err, "unable to create/update the cronjob", "cronjob", cronjob)
		return fmt.Errorf("unable to create/update the cronjob %q: %w", cronjob.Name, err)
	}
	var result batchv1beta1.CronJob
	err = r.client.Get(context.TODO(), util.GetNN(&cronjob), &result)

	if err != nil {
		log.Error(err, "client failed to retrieve CronJob", "cronjob", cronjob.Name, "namespace", cronjob.Namespace)
		return fmt.Errorf("client failed to retrieve CronJob %q from namespace %q: %w", cronjob.Name, cronjob.Namespace, err)
	}
	return nil
}

// GetAll retrieves all the gitops config in the cluster
func (r *Reconciler) GetAll() (gitopsv1alpha1.GitOpsConfigList, error) {
	instanceList := &gitopsv1alpha1.GitOpsConfigList{}

	err := r.client.List(context.TODO(), instanceList, []client.ListOption{}...)

	if err != nil {
		log.Error(err, "unable to retrieve list of all GitOpsConfig in the cluster")
		return *instanceList, fmt.Errorf("unable to retrieve list of all GitOpsConfig in the cluster: %w", err)
	}
	return *instanceList, nil
}

func (r *Reconciler) initialize(instance *gitopsv1alpha1.GitOpsConfig) error {
	// verify mandatory field exist and set defaults
	spec := &instance.Spec
	if spec.TemplateSource.URI == "" {
		//TODO set wrong status
		return errors.New("template source URI cannot be empty")
	}
	replaceEmpty(&spec.TemplateSource.Ref, "master")
	replaceEmpty(&spec.TemplateSource.ContextDir, ".")
	replaceEmpty(&spec.ParameterSource.URI, spec.TemplateSource.URI)
	replaceEmpty(&spec.ParameterSource.Ref, "master")
	replaceEmpty(&spec.ParameterSource.ContextDir, ".")
	replaceEmpty(&spec.ServiceAccountRef, "default")
	replaceEmpty(&spec.ResourceHandlingMode, "Apply")
	replaceEmpty(&spec.ResourceDeletionMode, "Delete")

	// add finalizer and mark the object as initialized
	syncFinalizer(instance)
	if instance.Annotations == nil {
		instance.Annotations = map[string]string{}
	}
	instance.Annotations[tagInitialized] = "true"

	err := r.client.Update(context.TODO(), instance)
	if err != nil {
		log.Error(err, "unable to update initialized GitOpsConfig", "instance", instance)
		return fmt.Errorf("unable to update initialized GitOpsConfig %q: %w", instance.Name, err)
	}
	return nil
}

// replaceEmpty sets s to defaultValue if s is empty
func replaceEmpty(s *string, defaultValue string) {
	if *s == "" {
		*s = defaultValue
	}
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// syncFinalizer adds or removes finalizer in instance depending on its
// ResourceDeletionMode field. The function returns true if it modified the
// instance. Note: the function does only local modification, propagating the
// change into the cluster is the caller's responsibility.
func syncFinalizer(instance *gitopsv1alpha1.GitOpsConfig) bool {
	var (
		found  = containsString(instance.Finalizers, tagFinalizer)
		wanted = instance.Spec.ResourceDeletionMode != "Retain"
	)
	switch {
	case wanted && !found:
		instance.Finalizers = append(instance.Finalizers, tagFinalizer)
		return true
	case !wanted && found:
		instance.Finalizers = removeString(instance.Finalizers, tagFinalizer)
		return true
	default:
		return false
	}
}

func (r *Reconciler) manageDeletion(instance *gitopsv1alpha1.GitOpsConfig) (reconcile.Result, error) {
	log.Info("Instance is being deleted", "instance", instance.GetName())
	if !containsString(instance.ObjectMeta.Finalizers, tagFinalizer) {
		return reconcile.Result{}, nil
	}

	// To avoid a deadlock situation let's check if the namespace in which we are is maybe being deleted
	ns := &corev1.Namespace{}
	err := r.client.Get(context.TODO(), util.NN{Name: instance.GetNamespace(), Namespace: instance.GetNamespace()}, ns)
	if err != nil {
		log.Error(err, "GitOpsConfig finalizer unable to lookup instance's namespace", "instance", instance.Name)
		return reconcile.Result{}, fmt.Errorf("GitOpsConfig finalizer unable to lookup instance's namespace for %q: %w", instance.Name, err)
	}
	if !ns.DeletionTimestamp.IsZero() {
		// Namespace is being deleted. The best we can do in this situation is
		// to let the instance be deleted and hope that this instance was
		// creating objects only in this namespace
		log.Info("Namespace is being deleted, removing finalizer", "namespace", instance.Namespace, "instance", instance.Name)
		return r.removeFinalizer(context.TODO(), instance)
	}

	// TODO: also search and delete a CronJob

	// We list all jobs that were created because of this GitOpsConfig. Then,
	// further down, we will take one of a few different actions depending on
	// the contents of this list.
	jobs, err := ownedJobs(context.TODO(), r.client, instance)
	if err != nil {
		log.Error(err, "GitOpsConfig finalizer unable to list owned jobs", "instance", instance.Name)
		return reconcile.Result{}, fmt.Errorf("GitOpsConfig finalizer unable to list owned jobs for %q: %w", instance.Name, err)
	}

	log.Info("Active Jobs", "n", len(jobs), "instance", instance.Name)

	// If exactly 1 job exists, but it's blocked because of bad image, we
	// assume the GitOpsConfig never managed to successfully deploy, so we can
	// just delete the job, remove the finalizer, and be done (#216). It may be
	// either action=create or action=delete job.
	// If a job is blocked because of bad image, it only has one active pod
	if len(jobs) == 1 && jobs[0].Status.Succeeded == 0 && jobs[0].Status.Failed == 0 && jobs[0].Status.Active == 1 {
		status, err := jobContainerStatus(context.TODO(), r.client, &jobs[0])
		if err != nil {
			log.Error(err, "GitOpsConfig finalizer unable to get job pod's status", "instance", instance.Name)
			return reconcile.Result{}, fmt.Errorf("GitOpsConfig finalizer unable to get job pod's status for %q: %w", instance.Name, err)
		}
		log.Info("GitOpsConfig finalizer found one job", "instance", instance.Name, "podStatus", status)
		safeReasons := []string{"ErrImagePull", "ImagePullBackOff", "InvalidImageName"}
		if status != nil && status.Waiting != nil && containsString(safeReasons, status.Waiting.Reason) {
			err = r.client.Delete(context.TODO(), &jobs[0], client.PropagationPolicy(metav1.DeletePropagationBackground))
			if err != nil {
				log.Error(err, "GitOpsConfig finalizer unable to delete job", "instance", instance.Name, "job", jobs[0].Name)
				return reconcile.Result{}, fmt.Errorf("GitOpsConfig finalizer unable to delete job %q for %q: %w", jobs[0].Name, instance.Name, err)
			}
			log.Info("GitOpsConfig finalizer deleted stuck job", "instance", instance.Name, "job", jobs[0].Name)
			return r.removeFinalizer(context.TODO(), instance)
		}
	}

	// If a delete job exists, we wait (Requeue) until we can see that it
	// completed successfully, then we remove the finalizer and we're done.
	deleters := []batchv1.Job{}
	for _, j := range jobs {
		if j.Labels != nil && j.Labels["action"] == "delete" {
			deleters = append(deleters, j)
		}
	}
	log.Info("Delete Jobs", "n", len(deleters), "instance", instance.Name)
	if len(deleters) > 0 {
		if len(deleters) > 1 {
			log.Error(nil, "too many delete jobs found (expected 1)", "n", len(deleters), "instance", instance.Name)
			// TODO: should we return here, or try to still do something sensible for user?
		}
		done := deleters[0].Status.Succeeded > 0
		if !done {
			// if it's not succeeded we wait for 5 seconds
			// TODO add logic to stop at a certain point ... or not ...
			// TODO add exponential backoff, possibly like in CronJob
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}
		return r.removeFinalizer(context.TODO(), instance)
	}

	log.Info("Launching delete job for instance", "instance", instance.Name)
	_, err = r.createJob("delete", instance)
	if err != nil {
		log.Error(err, "unable to create deletion job", "instance", instance.Name)
		return reconcile.Result{}, err
	}
	// we return because we need to wait for the job to stop
	return reconcile.Result{
		Requeue:      true,
		RequeueAfter: time.Second * 5,
	}, nil

}

func (r *Reconciler) removeFinalizer(ctx context.Context, instance *gitopsv1alpha1.GitOpsConfig) (reconcile.Result, error) {
	instance.Finalizers = removeString(instance.Finalizers, tagFinalizer)
	err := r.client.Update(ctx, instance)
	if err != nil {
		// if apierrors.IsConflict, then requeue
		var errAPI apierrors.APIStatus
		if errors.As(err, &errAPI) && errAPI.Status().Reason == metav1.StatusReasonConflict {
			log.Error(err, "GitOpsConfig finalizer unable to remove itself; will retry", "instance", instance.Name)
			return reconcile.Result{
				RequeueAfter: 5 * time.Second,
			}, fmt.Errorf("GitOpsConfig finalizer unable to remove itself from %q, will retry: %w", instance.Name, err)
		}
		log.Error(err, "GitOpsConfig finalizer unable to remove itself", "instance", instance.Name)
		return reconcile.Result{}, fmt.Errorf("GitOpsConfig finalizer unable to remove itself from %q: %w", instance.Name, err)
	}
	log.Info("GitOpsConfig finalizer successfully removed itself from CR", "instance", instance.Name)
	return reconcile.Result{}, nil
}

// ownedCronJobs retrieves all cronjobs in namespace owner.Namespace whose owner is the passed GitOpsConfig.
func ownedCronJobs(ctx context.Context, kube client.Client, owner *gitopsv1alpha1.GitOpsConfig) ([]batchv1beta1.CronJob, error) {
	cronJobs := batchv1beta1.CronJobList{}
	listOpts := []client.ListOption{
		client.InNamespace(owner.Namespace),
	}
	err := kube.List(ctx, &cronJobs, listOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to list cronjobs in namespace %s: %w", owner.Namespace, err)
	}

	owned := []batchv1beta1.CronJob{}
	for _, cronJob := range cronJobs.Items {
		ownerRefs := cronJob.GetOwnerReferences()
		for _, ownerRef := range ownerRefs {
			if *ownerRef.Controller == true &&
				ownerRef.APIVersion == owner.APIVersion &&
				ownerRef.Kind == owner.Kind &&
				ownerRef.Name == owner.ObjectMeta.Name {
				owned = append(owned, cronJob)
				log.Info("ownedCronJobs", "Name", cronJob.Name, "Owner", owner.Name, "Namespace", owner.Namespace)
				break
			}
		}
	}
	return owned, nil
}

// ownedJobs retrieves all jobs in namespace owner.Namespace with value of label tagJobOwner equal to owner.Name.
func ownedJobs(ctx context.Context, kube client.Client, owner *gitopsv1alpha1.GitOpsConfig) ([]batchv1.Job, error) {
	jobs := &batchv1.JobList{}
	listOpts := []client.ListOption{
		client.InNamespace(owner.Namespace),
		client.MatchingLabels{tagJobOwner: owner.Name},
	}
	err := kube.List(ctx, jobs, listOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs for jobOwner==%q (ns: %s): %w", owner.Name, owner.Namespace, err)
	}
	for _, j := range jobs.Items {
		log.Info("ownedJobs", "Name", j.Name, "Owner", owner.Name, "Namespace", owner.Namespace)
	}
	return jobs.Items, nil
}

// jobContainerStatus retrieves the job's pod from kube cluster, and returns
// the status of its container. If the job controls more than one pod, or the
// pod contains more than one container, an error is returned. If the job
// controls no pods, nil is returned.
func jobContainerStatus(ctx context.Context, kube client.Client, job *batchv1.Job) (*corev1.ContainerState, error) {
	// Find pod(s) of the job
	pods := &corev1.PodList{}

	listOpts := []client.ListOption{
		client.InNamespace(job.Namespace),
		client.MatchingLabels{"job-name": job.Name},
	}
	err := kube.List(ctx, pods, listOpts...)

	if err != nil {
		return nil, fmt.Errorf("unable to list pods for job %q: %w", job.Name, err)
	}
	if len(pods.Items) == 0 {
		return nil, nil
	} else if len(pods.Items) > 1 {
		return nil, fmt.Errorf("in Job %q expected 1 pod, got %d", job.Name, len(pods.Items))
	}
	// Get status of container(s) of pod
	stats := pods.Items[0].Status.ContainerStatuses
	if len(stats) != 1 {
		return nil, fmt.Errorf("in Pod %q expected 1 container, got %d", pods.Items[0].Name, len(stats))
	}
	return &stats[0].State, nil
}
