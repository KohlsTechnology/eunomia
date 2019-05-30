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
	"io/ioutil"
	"text/template"

	"github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	"github.com/dchest/uniuri"
)

// JobMergeData is the structs that will be used to merge with the job template
type JobMergeData struct {
	Config v1alpha1.GitOpsConfig `json:"config,omitempty"`

	// Action can be create, delete
	Action string `json:"action,omitempty"`
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

// InitializeTemplates initializes the temolates needed by this controller, it must be called at controller boot time
func InitializeTemplates(jobTempateFileName string, cronJobTemplateFilename string) error {
	text, err := ioutil.ReadFile(jobTempateFileName)
	if err != nil {
		log.Error(err, "Error reading job template file", "filename", jobTempateFileName)
		return err
	}
	jobTemplate = template.New("Job").Funcs(template.FuncMap{
		"getID": func() string {
			return uniuri.NewLenChars(6, []byte("abcdefghijklmnopqrstuvwxyz0123456789"))
		},
	})

	jobTemplate, err = jobTemplate.Parse(string(text))
	if err != nil {
		log.Error(err, "Error parsing template", "template", text)
		return err
	}

	text, err = ioutil.ReadFile(cronJobTemplateFilename)
	if err != nil {
		log.Error(err, "Error reading job template file", "filename", cronJobTemplateFilename)
		return err
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
		log.Error(err, "Error parsing template", "template", text)
		return err
	}
	return nil
}
