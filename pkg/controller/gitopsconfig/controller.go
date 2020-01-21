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
	goerrors "errors"
	"time"

	"golang.org/x/xerrors"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	"github.com/KohlsTechnology/eunomia/pkg/util"
)

var log = logf.Log.WithName(controllerName)

const (
	tagInitialized string = "gitopsconfig.eunomia.kohls.io/initialized"
	tagFinalizer   string = "gitopsconfig.eunomia.kohls.io/finalizer"
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
		return xerrors.Errorf("failed creation of controller %q: %w", controllerName, err)
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
		return xerrors.Errorf("controller watch for changes to primary resource GitOpsConfig failed: %w", err)
	}

	err = c.Watch(
		&source.Channel{Source: PushEvents},
		&handler.EnqueueRequestForObject{},
	)
	if err != nil {
		return xerrors.Errorf("controller watch for PushEvents failed: %w", err)
	}

	// TODO: we should somehow detect when Reconciler is stopped, and run the
	// stop func returned by addJobWatch, to not leak resources (though if it's
	// done only once, it's not such a big problem)
	_, err = addJobWatch(mgr.GetConfig(), &jobCompletionEmitter{
		client:        mgr.GetClient(),
		eventRecorder: mgr.GetRecorder(controllerName),
	})
	if err != nil {
		return xerrors.Errorf("cannot create watch job for jobCompletionEmitter handler: %w", err)
	}

	// TODO: detect when Reconciler stops and run stop func returned by addJobWatch
	// TODO: make this and above addJobWatch share a single cache.SharedInformer
	_, err = addJobWatch(mgr.GetConfig(), &statusUpdater{
		client: mgr.GetClient(),
	})
	if err != nil {
		return xerrors.Errorf("cannot create watch job for statusUpdater handler: %w", err)
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
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, xerrors.Errorf("reconciler failed to read GitOpsConfig from kubernetes: %w", err)
	}
	reqLogger.Info("found instance", "instance", instance.GetName())

	//object is being deleted
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.manageDeletion(instance)
	}

	if _, ok := instance.GetAnnotations()[tagInitialized]; !ok {
		reqLogger.Info("Instance needs to be initialized", "instance", instance.GetName())
		return reconcile.Result{}, r.initialize(instance)
	}

	reqLogger.Info("Instance is initialized", "instance", instance.GetName())

	if ContainsTrigger(instance, "Periodic") {
		reqLogger.Info("Instance has a periodic trigger, creating/updating cronjob", "instance", instance.GetName())
		err = r.createCronJob(instance)
		if err != nil {
			reqLogger.Error(err, "error creating the cronjob, continuing...")
		}
	}

	if ContainsTrigger(instance, "Change") || ContainsTrigger(instance, "Webhook") {
		reqLogger.Info("Instance has a change or Webhook trigger, creating job", "instance", instance.GetName())
		reconcileResult, err := r.createJob("create", instance)
		if err != nil {
			reqLogger.Error(err, "reconciler failed to create a job, continuing...", "instance", instance.GetName())
			return reconcileResult, xerrors.Errorf("reconciler failed to create a job for GitOpsConfig instance %q: %w", instance.GetName(), err)
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
	//TODO add logic to ignore if another job was created sooner than x (5 minutes?) time and it is still running.
	mergedata := util.JobMergeData{
		Config: *instance,
		Action: jobtype,
	}
	// looking up for running jobs
	jobList := &batchv1.JobList{}
	err := r.client.List(context.TODO(), &client.ListOptions{
		Namespace: instance.Namespace,
	}, jobList)
	if err != nil {
		log.Error(err, "unable to list the jobs", "namespace", instance.Namespace)
		return reconcile.Result{}, xerrors.Errorf("unable to list the jobs in namespace %q: %w", instance.Namespace, err)
	}
	for _, j := range jobList.Items {
		if isOwner(instance, &j) && j.Status.Active != 0 {
			log.Info("Job is already running for this instance, postponing new job creation", "instance", instance.Name, "job", j.Name)
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Second * 5,
			}, nil
		}
	}
	job, err := util.CreateJob(mergedata)
	if err != nil {
		log.Error(err, "unable to create job manifest from merge data", "mergedata", mergedata)
		return reconcile.Result{}, xerrors.Errorf("unable to create job manifest from merge data: %w", err)
	}
	err = controllerutil.SetControllerReference(instance, &job, r.scheme)
	if err != nil {
		log.Error(err, "unable to set GitOpsConfig instance as Controller OwnerReference on owned job", "instanceName", instance.Name, "job", job)
		return reconcile.Result{}, xerrors.Errorf("unable to set GitOpsConfig instance %q as Controller OwnerReference on owned job %q: %w", instance.Name, job.Name, err)
	}

	log.Info("Creating a new Job", "job.Namespace", job.Namespace, "job.Name", job.Name)
	err = r.client.Create(context.TODO(), &job)
	if err != nil {
		log.Error(err, "unable to create the job", "job", job, "namespace", job.Namespace)
		return reconcile.Result{}, xerrors.Errorf("unable to create the job %q in namespace %q: %w", job.Name, job.Namespace, err)
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
		return xerrors.Errorf("unable to create cronjob manifest from merge data: %w", err)
	}

	err = r.client.Get(context.TODO(),
		types.NamespacedName{Name: cronjob.GetName(), Namespace: cronjob.GetNamespace()},
		&batchv1beta1.CronJob{})
	update := true
	if err != nil {
		if errors.IsNotFound(err) {
			update = false
		} else {
			// Error reading the object - requeue the request.
			return xerrors.Errorf("client failed to retrieve CronJob %q from namespace %q: %w", cronjob.GetName(), cronjob.GetNamespace(), err)
		}
	}

	err = controllerutil.SetControllerReference(instance, &cronjob, r.scheme)
	if err != nil {
		log.Error(err, "unable to set GitOpsConfig instance as Controller OwnerReference on owned cronjob", "instanceName", instance.Name, "cronjob", cronjob)
		return xerrors.Errorf("unable to set GitOpsConfig instance %q as Controller OwnerReference on owned cronjob %q: %w", instance.Name, cronjob.Name, err)
	}
	log.Info("Creating/updating CronJob", "cronjob.Namespace", cronjob.Namespace, "cronjob.Name", cronjob.Name)
	if update {
		err = r.client.Update(context.TODO(), &cronjob)
	} else {
		err = r.client.Create(context.TODO(), &cronjob)
	}

	if err != nil {
		log.Error(err, "unable to create/update the cronjob", "cronjob", cronjob)
		return xerrors.Errorf("unable to create/update the cronjob %q: %w", cronjob.Name, err)
	}
	var result batchv1beta1.CronJob
	err = r.client.Get(context.TODO(),
		types.NamespacedName{Name: cronjob.Name, Namespace: cronjob.Namespace},
		&result)

	if err != nil {
		log.Error(err, "client failed to retrieve CronJob", "cronjob", cronjob.Name, "namespace", cronjob.Namespace)
		return xerrors.Errorf("client failed to retrieve CronJob %q from namespace %q: %w", cronjob.Name, cronjob.Namespace, err)
	}
	return nil
}

// GetAll retrieves all the gitops config in the cluster
func (r *Reconciler) GetAll() (gitopsv1alpha1.GitOpsConfigList, error) {
	instanceList := &gitopsv1alpha1.GitOpsConfigList{}
	err := r.client.List(context.TODO(), &client.ListOptions{}, instanceList)
	if err != nil {
		log.Error(err, "unable to retrieve list of all GitOpsConfig in the cluster")
		return *instanceList, xerrors.Errorf("unable to retrieve list of all GitOpsConfig in the cluster: %w", err)
	}
	return *instanceList, nil
}

func (r *Reconciler) initialize(instance *gitopsv1alpha1.GitOpsConfig) error {
	// verify mandatory field exist and set defaults
	spec := &instance.Spec
	if spec.TemplateSource.URI == "" {
		//TODO set wrong status
		return goerrors.New("template source URI cannot be empty")
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
	meta := &instance.ObjectMeta
	if !containsString(meta.Finalizers, tagFinalizer) && spec.ResourceDeletionMode != "Retain" {
		meta.Finalizers = append(meta.Finalizers, tagFinalizer)
	}
	meta.Annotations[tagInitialized] = "true"

	err := r.client.Update(context.TODO(), instance)
	if err != nil {
		log.Error(err, "unable to update initialized GitOpsConfig", "instance", instance)
		return xerrors.Errorf("unable to update initialized GitOpsConfig %q: %w", instance.Name, err)
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

func (r *Reconciler) manageDeletion(instance *gitopsv1alpha1.GitOpsConfig) (reconcile.Result, error) {
	log.Info("Instance is being deleted", "instance", instance.GetName())
	if !containsString(instance.ObjectMeta.Finalizers, tagFinalizer) {
		return reconcile.Result{}, nil
	}
	// we need to lookup the delete job and if it doesn't exist we launch it, then we see if it is completed successfully if yes we remove the finalizers, if no we return.
	jobList := &batchv1.JobList{}
	selector, err := labels.Parse("action=delete")
	if err != nil {
		log.Error(err, "unable to parse label selector 'action=delete'")
		return reconcile.Result{}, xerrors.Errorf("unable to parse label selector 'action=delete': %w", err)
	}
	// looking up all delete jobs
	err = r.client.List(context.TODO(), &client.ListOptions{
		Namespace:     instance.GetNamespace(),
		LabelSelector: selector,
	}, jobList)
	if err != nil {
		log.Error(err, "unable to list all delete jobs", "namespace", instance.GetNamespace())
		return reconcile.Result{}, xerrors.Errorf("unable to list all delete jobs in namespace %q: %w", instance.GetNamespace(), err)
	}
	applicableJobList := []batchv1.Job{}
	//filtering by those that are might have been created by this gitopsconfig
	// TODO better filter by owner reference
	for _, job := range jobList.Items {
		if isOwner(instance, &job) && job.GetLabels()["action"] == "delete" {
			applicableJobList = append(applicableJobList, job)
		}
	}
	if len(applicableJobList) == 0 {
		// to avoid a deadlock situation let's check that the namespace in which we are is not being deleted
		ns := &corev1.Namespace{}
		err := r.client.Get(context.TODO(), types.NamespacedName{
			Name: instance.GetNamespace(),
		}, ns)
		if err != nil {
			log.Error(err, "unable to lookup instance's namespace", "instanceName", instance.Name)
			return reconcile.Result{}, xerrors.Errorf("unable to lookup instance %q namespace: %w", instance.Name, err)
		}
		if !ns.ObjectMeta.DeletionTimestamp.IsZero() {
			//namespace is being deleted
			// the best we can do in this situation is to let the instance be deleted and hope that this instance was creating objects only in this namespace
			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, tagFinalizer)
			if err := r.client.Update(context.TODO(), instance); err != nil {
				log.Error(err, "unable to create update instance to remove finalizers", "instanceName", instance.Name)
				return reconcile.Result{}, xerrors.Errorf("unable to create update instance %q to remove finalizers: %w", instance.Name, err)
			}
			return reconcile.Result{}, nil
		}
		log.Info("Launching delete job for instance", "instance", instance.GetName())
		_, err = r.createJob("delete", instance)
		if err != nil {
			log.Error(err, "unable to create deletion job for instance", "instanceName", instance.Name)
			return reconcile.Result{}, xerrors.Errorf("unable to create deletion job for instance %q: %w", instance.Name, err)
		}
		//we return because we need to wait for the job to stop
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Second * 5,
		}, nil
	}
	//There should be only one pending job
	job := applicableJobList[0]
	if job.Status.Succeeded > 0 {
		instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, tagFinalizer)
		if err := r.client.Update(context.TODO(), instance); err != nil {
			log.Error(err, "unable to create update instance to remove finalizers", "instanceName", instance.Name)
			return reconcile.Result{}, xerrors.Errorf("unable to create update instance %q to remove finalizers: %w", instance.Name, err)
		}
		return reconcile.Result{}, nil
	}
	//if it's not succeeded we wait for 5 seconds
	//TODO add logic to stop at a certain point ... or not ...
	//TODO add exponential backoff, possibly like in CronJob
	return reconcile.Result{
		Requeue:      true,
		RequeueAfter: time.Second * 5,
	}, nil
}

func isOwner(owner, owned metav1.Object) bool {
	runtimeObj, ok := (owner).(runtime.Object)
	if !ok {
		return false
	}
	for _, ownerRef := range owned.GetOwnerReferences() {
		if ownerRef.Name == owner.GetName() && ownerRef.UID == owner.GetUID() && ownerRef.Kind == runtimeObj.GetObjectKind().GroupVersionKind().Kind {
			return true
		}
	}
	return false
}
