// Copyright 2019 Bull S.A.S. Atos Technologies - Bull, Rue Jean Jaures, B.P.68, 78340, Les Clayes-sous-Bois, France.
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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func Test_applicationService_IsApplicationExists(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/applications/unknown`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
			return
		case regexp.MustCompile(`.*/applications/existing`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":""}`))
			return
		}

		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))

	defer ts.Close()

	type args struct {
		ctx   context.Context
		appID string
	}
	tests := []struct {
		name   string
		args   args
		exists bool
	}{
		{"ExistingApp", args{context.Background(), "existing"}, true},
		{"UnknownApp", args{context.Background(), "unknown"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			a := &applicationService{
				client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
			}

			found, err := a.IsApplicationExist(tt.args.ctx, tt.args.appID)
			if err != nil {
				t.Errorf("applicationService.IsApplicationExist() error = %v", err)
			}
			assert.Equal(t, tt.exists, found, "Unexpected result for IsApplicationExist %s", tt.args.appID)
		})
	}
}

func Test_applicationService_GetApplicationsID(t *testing.T) {
	ts := newHTTPServerTestApplicationSearch(t)
	defer ts.Close()
	type args struct {
		ctx   context.Context
		appID string
	}
	tests := []struct {
		name   string
		args   args
		number int
	}{
		{"ExistingApp", args{context.Background(), "existingApp"}, 1},
		{"UnknownApp", args{context.Background(), "unknownApp"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			a := &applicationService{
				client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
			}

			results, err := a.GetApplicationsID(tt.args.ctx, tt.args.appID)
			if err != nil {
				t.Errorf("applicationService.GetApplicationsID() error = %v", err)
			}
			assert.Equal(t, len(results), tt.number, "Unexpected number of results for GetApplicationsID")
		})
	}
}

func Test_applicationService_GetApplicationByID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/applications/unknownID`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"code": 404,"message":"cannot be found"}}`))
			return
		case regexp.MustCompile(`.*/applications/existingID`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"id":"existingID","name":"existingApp","tags":[]}}`))
			return
		}

		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))

	defer ts.Close()
	type args struct {
		ctx   context.Context
		appID string
	}
	tests := []struct {
		id      string
		args    args
		wantErr bool
		want    string
	}{
		{"existingID", args{context.Background(), "existingID"}, false, "existingApp"},
		{"unknownID", args{context.Background(), "unknownID"}, true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {

			a := &applicationService{
				client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
			}

			app, err := a.GetApplicationByID(tt.args.ctx, tt.args.appID)
			if (err != nil) != tt.wantErr {
				t.Errorf("catalogService.UploadCSAR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want, app.Name, "Unexpected result for GetApplicationByID()")
			}
		})
	}
}

func newHTTPServerTestApplicationSearch(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if regexp.MustCompile(`.*/applications`).Match([]byte(r.URL.Path)) {

			var searchReq SearchRequest
			rb, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Errorf("Failed to read request body %+v", r)
			}
			defer r.Body.Close()
			s := string(rb)
			t.Logf("request: %s", s)

			err = json.Unmarshal(rb, &searchReq)
			if err != nil {
				t.Errorf("Failed to unmarshal request body %+v", r)
			}
			if strings.Contains(searchReq.Query, "existingApp") {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data":{"types":["Application"],"data":[{"id":"existingApp"}],"totalResults":1}}`))
			} else {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
			}
			return
		}

		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))
}
