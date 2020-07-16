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
	"net/http"
	"testing"
)

func TestReadinessAndLivelinessProbes(t *testing.T) {
	ctx, err := NewContext(t)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Cleanup()

	hostUrl := ExposeOperatorAsService(t, ctx)

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
		resp, err := http.Get(fmt.Sprintf("%s/%s", hostUrl, tt.endpoint))
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
