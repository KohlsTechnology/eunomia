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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	f "github.com/operator-framework/operator-sdk/pkg/test"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestMain(m *testing.M) {
	if err := setupTestEnv(); err != nil {
		log.Fatal(err)
	}
	defer cleanup()
	f.MainEntry(m)
}

func setupTestEnv() error {
	// Ensure minikube is running
	cmd := exec.Command("minikube", "status")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Minikube is not running, cannot perform e2e tests")
	}

	_, mainTestFilename, _, _ := runtime.Caller(0)
	eunomiaRoot := filepath.Join(filepath.Dir(mainTestFilename), "../..")

	// testFlags are set up here - thanks to this the command to make
	// end-to-end tests is only "go test ./test/e2e/... -tags e2e".
	// The flags have to be set to satisfy operator-sdk framework.
	testFlags := []string{
		"-namespacedMan", os.DevNull,
		"-globalMan", os.DevNull,
		"-root", eunomiaRoot,
		"-singleNamespace",
		"-test.parallel=1",
		"-localOperator",
	}
	for _, v := range testFlags {
		os.Args = append(os.Args, v)
	}

	// set up environmental variables as these are end-to-end tests
	eunomiaEnvs := map[string]string{
		"JOB_TEMPLATE":     filepath.Join(eunomiaRoot, "./build/job-templates/job.yaml"),
		"CRONJOB_TEMPLATE": filepath.Join(eunomiaRoot, "./build/job-templates/cronjob.yaml"),
		"WATCH_NAMESPACE":  "",
		"TEST_NAMESPACE":   "test-eunomia-operator",
		"OPERATOR_NAME":    "eunomia-operator",
		"GO111MODULE":      "on",
	}
	for k, v := range eunomiaEnvs {
		if err := os.Setenv(k, v); err != nil {
			return fmt.Errorf("Could not set env variable: %s", k)
		}
	}

	// If we're called as part of CI build on a PR, make sure we test the resources
	// (templates etc.) from the PR, instead of the master branch of the main repo
	prBranchEnv := "TRAVIS_PULL_REQUEST_BRANCH"
	prSlugEnv := "TRAVIS_PULL_REQUEST_SLUG"
	travisPrBranch, ok := os.LookupEnv(prBranchEnv)
	if ok {
		travisPrSlug, ok := os.LookupEnv(prSlugEnv)
		if !ok {
			return fmt.Errorf("%s not set, it must be set if %s is set", prSlugEnv, prBranchEnv)
		}
		if err := os.Setenv("EUNOMIA_URI", "https://github.com/"+travisPrSlug); err != nil {
			return fmt.Errorf("Could not set env variable: EUNOMIA_URI")
		}
		if err := os.Setenv("EUNOMIA_REF", travisPrBranch); err != nil {
			return fmt.Errorf("Could not set env variable: EUNOMIA_REF")
		}
	}

	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("Failed to connect to the k8s API server")
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("Failed to create a clientset to access k8s API objects")
	}
	api := clientset.CoreV1()

	// ensure clean namespace
	ns, err := api.Namespaces().Get(eunomiaEnvs["TEST_NAMESPACE"], v1.GetOptions{})
	if err == nil {
		if err = api.Namespaces().Delete(ns.GetName(), nil); err != nil {
			return fmt.Errorf("Failed to clean up %s namespace", ns.GetName())
		}
	}

	// Pre-populate the Docker registry in minikube with images built from the current commit
	// See also: https://stackoverflow.com/q/42564058
	minikubePort := "2376" // this port is hardcoded in minikube
	minikubeIp, err := exec.Command("minikube", "ip").Output()
	if err != nil {
		return fmt.Errorf("Could not retrieve minikube ip")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("Could not retrieve user home directory")
	}
	// Configure Docker client to use minikube Docker host
	hostDockerEnvs := map[string]string{
		"DOCKER_TLS_VERIFY": "1",
		"DOCKER_HOST":       "tcp://" + strings.TrimSpace(string(minikubeIp)) + ":" + minikubePort,
		"DOCKER_CERT_PATH":  filepath.Join(home, ".minikube/certs"),
	}
	for k, v := range hostDockerEnvs {
		if err := os.Setenv(k, v); err != nil {
			return fmt.Errorf("Could not set env variable: %s", k)
		}
	}
	// Build Docker images
	if err = os.Setenv("GOOS", "linux"); err != nil {
		return fmt.Errorf("Could not set GOOS to linux")
	}
	cmd = exec.Command(filepath.Join(eunomiaRoot, "scripts/build-images.sh"), "quay.io/kohlstechnology")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to build docker images")
	}

	// save eunomia helm template to temporary file
	tmpFile := filepath.Join(eunomiaRoot, "deploy/test/helm_temp.yml")
	out, err := exec.Command(
		"helm",
		"template",
		filepath.Join(eunomiaRoot, "deploy/helm/eunomia-operator/"),
		"--set", "eunomia.operator.deployment.enabled=",
		"--set", "eunomia.operator.namespace="+os.Getenv("TEST_NAMESPACE"),
	).Output()
	if err != nil {
		return fmt.Errorf("Failed to generate eunomia helm template")
	}
	if err = ioutil.WriteFile(tmpFile, out, 0644); err != nil {
		return fmt.Errorf("Failed to write helm template to temp file")
	}
	// deploy eunomia namespace etc.
	cmd = exec.Command("kubectl", "apply", "-f", tmpFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to deploy eunomia ns and helper files")
	}
	if err = os.Setenv("GOOS", ""); err != nil {
		return fmt.Errorf("Could not revert GOOS")
	}

	return nil
}

func cleanup() {
	// need to clean up the namespace, remove the tmpFile etc.
	fmt.Println("FINISH")
}
