// +build !ignore_autogenerated

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

// Code generated by openapi-gen. DO NOT EDIT.

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1.GitOpsConfig":       schema_pkg_apis_eunomia_v1alpha1_GitOpsConfig(ref),
		"github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1.GitOpsConfigSpec":   schema_pkg_apis_eunomia_v1alpha1_GitOpsConfigSpec(ref),
		"github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1.GitOpsConfigStatus": schema_pkg_apis_eunomia_v1alpha1_GitOpsConfigStatus(ref),
	}
}

func schema_pkg_apis_eunomia_v1alpha1_GitOpsConfig(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "GitOpsConfig is the Schema for the gitopsconfigs API",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1.GitOpsConfigSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1.GitOpsConfigStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1.GitOpsConfigSpec", "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1.GitOpsConfigStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_eunomia_v1alpha1_GitOpsConfigSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "GitOpsConfigSpec defines the desired state of GitOpsConfig",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"templateSource": {
						SchemaProps: spec.SchemaProps{
							Description: "TemplateSource is the location of the templated resources",
							Ref:         ref("github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1.GitConfig"),
						},
					},
					"parameterSource": {
						SchemaProps: spec.SchemaProps{
							Description: "ParameterSource is the location of the parameters, only contextDir is mandatory, if other filed are left blank they are assumed to be the same as ParameterSource",
							Ref:         ref("github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1.GitConfig"),
						},
					},
					"triggers": {
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								"x-kubernetes-list-type": "set",
							},
						},
						SchemaProps: spec.SchemaProps{
							Description: "Triggers is an array of triggers that will launch this configuration",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1.GitOpsTrigger"),
									},
								},
							},
						},
					},
					"serviceAccountRef": {
						SchemaProps: spec.SchemaProps{
							Description: "ServiceAccountRef references to the service account under which the template engine job will run, it must exists in the namespace in which this CR is created",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"templateProcessorImage": {
						SchemaProps: spec.SchemaProps{
							Description: "TemplateEngine, the gitops operator config map contains the list of available template engines, the value used here must exist in that list. Identity (i.e. no resource processing) is the default",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"resourceHandlingMode": {
						SchemaProps: spec.SchemaProps{
							Description: "ResourceHandlingMode represents how resource creation/update should be handled. Supported values are Apply,Create,Delete,Patch,Replace,None. Default is Apply.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"resourceDeletionMode": {
						SchemaProps: spec.SchemaProps{
							Description: "ResourceDeletionMode represents how resource deletion should be handled. Supported values are Retain,Delete,None. Default is Delete",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"templateProcessorArgs": {
						SchemaProps: spec.SchemaProps{
							Description: "TemplateProcessorArgs references to the run time parameters, we can pass additional arguments/flags to the template processor.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1.GitConfig", "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1.GitOpsTrigger"},
	}
}

func schema_pkg_apis_eunomia_v1alpha1_GitOpsConfigStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "GitOpsConfigStatus defines the observed state of GitOpsConfig",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"state": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"startTime": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.Time"),
						},
					},
					"completionTime": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.Time"),
						},
					},
					"message": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"lastScheduleTime": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.Time"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"k8s.io/apimachinery/pkg/apis/meta/v1.Time"},
	}
}
