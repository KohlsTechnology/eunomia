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

package test

import (
	"time"

	eventv1beta1 "k8s.io/api/events/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// WatchEvents starts a Kubernetes watcher, collecting any events emitted in
// the provided namespace, filtering only the ones regarding the resource with
// the provided name. The watch is expected to finish after specified timeout
// (NOTE: rounded down to seconds). Returns a function that, when called, stops
// the watch and frees its associated resources.
func WatchEvents(client kubernetes.Interface, events chan<- *eventv1beta1.Event, namespace, name string, timeout time.Duration) (closer func(), err error) {
	timeoutSeconds := int64(timeout / time.Second)
	watcher, err := client.EventsV1beta1().Events(namespace).Watch(metav1.ListOptions{
		TimeoutSeconds: &timeoutSeconds,
	})
	if err != nil {
		return nil, err
	}
	go func() { // based on: https://stackoverflow.com/a/54930836
		ch := watcher.ResultChan()
		for { //nolint:gosimple
			select {
			case change, ok := <-ch:
				if !ok {
					// Channel closed, finish watching.
					close(events)
					return
				}
				if change.Type != watch.Added {
					continue
				}
				event, ok := change.Object.(*eventv1beta1.Event)
				if !ok || event.Regarding.Name != name {
					continue
				}
				events <- event
			}
		}
	}()
	return func() { watcher.Stop() }, nil
}
