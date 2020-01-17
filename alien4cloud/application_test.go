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
	"net/http"
	"net/http/httptest"
	"regexp"
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
			// wait until test are finish to simulate long running op that will be cancelled
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":""}`))
			return
		}

		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))

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
