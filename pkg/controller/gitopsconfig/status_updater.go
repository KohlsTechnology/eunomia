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

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
)

// statusUpdater updates Status of GitOpsConfig objects in the cluster when it
// is used as a ResourceEventHandler on batchv1.Job objects. For more details,
// see OnUpdate method. The OnAdd and OnDelete methods pass their argument to
// OnUpdate.
type statusUpdater struct {
	client client.Client
}

var _ cache.ResourceEventHandler = &statusUpdater{}

func (u *statusUpdater) OnAdd(newObj interface{})    { u.OnUpdate(nil, newObj) }
func (u *statusUpdater) OnDelete(oldObj interface{}) { u.OnUpdate(oldObj, nil) }

// OnUpdate finds a GitOpsConfig object owning newObj if it is a batchv1.Job
// (or of oldObj if newObj is nil). Then, it updates the GitOpsConfig's Status
// based on the Status of the Job.
func (u *statusUpdater) OnUpdate(oldObj, newObj interface{}) {
	// Extract Job objects from arguments
	oldJob, ok := oldObj.(*batchv1.Job)
	if !ok && oldObj != nil {
		log.Error(nil, "non-Job object passed to statusUpdater", "oldObj", oldObj, "newObj", newObj)
		return
	}
	newJob, ok := newObj.(*batchv1.Job)
	if !ok && newObj != nil {
		log.Error(nil, "non-Job object passed to statusUpdater", "oldObj", oldObj, "newObj", newObj)
		return
	}

	// In case of job deletion event, newJob is nil. However, we can still work
	// with the oldJob, to make sure that any changes it contained were taken
	// into account in Status.
	if newJob == nil {
		newJob = oldJob
	}
	if newJob.Status.StartTime == nil {
		log.Info("unstarted Job found; cannot properly set GitOpsConfig.Status based on it, ignoring", "job", newJob.Name)
		return
	}

	// Check if this is a Job that's owned by GitOpsConfig.
	gitopsName := ""
	if newJob.Labels != nil {
		gitopsName = newJob.Labels[tagJobOwner]
	}
	if gitopsName == "" {
		// Got an event for a job not owned by GitOpsConfig - ignore it.
		return
	}

	// TODO: This function currently operates on changes, i.e. it's
	// "edge-based". It would be better to use a "level-based" approach, i.e.:
	// 1. query all jobs belonging to the same GitOpsConfig as newJob, and find the newest one among them
	// 2. set status of GitOpsConfig based on status of the newest Job
	// 3. additionally, run this procedure periodically for all GitOpsConfigs,
	//    to ensure any missed changes are eventually caught and reflected in
	//    GitOpsConfig.Status.

	// Calculate status
	status := gitopsv1alpha1.GitOpsConfigStatus{
		StartTime:      newJob.Status.StartTime,
		CompletionTime: newJob.Status.CompletionTime,
	}
	switch {
	case newJob.Status.Active > 0:
		status.State = "InProgress"
	case newJob.Status.Succeeded == 1:
		status.State = "Success"
	case newJob.Status.Succeeded == 0 && newJob.Status.Failed > 0:
		status.State = "Failure"
	}

	// Update status
	gitops := &gitopsv1alpha1.GitOpsConfig{}
	err := u.client.Get(context.TODO(), types.NamespacedName{Name: gitopsName, Namespace: newJob.GetNamespace()}, gitops)
	if err != nil {
		log.Error(err, "cannot update GitOpsConfig")
		return
	}
	// NOTE: OnUpdate calls may come reordered. We must try to ensure that some
	// past Job won't accidentally overwrite a Status set based on a newer Job.
	// This is expected to work correctly when there's at most one Job per
	// GitOpsConfig running at a time (see #179).
	if gitops.Status.StartTime != nil && status.StartTime.Before(gitops.Status.StartTime) {
		log.Info("Status is already set, with newer StartTime - skipping; reordered events?", "GitOpsConfig", gitops.Name)
		return
	}
	// TODO: don't update if status didn't change
	gitops.Status = status
	err = u.client.Status().Update(context.TODO(), gitops)
	if err != nil {
		// FIXME: find a way to retry this, starting from Get above, in case when errors.IsConflict(err)
		log.Error(err, "Failed to update status", "GitOpsConfig", gitops.Name, "job", newJob.Name)
		return
	}
}
