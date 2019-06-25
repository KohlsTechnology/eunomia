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

	"k8s.io/apimachinery/pkg/labels"

	"time"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	util "github.com/KohlsTechnology/eunomia/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_gitopsconfig")

const initLabel string = "gitopsconfig.eunomia.kohls.io/initialized"
const kubeGitopsFinalizer string = "eunomia-finalizer"

// PushEvents channel on which we get the github webhook push events
var PushEvents = make(chan event.GenericEvent)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new GitOpsConfig Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// NewGitOpsReconciler creates a new git ops reconciler
func NewGitOpsReconciler(mgr manager.Manager) ReconcileGitOpsConfig {
	return ReconcileGitOpsConfig{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGitOpsConfig{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("gitopsconfig-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GitOpsConfig
	err = c.Watch(&source.Kind{Type: &gitopsv1alpha1.GitOpsConfig{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(
		&source.Channel{Source: PushEvents},
		&handler.EnqueueRequestForObject{},
	)
	if err != nil {
		return err
	}
	return nil

}

var _ reconcile.Reconciler = &ReconcileGitOpsConfig{}

// ReconcileGitOpsConfig reconciles a GitOpsConfig object
type ReconcileGitOpsConfig struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a GitOpsConfig object and makes changes based on the state read
// and what is in the GitOpsConfig.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGitOpsConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
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
		return reconcile.Result{}, err
	}
	reqLogger.Info("found instance", "instance", instance.GetName())

	//object is being deleted
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.manageDeletion(instance)
	}

	if _, ok := instance.GetAnnotations()[initLabel]; !ok {
		reqLogger.Info("Instance needs to be initialized", "instance", instance.GetName())
		return r.initializeGitOpsConfig(instance)
	}

	reqLogger.Info("Instance is initialized", "instance", instance.GetName())

	if ContainsTrigger(instance, "Periodic") {
		reqLogger.Info("Instance has a periodic trigger, creating/updating cronjob", "instance", instance.GetName())
		_, err = r.createCronJob(instance)
		if err != nil {
			reqLogger.Error(err, "error creating the cronjob, continuing...")
		}
	}

	if ContainsTrigger(instance, "Change") || ContainsTrigger(instance, "Webhook") {
		reqLogger.Info("Instance has a change or Webhook trigger, creating job", "instance", instance.GetName())
		_, err = r.CreateJob("create", instance)
		if err != nil {
			reqLogger.Error(err, "error creating the job, continuing...")
		}
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

// CreateJob creates a new gitops job for the passed instance
func (r *ReconcileGitOpsConfig) CreateJob(jobtype string, instance *gitopsv1alpha1.GitOpsConfig) (reconcile.Result, error) {
	//TODO add logic to ignore if another job was created sooner than x (5 minutes?) time and it is still running.
	mergedata := util.JobMergeData{
		Config: *instance,
		Action: jobtype,
	}
	job, err := util.CreateJob(mergedata)
	if err != nil {
		log.Error(err, "unable to create job manifest from merge data", "mergedata", mergedata)
		return reconcile.Result{}, err
	}
	err = controllerutil.SetControllerReference(instance, &job, r.scheme)
	if err != nil {
		log.Error(err, "unable to the owner for job", "job", job)
		return reconcile.Result{}, err
	}

	log.Info("Creating a new Job", "job.Namespace", job.Namespace, "job.Name", job.Name)
	err = r.client.Create(context.TODO(), &job)
	if err != nil {
		log.Error(err, "unable to create the job", "job", job)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileGitOpsConfig) createCronJob(instance *gitopsv1alpha1.GitOpsConfig) (reconcile.Result, error) {
	mergedata := util.JobMergeData{
		Config: *instance,
		Action: "create",
	}

	var update bool

	cronjob, err := util.CreateCronJob(mergedata)
	if err != nil {
		log.Error(err, "unable to create cronjob manifest from merge data", "mergedata", mergedata)
		return reconcile.Result{}, err
	}

	pCronjob := batchv1beta1.CronJob{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: cronjob.GetName(), Namespace: cronjob.GetNamespace()}, &pCronjob)
	if err != nil {
		if errors.IsNotFound(err) {
			update = false
		} else {
			// Error reading the object - requeue the request.
			return reconcile.Result{}, err
		}
	} else {
		update = true
	}

	err = controllerutil.SetControllerReference(instance, &cronjob, r.scheme)
	if err != nil {
		log.Error(err, "unable to the owner for cronjob", "cronjob", cronjob)
		return reconcile.Result{}, err
	}
	log.Info("Creating/updating CronJob", "cronjob.Namespace", cronjob.Namespace, "cronjob.Name", cronjob.Name)
	if update {
		err = r.client.Update(context.TODO(), &cronjob)
	} else {
		err = r.client.Create(context.TODO(), &cronjob)
	}

	if err != nil {
		log.Error(err, "unable to create/update the cronjob", "cronjob", cronjob)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

// GetAllGitOpsConfig retrieves all the gitops config in the cluster
func (r *ReconcileGitOpsConfig) GetAllGitOpsConfig() (gitopsv1alpha1.GitOpsConfigList, error) {
	instanceList := &gitopsv1alpha1.GitOpsConfigList{}
	err := r.client.List(context.TODO(), &client.ListOptions{}, instanceList)
	if err != nil {
		log.Error(err, "unable to get the list of GitOpsCionfig")
		return *instanceList, err
	}
	return *instanceList, nil
}

func (r *ReconcileGitOpsConfig) initializeGitOpsConfig(instance *gitopsv1alpha1.GitOpsConfig) (reconcile.Result, error) {
	// verify mandatory field exist and set defaults
	if instance.Spec.TemplateSource.URI == "" {
		//TODO set wrong status
		return reconcile.Result{}, goerrors.New("template source URI cannot be empty")
	}

	if instance.Spec.TemplateSource.Ref == "" {
		instance.Spec.TemplateSource.Ref = "master"
	}

	if instance.Spec.TemplateSource.ContextDir == "" {
		instance.Spec.TemplateSource.ContextDir = "."
	}

	if instance.Spec.ParameterSource.URI == "" {
		instance.Spec.ParameterSource.URI = instance.Spec.TemplateSource.URI
	}
	if instance.Spec.ParameterSource.Ref == "" {
		instance.Spec.ParameterSource.Ref = "master"
	}

	if instance.Spec.ParameterSource.ContextDir == "" {
		instance.Spec.ParameterSource.ContextDir = "."
	}

	if instance.Spec.ServiceAccountRef == "" {
		instance.Spec.ServiceAccountRef = "default"
	}

	if instance.Spec.ResourceHandlingMode == "" {
		instance.Spec.ResourceHandlingMode = "CreateOrMerge"
	}

	if instance.Spec.ResourceDeletionMode == "" {
		instance.Spec.ResourceDeletionMode = "Delete"
	}

	if instance.Spec.Namespace == "" {
		instance.Spec.Namespace = instance.GetNamespace()
	}

	instance.ObjectMeta.Annotations[initLabel] = "true"

	if !containsString(instance.ObjectMeta.Finalizers, kubeGitopsFinalizer) && instance.Spec.ResourceDeletionMode != "Retain" {
		instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, kubeGitopsFinalizer)
	}

	err := r.client.Update(context.TODO(), instance)
	if err != nil {
		log.Error(err, "unable to update initialized GitOpsCionfig", "instance", instance)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
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

func (r *ReconcileGitOpsConfig) manageDeletion(instance *gitopsv1alpha1.GitOpsConfig) (reconcile.Result, error) {
	log.Info("Instance is being deleted", "instance", instance.GetName())
	if containsString(instance.ObjectMeta.Finalizers, kubeGitopsFinalizer) {
		// we need to lookup the delete job and if it doesn't exist we launch it, then we see if it is completed successfully if yes we remove the finalizers, if no we return.
		jobList := &batchv1.JobList{}
		selector, err := labels.Parse("action=delete")
		if err != nil {
			log.Error(err, "unable to parse label selector 'action=delete' ")
			return reconcile.Result{}, err
		}
		// looking up all delete jobs
		err = r.client.List(context.TODO(), &client.ListOptions{
			Namespace:     instance.GetNamespace(),
			LabelSelector: selector,
		}, jobList)
		if err != nil {
			log.Error(err, "unable to list jobs ")
			return reconcile.Result{}, err
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
				log.Error(err, "unable to lookup instance's namespace")
				return reconcile.Result{}, err
			}
			if !ns.ObjectMeta.DeletionTimestamp.IsZero() {
				//namespace is being deleted
				// the best we can do in this situation is to let the instance be deleted and hope that this instance was creating objects only in this namespace
				instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, kubeGitopsFinalizer)
				if err := r.client.Update(context.TODO(), instance); err != nil {
					log.Error(err, "unable to create update instace to remove finalizers")
					return reconcile.Result{}, err
				}
				return reconcile.Result{}, nil
			}
			log.Info("Launching delete job for instance", "instance", instance.GetName())
			_, err = r.CreateJob("delete", instance)
			if err != nil {
				log.Error(err, "unable to create deletion job")
				return reconcile.Result{}, err
			}
			//we return because we need to wait for the job to stop
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Minute,
			}, nil
		}
		//There should be only one pending job
		job := applicableJobList[0]
		if job.Status.Succeeded > 0 {
			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, kubeGitopsFinalizer)
			if err := r.client.Update(context.TODO(), instance); err != nil {
				log.Error(err, "unable to create update instace to remove finalizers")
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
		//if it's not succeeded we wait for 1 minute
		//TODO add logic to stop at a certain point ... or not ...
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Minute,
		}, nil
	}
	return reconcile.Result{}, nil
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
