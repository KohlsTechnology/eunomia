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

package handler

import (
	"io/ioutil"
	"net/http"
	"strings"

	gitopsv1alpha1 "github.com/KohlsTechnology/eunomia/pkg/apis/eunomia/v1alpha1"
	"github.com/KohlsTechnology/eunomia/pkg/controller/gitopsconfig"
	"github.com/google/go-github/github"
	k8sevent "sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("handler")

// WebhookHandler manages the calls from github
func WebhookHandler(w http.ResponseWriter, r *http.Request, reconciler gitopsconfig.Reconciler) {
	log.Info("received webhook call")
	if r.Method != "POST" {
		log.Info("webhook handler only accepts the POST method", "sent_method", r.Method)
		w.WriteHeader(405)
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err, "error reading request body")
		return
	}
	defer r.Body.Close()

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Error(err, "error parsing webhook event payload")
		return
	}

	switch e := event.(type) {
	case *github.PushEvent:
		// A commit push was received, determine if there is are GitOpsConfigs that match the event
		// The repository url and Git ref must match for the templateSource or parameterSource
		{
			list, err := reconciler.GetAll()
			if err != nil {
				log.Error(err, "error getting the list of GitOpsConfigs")
				w.WriteHeader(500)
				return
			}

			targetList := gitopsv1alpha1.GitOpsConfigList{
				TypeMeta: list.TypeMeta,
				ListMeta: list.ListMeta,
				Items:    make([]gitopsv1alpha1.GitOpsConfig, 0, len(list.Items)),
			}

			for _, instance := range list.Items {
				if !gitopsconfig.ContainsTrigger(&instance, "Webhook") {
					log.Info("skip instance without webhook trigger", "instance_name", instance.Name)
					continue
				}

				log.Info("comparing instance and event metadata", "event_name", e.Repo.GetFullName(), "event_ref", e.GetRef(),
					"template_uri", instance.Spec.TemplateSource.URI, "template_ref", instance.Spec.TemplateSource.Ref,
					"parameter_uri", instance.Spec.ParameterSource.URI, "parameter_ref", instance.Spec.ParameterSource.Ref)

				if !repoURLAndRefMatch(&instance, e) {
					log.Info("skip instance without matching repo url or git ref of the event", "instance_name", instance.Name)
					continue
				}

				log.Info("found matching instance", "instance_name", instance.Name)
				targetList.Items = append(targetList.Items, instance)
			}

			if len(targetList.Items) == 0 {
				log.Info("no gitopsconfigs match the webhook event", "event_repo", e.Repo.GetFullName(), "event_ref", strings.TrimPrefix(e.GetRef(), "refs/heads/"))
				return
			}

			log.Info("event is applicable to the following instances", "matching_instance_count", len(targetList.Items), "matching_instances", targetList.Items)

			for _, instance := range targetList.Items {
				//if secured discard those that do not validate
				//log.Info("managing instance", "instances", instance)
				secret := getWebhookSecret(&instance)
				if secret != "" {
					//log.Info("validating payload instance")
					_, err := github.ValidatePayload(r, []byte(secret))
					if err != nil {
						log.Error(err, "webhook payload could not be validated with instance secret, ignoring this instance")
						continue
					}
				}
				//log.Info("payload validated")
				//log.Info("creating job")
				gitopsconfig.PushEvents <- k8sevent.GenericEvent{
					Meta:   instance.GetObjectMeta(),
					Object: instance.DeepCopyObject(),
				}

				// _, err := reconciler.CreateJob("create", &instance)
				// if err != nil {
				// 	log.Error(err, "unable to create job for instance", "instance", instance)
				// }
			}
		}
	default:
		{
			log.Info("unknown event type", "type", github.WebHookType(r))
			return
		}
	}
	log.Info("webhook handling concluded correctly")
}

func repoURLAndRefMatch(instance *gitopsv1alpha1.GitOpsConfig, event *github.PushEvent) bool {
	return event.Repo != nil && event.Repo.FullName != nil && event.Ref != nil &&
		((strings.Contains(instance.Spec.TemplateSource.URI, *event.Repo.FullName) &&
			instance.Spec.TemplateSource.Ref == strings.TrimPrefix(*event.Ref, "refs/heads/")) ||
			(strings.Contains(instance.Spec.ParameterSource.URI, *event.Repo.FullName) &&
				instance.Spec.ParameterSource.Ref == strings.TrimPrefix(*event.Ref, "refs/heads/")))
}

func getWebhookSecret(instance *gitopsv1alpha1.GitOpsConfig) string {
	for _, trigger := range instance.Spec.Triggers {
		if trigger.Type == "Webhook" {
			return trigger.Secret
		}
	}
	return ""
}
