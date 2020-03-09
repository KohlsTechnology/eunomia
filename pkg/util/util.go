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

package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	"github.com/dchest/uniuri"
	"github.com/ghodss/yaml"
	batch "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var jobTemplate *template.Template
var cronJobTemplate *template.Template
var log = logf.Log.WithName("util")

// JobMergeData is the structs that will be used to merge with the job template
type JobMergeData struct {
	Config v1alpha1.GitOpsConfig `json:"config,omitempty"`

	// Action can be create, delete
	Action string `json:"action,omitempty"`
}

// InitializeTemplates initializes the templates needed by this controller, it must be called at controller boot time
func InitializeTemplates(jobTemplateFileName string, cronJobTemplateFileName string) error {
	text, err := ioutil.ReadFile(jobTemplateFileName)
	if err != nil {
		log.Error(err, "error reading job template file", "filename", jobTemplateFileName)
		return fmt.Errorf("error reading job template file %q: %w", jobTemplateFileName, err)
	}
	jobTemplate = template.New("Job").Funcs(template.FuncMap{
		"getID": func() string {
			return uniuri.NewLenChars(6, []byte("abcdefghijklmnopqrstuvwxyz0123456789"))
		},
	})

	jobTemplate, err = jobTemplate.Parse(string(text))
	if err != nil {
		log.Error(err, "error parsing template", "template", text)
		return fmt.Errorf("error parsing template: %w", err)
	}

	text, err = ioutil.ReadFile(cronJobTemplateFileName)
	if err != nil {
		log.Error(err, "error reading cron job template file", "filename", cronJobTemplateFileName)
		return fmt.Errorf("error reading cron job template file %q: %w", cronJobTemplateFileName, err)
	}
	cronJobTemplate = template.New("Job").Funcs(template.FuncMap{
		"getCron": func(config v1alpha1.GitOpsConfig) string {
			for _, trigger := range config.Spec.Triggers {
				if trigger.Type == "Periodic" {
					return trigger.Cron
				}
			}
			return ""
		},
	})

	cronJobTemplate, err = cronJobTemplate.Parse(string(text))
	if err != nil {
		log.Error(err, "error parsing cron job template", "template", text)
		return fmt.Errorf("error parsing cron job template: %w", err)
	}
	return nil
}

// CreateJob returns a Job type from a template merge data
func CreateJob(jobmergedata JobMergeData) (batch.Job, error) {
	job := batch.Job{}
	var b bytes.Buffer
	err := jobTemplate.Execute(&b, &jobmergedata)
	if err != nil {
		log.Error(err, "error executing job template from a template merge data")
		return job, fmt.Errorf("error executing job template from a template merge data: %w", err)
	}
	err = yaml.Unmarshal(b.Bytes(), &job)
	if err != nil {
		log.Error(err, "error unmarshalling a job manifest", "manifest", string(b.Bytes()))
		return job, fmt.Errorf("error unmarshalling a job manifest: %w", err)
	}
	return job, nil
}

// CreateCronJob returns a Job type from a template merge data
func CreateCronJob(jobmergedata JobMergeData) (batchv1beta1.CronJob, error) {
	cronjob := batchv1beta1.CronJob{}
	var b bytes.Buffer
	err := cronJobTemplate.Execute(&b, &jobmergedata)
	if err != nil {
		log.Error(err, "error executing cron job template from a template merge data")
		return cronjob, fmt.Errorf("error executing cron job template from a template merge data: %w", err)
	}
	err = yaml.Unmarshal(b.Bytes(), &cronjob)
	if err != nil {
		log.Error(err, "error unmarshalling a cron job manifest", "manifest", string(b.Bytes()))
		return cronjob, fmt.Errorf("error unmarshalling a cron job manifest: %w", err)
	}
	return cronjob, nil
}
