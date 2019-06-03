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
	"io/ioutil"
	"testing"
	"text/template"

	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	"github.com/dchest/uniuri"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var fullconfig = JobMergeData{
	Action: "create",
	Config: gitopsv1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GitOpsConfig",
			APIVersion: "eunomia.kohls.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops",
			Namespace: "gitops-operator",
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
			ResourceHandlingMode:   "CreateOrMerge",
		},
	},
}

const templateFile string = "../../templates/job.yaml"

func TestFullConfig(t *testing.T) {
	text, err := ioutil.ReadFile(templateFile)
	if err != nil {
		t.Errorf("Error reading template file: %v", err)
		return
	}
	template := template.New("Job").Funcs(template.FuncMap{
		"getID": func() string {
			return uniuri.NewLenChars(6, []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"))
		},
	})

	template, err = template.Parse(string(text))
	if err != nil {
		t.Errorf("Error parsing template: %v", err)
		return
	}
	var b bytes.Buffer
	err = template.Execute(&b, &fullconfig)
	if err != nil {
		t.Errorf("Error executing template: %v", err)
		return
	}

	t.Logf("resulting manifest: %v", b.String())
}
