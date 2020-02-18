// +build e2e

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

package e2e

import (
	"bytes"
	goctx "context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/KohlsTechnology/eunomia/pkg/util"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// GetPod retrieves a given pod based on namespace, the pod name prefix, and the image used
// Original Source https://github.com/jaegertracing/jaeger-operator/blob/master/test/e2e/utils.go
func GetPod(namespace, namePrefix, containsImage string, kubeclient kubernetes.Interface) (*v1.Pod, error) {
	pods, err := kubeclient.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("cannot retrieve pods in namespace %q: %w", namespace, err)
	}
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, namePrefix) {
			for _, c := range pod.Spec.Containers {
				fmt.Printf("Found pod %s %q\n", c.Image, pod.Name)
				if strings.Contains(c.Image, containsImage) {
					return &pod, nil
				}
			}
		}
	}
	return nil, nil
}

// WaitForPod retrieves a specific pod with a known name and namespace and waits for it to be running and available
func WaitForPod(t *testing.T, f *framework.Framework, namespace, name string, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		// Check if the CRD has been created
		pod := &v1.Pod{}
		err = f.Client.Get(goctx.TODO(), util.NN{Name: name, Namespace: namespace}, pod)
		switch {
		case apierrors.IsNotFound(err):
			t.Logf("Waiting for availability of %s pod", name)
			return false, nil
		case err != nil:
			return false, xerrors.Errorf("client failed to retrieve pod %q in namespace %q: %w", name, namespace, err)
		case pod.Status.Phase == "Running":
			return true, nil
		default:
			t.Logf("Waiting for full availability of %s pod", name)
			return false, nil
		}
	})
	if err != nil {
		return xerrors.Errorf("pod %q in namespace %q cannot be retrieved or is not yet available: %w", name, namespace, err)
	}
	t.Logf("pod %s in namespace %s is available", name, namespace)
	return nil
}

// WaitForPodWithImage retrieves a pod using GetPod and waits for it to be running and available
func WaitForPodWithImage(t *testing.T, f *framework.Framework, namespace, name, image string, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		// Check if the CRD has been created
		pod, err := GetPod(namespace, name, image, f.KubeClient)
		switch {
		case apierrors.IsNotFound(err):
			t.Logf("Waiting for availability of %s pod", name)
			return false, nil
		case err != nil:
			return false, xerrors.Errorf("client failed to retrieve pod %q with image %q in namespace %q: %w", name, image, namespace, err)
		case pod != nil && pod.Status.Phase == "Running":
			return true, nil
		default:
			t.Logf("Waiting for full availability of %s pod", name)
			return false, nil
		}
	})
	if err != nil {
		pods := []string{}
		podsList, _ := f.KubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{})
		for _, p := range podsList.Items {
			images := []string{}
			for _, c := range p.Spec.Containers {
				images = append(images, c.Image)
			}
			pods = append(pods, fmt.Sprintf(`"%s" (%s)`, p.Name, strings.Join(images, " ")))
		}
		t.Logf("the following pods were found: %s", strings.Join(pods, ", "))
		return xerrors.Errorf("pod %q in namespace %q with image %q cannot be retrieved or is not yet available: %w", name, namespace, image, err)
	}
	t.Logf("pod %s in namespace %s is available", name, namespace)
	return nil
}

// WaitForPodAbsence waits until a pod with specified name (including namespace) and image is not found, or is Terminated.
func WaitForPodAbsence(t *testing.T, f *framework.Framework, namespace, name, image string, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		// Check that there's *no* pod with specified name
		pod, err := GetPod(namespace, name, image, f.KubeClient)
		switch {
		case apierrors.IsNotFound(err):
			return true, nil
		case err != nil:
			return false, xerrors.Errorf("client failed to retrieve pod %q with image %q in namespace %q: %w", name, image, namespace, err)
		case pod == nil || pod.Status.Phase == "Terminated":
			return true, nil
		default:
			t.Logf("Waiting for termination of %s pod [status: %s]", name, pod.Status.Phase)
			return false, nil
		}
	})
	if err != nil {
		return xerrors.Errorf("pod %q in namespace %q with image %q is still present: %w", name, namespace, image, err)
	}
	t.Logf("pod %s in namespace %s is absent", name, namespace)
	return nil
}

// DumpJobsLogsOnError checks if t is marked as failed, and if yes, dumps the logs of all pods in the specified namespace.
func DumpJobsLogsOnError(t *testing.T, f *framework.Framework, namespace string) {
	if !t.Failed() {
		return
	}
	pods := f.KubeClient.CoreV1().Pods(namespace)
	podsList, err := pods.List(metav1.ListOptions{})
	if err != nil {
		t.Logf("failed to list pods in namespace %s: %s", namespace, err)
		return
	}
	for _, p := range podsList.Items {
		match := false
		for _, c := range p.Spec.Containers {
			if strings.HasPrefix(c.Image, "quay.io/kohlstechnology/eunomia-") {
				match = true
				break
			}
		}
		if !match {
			continue
		}
		// Retrieve pod's logs
		req := pods.GetLogs(p.Name, &v1.PodLogOptions{Timestamps: true})
		logs, err := req.Stream()
		if err != nil {
			t.Logf("failed to retrieve logs for pod %s: %s", p.Name, err)
			continue
		}
		buf := &bytes.Buffer{}
		_, err = io.Copy(buf, logs)
		logs.Close()
		if err != nil {
			t.Logf("failed to retrieve logs for pod %s: %s", p.Name, err)
			continue
		}
		t.Logf("================ POD LOGS FOR %s ================\n%s\n\n", p.Name, buf.String())
	}
}

// debugJSON returns a possibly partial JSON representation of v, ignoring any
// errors. Intended to be used only in tests, to quickly display/dump various
// kinds of data for debugging/error message purposes.
func debugJSON(v interface{}) string {
	raw, _ := json.MarshalIndent(v, "", "  ")
	return string(raw)
}

// SetupRbacInNamespace deploys appropriate role and its binding to service
// account in given namespace for the test to be able to run seccessfuly.
//
// This function is needed because the function which normally handles
// per-namespace recource creation in operator-sdk framework
// https://github.com/operator-framework/operator-sdk/blob/v0.8.1/pkg/test/resource_creator.go#L108
// does not do the job in eunomia's case. It returns a call to a function
// CreateFromYAML (defined just above it) which uses SetNamespace method which
// sets a namespace in the yaml provided with "operator-sdk test local
// --namespaced-manifest" flag. However, this function
// https://github.com/kubernetes/apimachinery/blob/master/pkg/apis/meta/v1/unstructured/unstructured.go#L237-L243
// sets the namespace only in the metadata section of the resource. It is not
// enough in eunomia's case because the namespace needs to be also set in
// RoleBinding's subjects section:
// https://github.com/KohlsTechnology/eunomia/blob/v0.1.1/deploy/helm/eunomia-operator/templates/service-account-operator.yaml#L17
func SetupRbacInNamespace(namespace string) error {
	// get eunomia root dir
	_, utilTestFilename, _, _ := runtime.Caller(0)
	eunomiaRoot := filepath.Join(filepath.Dir(utilTestFilename), "../..")

	// create kubectl input
	out, err := exec.Command(
		"helm", "template",
		filepath.Join(eunomiaRoot, "deploy/helm/eunomia-operator/"),
		"--set", "eunomia.operator.namespace="+namespace,
		"--set", "eunomia.operator.deployment.nsRbacOnly=true",
	).Output()
	if err != nil {
		return xerrors.Errorf("Failed to generate manifest files: %w", err)
	}

	// create role, role binding and service account
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = bytes.NewReader(out)
	if err := cmd.Run(); err != nil {
		return xerrors.Errorf("Failed to create RBAC in %s namespace: %w", namespace, err)
	}

	return nil
}
