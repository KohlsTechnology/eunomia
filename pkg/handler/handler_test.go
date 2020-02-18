/*
Copyright 2020 Kohl's Department Stores, Inc.

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

package handler

import (
	"testing"

	"github.com/google/go-github/github"

	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
)

func newstring(v string) *string { return &v }

func TestRepoURLAndRefMatch(t *testing.T) {

	defaultGit := gitopsv1alpha1.GitConfig{
		URI: "https://github.com/kohlstechnology/eunomia",
		Ref: "master",
	}

	tests := []struct {
		comment string
		spec    gitopsv1alpha1.GitOpsConfigSpec
		event   github.PushEvent
		want    bool
	}{
		{
			comment: "full match - branch",
			spec:    gitopsv1alpha1.GitOpsConfigSpec{TemplateSource: defaultGit, ParameterSource: defaultGit},
			event: github.PushEvent{
				Ref: newstring("master"),
				Repo: &github.PushEventRepository{
					FullName: newstring("kohlstechnology/eunomia"),
				},
			},
			want: true,
		},
		{
			comment: "full match - git tag",
			spec: gitopsv1alpha1.GitOpsConfigSpec{
				TemplateSource: defaultGit,
				ParameterSource: gitopsv1alpha1.GitConfig{
					URI: defaultGit.URI,
					Ref: "refs/tags/0.1.4",
				},
			},
			event: github.PushEvent{
				Ref: newstring("refs/tags/0.1.4"),
				Repo: &github.PushEventRepository{
					FullName: newstring("kohlstechnology/eunomia"),
				},
			},
			want: true,
		},
		{
			comment: "TemplateSource match",
			spec: gitopsv1alpha1.GitOpsConfigSpec{
				TemplateSource: defaultGit,
				ParameterSource: gitopsv1alpha1.GitConfig{
					URI: defaultGit.URI,
					Ref: "master2",
				},
			},
			event: github.PushEvent{
				Ref: newstring("master"),
				Repo: &github.PushEventRepository{
					FullName: newstring("kohlstechnology/eunomia"),
				},
			},
			want: true,
		},
		{
			comment: "ParameterSource match",
			spec: gitopsv1alpha1.GitOpsConfigSpec{
				TemplateSource: gitopsv1alpha1.GitConfig{
					URI: defaultGit.URI,
					Ref: "master2",
				},
				ParameterSource: defaultGit,
			},
			event: github.PushEvent{
				Ref: newstring("master"),
				Repo: &github.PushEventRepository{
					FullName: newstring("kohlstechnology/eunomia"),
				},
			},
			want: true,
		},
		{
			comment: "Ref differs",
			spec: gitopsv1alpha1.GitOpsConfigSpec{
				TemplateSource:  defaultGit,
				ParameterSource: defaultGit,
			},
			event: github.PushEvent{
				Ref: newstring("refs/tags/0.1.4"),
				Repo: &github.PushEventRepository{
					FullName: newstring("kohlstechnology/eunomia"),
				},
			},
			want: false,
		},
		{
			comment: "URI differs",
			spec:    gitopsv1alpha1.GitOpsConfigSpec{TemplateSource: defaultGit, ParameterSource: defaultGit},
			event: github.PushEvent{
				Ref: newstring("master"),
				Repo: &github.PushEventRepository{
					FullName: newstring("kohlstechnology/git2consul-go"),
				},
			},
			want: false,
		},
		{
			comment: "URI and Ref differs",
			spec:    gitopsv1alpha1.GitOpsConfigSpec{TemplateSource: defaultGit, ParameterSource: defaultGit},
			event: github.PushEvent{
				Ref: newstring("master2"),
				Repo: &github.PushEventRepository{
					FullName: newstring("kohlstechnology/git2consul-go"),
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		gitops := &gitopsv1alpha1.GitOpsConfig{Spec: tt.spec}
		pushEvent := &tt.event
		result := repoURLAndRefMatch(gitops, pushEvent)
		if result != tt.want {
			t.Errorf("%q: expected %v, got %v", tt.comment, tt.want, result)
		}
	}
}
