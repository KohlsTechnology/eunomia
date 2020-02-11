// +build e2e

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

package e2e

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	util "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestReadinessAndLivelinessProbes(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	operatorName, found := os.LookupEnv("OPERATOR_NAME")
	if !found {
		t.Fatal("OPERATOR_NAME environment value missing")
	}
	operatorNamespace, found := os.LookupEnv("OPERATOR_NAMESPACE")
	if !found {
		t.Fatal("OPERATOR_NAMESPACE environment value missing")
	}
	minikubeIP, found := os.LookupEnv("MINIKUBE_IP")
	if !found {
		t.Fatal("MINIKUBE_IP environment value missing")
	}
	webHookPort, found := os.LookupEnv("OPERATOR_WEBHOOK_PORT")
	if !found {
		t.Fatal("OPERATOR_WEBHOOK_PORT environment value missing")
	}
	webHookPortInt, err := strconv.Atoi(webHookPort)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("minikube IP: %s", minikubeIP)

	// The aim of this test is to validate that the required endpoints are there and are accessible. To
	// achieve this, we need to expose the deployment externally outside of minikube as this test runs
	// outside of it. To ensure that this is done correctly, a Service with Type: "NodePort" is created,
	// to expose the operator's webhook via high, random port.
	service := &corev1.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "minikube-exposing-service",
			Namespace: operatorNamespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "webhook",
					Protocol: corev1.ProtocolTCP,
					Port:     int32(webHookPortInt),
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(webHookPortInt),
					},
				},
			},
			Selector: map[string]string{"name": operatorName},
			Type:     "NodePort",
		},
	}

	err = framework.Global.Client.Create(context.TODO(), service, &framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       timeout,
		RetryInterval: retryInterval,
	})
	if err != nil {
		t.Error(err)
	}
	nodePort := service.Spec.Ports[0].NodePort

	t.Logf("minikube exposing service Node Port: %d", nodePort)

	err = util.WaitForOperatorDeployment(t, framework.Global.KubeClient, operatorNamespace, operatorName, 1, retryInterval, timeout)
	if err != nil {
		t.Error(err)
	}

	//Waiting for service to get connection to operator pod
	for retryCount := 0; retryCount < 50; retryCount++ {
		t.Logf("retrying %d", retryCount)
		resp, err := http.Get(fmt.Sprintf("http://%s:%d/readyz", minikubeIP, nodePort))
		if err != nil {
			t.Log(err)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			break
		}
	}

	tests := []struct {
		endpoint string
	}{
		{
			endpoint: "readyz",
		},
		{
			endpoint: "healthz",
		},
	}

	for _, tt := range tests {
		resp, err := http.Get(fmt.Sprintf("http://%s:%d/%s", minikubeIP, nodePort, tt.endpoint))
		if err != nil {
			t.Errorf("%q: %s", tt.endpoint, err)
			continue
		}
		defer resp.Body.Close()
		if http.StatusOK != resp.StatusCode {
			t.Errorf("%q: returned status: %d, wanted: %d", tt.endpoint, resp.StatusCode, http.StatusOK)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("%q: %s", tt.endpoint, err)
			continue
		}
		if "ok" != string(body) {
			t.Errorf("%q: returned body: %s, wanted: %s", tt.endpoint, string(body), "ok")
		}
	}
}
