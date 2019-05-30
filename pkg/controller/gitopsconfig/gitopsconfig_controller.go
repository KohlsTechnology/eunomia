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

	"text/template"
	"time"

	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	opsutil "github.com/redhat-cop/operator-utils/pkg/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
var jobTemplate *template.Template
var cronJobTemplate *template.Template

const initLabel string = "gitopsconfig.eunomia.kohls.io/initialized"
const kubeGitopsFinalizer string = "eunomia-finalizer"
const controllerName = "gitopsconfig_controller"

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
	return ReconcileGitOpsConfig{
		ReconcilerBase: opsutil.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetRecorder(controllerName)),
	}
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGitOpsConfig{
		ReconcilerBase: opsutil.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetRecorder(controllerName)),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("gitopsconfig-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GitOpsConfig
	err = c.Watch(&source.Kind{Type: &gitopsv1alpha1.GitOpsConfig{}}, &handler.EnqueueRequestForObject{}, opsutil.ResourceGenerationOrFinalizerChangedPredicate{})
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
	opsutil.ReconcilerBase
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
	err := r.GetClient().Get(context.TODO(), request.NamespacedName, instance)
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

	// check if CR is valid
	if ok, err := r.IsValid(instance); !ok {
		return r.ManageError(instance, err)
	}

	// check if CR is initialized
	if ok := r.IsInitialized(instance); !ok {
		err := r.GetClient().Update(context.TODO(), instance)
		if err != nil {
			log.Error(err, "unable to update instance", "instance", instance)
			return r.ManageError(instance, err)
		}
		return reconcile.Result{}, nil
	}

	//object is being deleted
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.manageDeletion(instance)
	}

	reqLogger.Info("Instance is initialized", "instance", instance.GetName())

	if ContainsTrigger(instance, "Periodic") {
		reqLogger.Info("Instance has a periodic trigger, creating/updating cronjob", "instance", instance.GetName())
		err = r.createCronJob(instance)
		if err != nil {
			r.ManageError(instance, err)
		}
	}

	if ContainsTrigger(instance, "Change") || ContainsTrigger(instance, "Webhook") {
		reqLogger.Info("Instance has a change or Webhook trigger, creating job", "instance", instance.GetName())
		err = r.CreateJob("create", instance)
		if err != nil {
			r.ManageError(instance, err)
		}
	}

	return r.ManageSuccess(instance)
}

// CreateJob creates a new gitops job for the passed instance
func (r *ReconcileGitOpsConfig) CreateJob(jobtype string, instance *gitopsv1alpha1.GitOpsConfig) error {
	//TODO add logic to ignore if another job was created sooner than x (5 minutes?) time and it is still running.
	mergedata := JobMergeData{
		Config: *instance,
		Action: jobtype,
	}
	job, err := opsutil.ProcessTemplate(mergedata, jobTemplate)
	if err != nil {
		log.Error(err, "unable to create job manifest from merge data", "mergedata", mergedata)
		return err
	}
	err = controllerutil.SetControllerReference(instance, job, r.GetScheme())
	if err != nil {
		log.Error(err, "unable to the owner for job", "job", job)
		return err
	}

	log.Info("Creating a new Job", "job.Namespace", job.GetNamespace(), "job.Name", job.GetName())
	err = r.GetClient().Create(context.TODO(), job)
	if err != nil {
		log.Error(err, "unable to create the job", "job", job)
		return err
	}
	return nil
}

func (r *ReconcileGitOpsConfig) createCronJob(instance *gitopsv1alpha1.GitOpsConfig) error {
	mergedata := JobMergeData{
		Config: *instance,
		Action: "create",
	}

	cronjob, err := opsutil.ProcessTemplate(mergedata, cronJobTemplate)
	if err != nil {
		log.Error(err, "unable to create cronjob manifest from merge data", "mergedata", mergedata)
		return err
	}
	err = r.CreateOrUpdateResource(instance, "", cronjob)

	if err != nil {
		log.Error(err, "unable to create/update the cronjob", "cronjob", cronjob)
		return err
	}
	return nil
}

// GetAllGitOpsConfig retrieves all the gitops config in the cluster
func (r *ReconcileGitOpsConfig) GetAllGitOpsConfig() (gitopsv1alpha1.GitOpsConfigList, error) {
	instanceList := &gitopsv1alpha1.GitOpsConfigList{}
	err := r.GetClient().List(context.TODO(), &client.ListOptions{}, instanceList)
	if err != nil {
		log.Error(err, "unable to get the list of GitOpsCionfig")
		return *instanceList, err
	}
	return *instanceList, nil
}

func (r *ReconcileGitOpsConfig) manageDeletion(instance *gitopsv1alpha1.GitOpsConfig) (reconcile.Result, error) {
	log.Info("Instance is being deleted", "instance", instance.GetName())
	if opsutil.HasFinalizer(instance, kubeGitopsFinalizer) {
		// we need to lookup the delete job and if it doesn't exist we launch it, then we see if it is completed successfully if yes we remove the finalizers, if no we return.
		jobList := &batchv1.JobList{}
		selector, err := labels.Parse("action=delete")
		if err != nil {
			log.Error(err, "unable to parse label selector 'action=delete' ")
			return reconcile.Result{}, err
		}
		// looking up all delete jobs
		err = r.GetClient().List(context.TODO(), &client.ListOptions{
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
			if opsutil.IsOwner(instance, &job) && job.GetLabels()["action"] == "delete" {
				applicableJobList = append(applicableJobList, job)
			}
		}
		if len(applicableJobList) == 0 {
			// to avoid a deadlock situation let's check that the namespace in which we are is not being deleted
			ns := &corev1.Namespace{}
			err := r.GetClient().Get(context.TODO(), types.NamespacedName{
				Name: instance.GetNamespace(),
			}, ns)
			if err != nil {
				log.Error(err, "unable to lookup instance's namespace")
				return reconcile.Result{}, err
			}
			if !ns.ObjectMeta.DeletionTimestamp.IsZero() {
				//namespace is being deleted
				// the best we can do in this situation is to let the instance be deleted and hope that this instance was creating objects only in this namespace
				opsutil.RemoveFinalizer(instance, kubeGitopsFinalizer)
				if err := r.GetClient().Update(context.TODO(), instance); err != nil {
					log.Error(err, "unable to create update instace to remove finalizers")
					return reconcile.Result{}, err
				}
				return reconcile.Result{}, nil
			}
			log.Info("Launching delete job for instance", "instance", instance.GetName())
			err = r.CreateJob("delete", instance)
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
			opsutil.RemoveFinalizer(instance, kubeGitopsFinalizer)
			if err := r.GetClient().Update(context.TODO(), instance); err != nil {
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

// IsValid returns true if the obj is balid and false with an error describing the nature of the violation if not.
func (r *ReconcileGitOpsConfig) IsValid(obj metav1.Object) (bool, error) {
	instance, ok := obj.(*gitopsv1alpha1.GitOpsConfig)
	if !ok {
		return false, goerrors.New("instance is not of gitopsv1alpha1.GitOpsConfig type")
	}
	if instance.Spec.TemplateSource.URI == "" {
		//TODO set wrong status
		return false, goerrors.New("template source URI cannot be empty")
	}
	return true, nil
}

// IsInitialized determines whether obj is initialzied. if ont it returns false and obj will be changed with initilzied values, so that it can be used to update the instance.
func (r *ReconcileGitOpsConfig) IsInitialized(obj metav1.Object) bool {
	instance, ok := obj.(*gitopsv1alpha1.GitOpsConfig)
	if !ok {
		return false
	}
	var initialized = true
	if instance.Spec.TemplateSource.Ref == "" {
		instance.Spec.TemplateSource.Ref = "master"
		initialized = false
	}

	if instance.Spec.TemplateSource.ContextDir == "" {
		instance.Spec.TemplateSource.ContextDir = "."
		initialized = false
	}

	if instance.Spec.ParameterSource.URI == "" {
		instance.Spec.ParameterSource.URI = instance.Spec.TemplateSource.URI
		initialized = false
	}
	if instance.Spec.ParameterSource.Ref == "" {
		instance.Spec.ParameterSource.Ref = "master"
		initialized = false
	}

	if instance.Spec.ParameterSource.ContextDir == "" {
		instance.Spec.ParameterSource.ContextDir = "."
		initialized = false
	}

	if instance.Spec.ServiceAccountRef == "" {
		instance.Spec.ServiceAccountRef = "default"
		initialized = false
	}

	if instance.Spec.ResourceHandlingMode == "" {
		instance.Spec.ResourceHandlingMode = "CreateOrMerge"
		initialized = false
	}

	if instance.Spec.ResourceDeletionMode == "" {
		instance.Spec.ResourceDeletionMode = "Delete"
		initialized = false
	}

	if !opsutil.HasFinalizer(instance, kubeGitopsFinalizer) && instance.Spec.ResourceDeletionMode != "Retain" {
		opsutil.AddFinalizer(instance, kubeGitopsFinalizer)
		initialized = false
	}

	return initialized
}
