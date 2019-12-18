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
	"sync"
	"time"

	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	"gopkg.in/robfig/cron.v2"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type jobmonitor struct {
	client       client.Client
	Name         string
	Namespace    string
	NextSchedule time.Time
}

//constant instead of using string
const (
	inprogress = "INPROGRESS ..."
	success    = "SUCCESS ..."
	failed     = "FAILED ..."
)

//sync.Map is used inorder to introduce syncronization for reading and writing the data
var runningJobMap sync.Map

func scheduleStatusForCronJobs(job jobmonitor, schedule string, instanceName string, instanceNamespace string) {
	result, err := job.getCronjobmonitor()

	if err != nil {
		log.Error(err, "Error in GetCronJob")
		return
	}

	//If reconciler calls the same cronjob before the final update of status, we might have an error while
	//updating the status. To avoid this error runningJobMap is used.
	if _, ok := runningJobMap.Load(job.Name); ok {
		log.Info("Job name in map " + job.Name)
		return
	}

	runningJobMap.Store(job.Name, true)
	if len(result.Status.Active) == 0 {
		log.Info("There is no active job")

		//This is to indicate that there is no active job before the cron is scheduled
		err = job.updateStatus(instanceName, instanceNamespace, gitopsv1alpha1.GitOpsConfigStatus{
			LastScheduleTime: result.Status.LastScheduleTime,
			Message:          "There is no active job",
		})

		if err != nil {
			log.Error(err, "Error in Updating the status for cron job")
		}
	}
	// TODO: convert from cron to an approach based on Watch, see "Creating watches" in: https://blog.openshift.com/kubernetes-operators-best-practices/
	c := cron.New()
	c.AddFunc(schedule, func() { statusForCronJobs(job, instanceName, instanceNamespace, c) })
	c.Start()
}

func statusForCronJobs(job jobmonitor, instanceName string, instanceNamespace string, c *cron.Cron) {
	//Some time is needed for the job to get created. Hence the sleep for 10 seconds.
	time.Sleep(time.Second * 10)
	result, err := job.getCronjobmonitor()

	if err != nil {
		log.Error(err, "Error in GetCronJob")
		c.Stop()
		return
	}

	length := len(result.Status.Active)
	if length == 0 {
		log.Info("Cron job scheduled, but there is no active job...")

		err = job.updateStatus(instanceName, instanceNamespace, gitopsv1alpha1.GitOpsConfigStatus{
			LastScheduleTime: result.Status.LastScheduleTime,
			Message:          "Cron job scheduled, but there is no active job...",
		})
		if err != nil {
			log.Error(err, "Error in Updating the status for cron job")
		}
		return
	}
	latestJobRunning := result.Status.Active[length-1]
	var nextSchedule time.Time

	// To fetch the next schedule time, so that once the new schedule has started the old one must be stopped
	entries := c.Entries()
	if len(entries) != 0 {
		nextSchedule = entries[0].Next
	}

	if _, ok := runningJobMap.Load(latestJobRunning.Name); ok {
		return
	}

	go func(cronjobNameInMap string) {
		watchJobForStatus(job.client, latestJobRunning.Name, latestJobRunning.Namespace, instanceName, instanceNamespace, nextSchedule)
		runningJobMap.Delete(cronjobNameInMap)
	}(job.Name)
}

//Creating the jobmonitor in order to watch the job.
func watchJobForStatus(client client.Client, jobName string, namespace string, instanceName string, instanceNamespace string, nextSchedule time.Time) {
	//If reconciler calls the same job before the final update of status, we might have an error while
	//updating the status. To avoid this error runningJobMap is used.
	if _, ok := runningJobMap.Load(jobName); ok {
		log.Info("Returning from watchJobForStatus for job " + jobName)
		return
	}

	runningJobMap.Store(jobName, true)
	var job = jobmonitor{
		client:       client,
		Name:         jobName,
		Namespace:    namespace,
		NextSchedule: nextSchedule,
	}
	job.watchjob(instanceName, instanceNamespace)
	runningJobMap.Delete(jobName)
}

//Watching the job for the get the appropriate status.
func (job *jobmonitor) watchjob(instanceName, instanceNamespace string) {
	log.Info("Watching the Job: " + job.Name)
	var initialized bool
	for {
		//If the cron job schedule time is less than 10 seconds, then we might miss the status update
		if !job.NextSchedule.IsZero() && time.Now().After(job.NextSchedule) {
			log.Info("Breaking out of infinite loop")
			return
		}
		time.Sleep(10 * time.Second)
		jobState, err := job.checkResult()
		if err != nil {
			// FIXME: change below error message detection to something more robust, string matching is super flaky
			if err.Error() == `Job.batch "`+instanceName+`" not found` {
				log.Info("The job has been deleted")
				break
			}
			continue
		}

		if jobState.State != inprogress {
			err := job.updateStatus(instanceName, instanceNamespace, jobState)
			if err != nil {
				log.Error(err, "Error in Updating the job status")
			}
			return
		}

		if !initialized {
			initialized = true
			log.Info("Initial Update")
			err := job.updateStatus(instanceName, instanceNamespace, jobState)
			if err != nil {
				log.Error(err, "Error in Initial Update of job Status:")
			}
		}

	}
}

//Checking the Job to get the actual state of the job.
func (job *jobmonitor) checkResult() (jobState gitopsv1alpha1.GitOpsConfigStatus, err error) {
	jobOutput, err := job.getjobmonitor()
	if err != nil {
		time.Sleep(time.Second * 10)
		log.Info("retrying the GetJob for CheckResult")
		jobOutput, err = job.getjobmonitor()
		if err != nil {
			log.Info("Retry of GetJob for CheckResult unsuccessful: " + err.Error())
			return jobState, err
		}
	}

	jobState.State = getJobStatus(jobOutput.Status)
	if jobOutput.Status.Active > 0 && jobOutput.Status.Failed > 0 {
		jobState.Message = "job failed, retrying..."
	}
	jobState.StartTime = jobOutput.Status.StartTime
	jobState.CompletionTime = jobOutput.Status.CompletionTime

	return jobState, nil
}

//To get the job
func (job *jobmonitor) getjobmonitor() (batchv1.Job, error) {
	var jobOutput batchv1.Job
	err := job.client.Get(context.TODO(),
		types.NamespacedName{
			Name:      job.Name,
			Namespace: job.Namespace,
		},
		&jobOutput)
	return jobOutput, err
}

//To get the cron job
func (job *jobmonitor) getCronjobmonitor() (batchv1beta1.CronJob, error) {
	var result batchv1beta1.CronJob
	err := job.client.Get(context.TODO(),
		types.NamespacedName{Name: job.Name, Namespace: job.Namespace},
		&result)
	return result, err
}

//Updating the instance of the job/cronjob
//Pass the instanceName and instanceNamespace from the instance and not from job
func (job *jobmonitor) updateStatus(instanceName string, instanceNamespace string, jobState gitopsv1alpha1.GitOpsConfigStatus) error {
	var newinstance = new(gitopsv1alpha1.GitOpsConfig)

	//Fetching the latest version of instance.
	err := job.client.Get(context.TODO(), types.NamespacedName{Name: instanceName, Namespace: instanceNamespace}, newinstance)
	if err != nil {
		log.Error(err, "Error in gettting the instance")
		return err
	}
	newinstance.Status = jobState
	err = job.client.Status().Update(context.TODO(), newinstance)
	if err != nil {
		log.Error(err, "Failed to update status", "job", job.Name)
		return err
	}

	log.Info("Status updated as " + jobState.State + " for job : " + job.Name)
	return nil
}

//Logic to define the status
func getJobStatus(jobStatus batchv1.JobStatus) string {
	var result = inprogress
	switch {
	case jobStatus.Active > 0:
		result = inprogress
	case jobStatus.Succeeded != 0:
		result = success
	case jobStatus.Failed != 0:
		result = failed
	}

	return result

}
