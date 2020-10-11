// Copyright 2020 Bull S.A.S. Atos Technologies - Bull, Rue Jean Jaures, B.P.68, 78340, Les Clayes-sous-Bois, France.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package alien4cloud

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func Test_reties(t *testing.T) {
	expectedBody := `
all my content
go
there
`
	loginCalled := new(bool)
	retryCalled := new(bool)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if regexp.MustCompile(`.*/login`).Match([]byte(r.URL.Path)) {
			*loginCalled = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
			return
		}

		if !*loginCalled {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":{"code": 403,"message":"login required"}}`))
			return
		}

		switch {
		case regexp.MustCompile(`.*/retry`).Match([]byte(r.URL.Path)):
			*retryCalled = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":"retried"}`))
			return
		default:
			if !*retryCalled {
				w.WriteHeader(http.StatusGatewayTimeout)
				_, _ = w.Write([]byte(`{"error":{"code": 504,"message":"This error should be retried"}}`))
				return
			}

			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf(`{"error":{"code": 500,"message": %q}}`, err.Error())))
				return
			}
			if string(b) != expectedBody {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf(`{"error":{"code": 500,"message": %q}}`, "not the expecting body: '"+string(b)+"'")))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":"success"}`))
			return
		}
	}))

	defer ts.Close()

	client, err := NewClient(ts.URL, "a", "a", "", false)
	assert.NilError(t, err)
	req, err := client.NewRequest(context.Background(), "POST", "/somepath", strings.NewReader(expectedBody))
	assert.NilError(t, err)

	// Checks that default retry function to re-login is call then this retry function should be called
	myRetryFn := func(c Client, request *http.Request, response *http.Response) (*http.Request, error) {
		if response.StatusCode != http.StatusGatewayTimeout {
			return nil, nil
		}
		r, err := c.NewRequest(context.Background(), "GET", "/retry", nil)
		if err != nil {
			return request, err
		}
		resp, err := c.Do(r)
		if resp != nil {
			err = ReadA4CResponse(resp, nil)
		}
		return request, err
	}

	respData := new(struct {
		Data string
	})

	resp, err := client.Do(req, myRetryFn)
	assert.NilError(t, err)
	err = ReadA4CResponse(resp, respData)
	assert.NilError(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	assert.Equal(t, respData.Data, "success")

}
