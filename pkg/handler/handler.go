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
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("handler")

// WebhookHandler manages the calls from github
func WebhookHandler(w http.ResponseWriter, r *http.Request, reconciler gitopsconfig.ReconcileGitOpsConfig) {
	log.Info("received webhook call")
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	//log.Info("webhook is of type post")
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err, "error reading request body")
		return
	}
	defer r.Body.Close()
	//log.Info("parsed body")
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Error(err, "could not parse webhook")
		return
	}
	//log.Info("parsed body, found event", "event", event)
	switch e := event.(type) {
	case *github.PushEvent:
		// this is a commit push, do something with it
		{
			//find the list of CR that have this url.
			//log.Info("event is of type push")
			list, err := reconciler.GetAllGitOpsConfig()
			if err != nil {
				log.Error(err, "unable to get the list of GitOpsCionfig")
				w.WriteHeader(500)
				return
			}

			targetList := gitopsv1alpha1.GitOpsConfigList{
				TypeMeta: list.TypeMeta,
				ListMeta: list.ListMeta,
				Items:    make([]gitopsv1alpha1.GitOpsConfig, len(list.Items)),
			}

			for _, instance := range list.Items {
				// if does not have the webhook trigger continue
				if !gitopsconfig.ContainsTrigger(&instance, "Webhook") {
					continue
				}
				// if the repo URL do not correspond continue
				if !repoURLMatch(&instance, e) {
					continue
				}
				targetList.Items = append(targetList.Items, instance)
			}
			//log.Info("event is applicable to the following instances", "instances", targetList)

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

func repoURLMatch(instance *gitopsv1alpha1.GitOpsConfig, event *github.PushEvent) bool {
	return strings.Contains(instance.Spec.TemplateSource.URI, *event.Repo.FullName) || strings.Contains(instance.Spec.ParameterSource.URI, *event.Repo.FullName)
}

func getWebhookSecret(instance *gitopsv1alpha1.GitOpsConfig) string {
	for _, trigger := range instance.Spec.Triggers {
		if trigger.Type == "Webhook" {
			return trigger.Secret
		}
	}
	return ""
}
