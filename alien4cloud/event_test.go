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
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"gotest.tools/v3/assert"
)

func Test_eventService_GetEvents(t *testing.T) {
	ts := newHTTPServerTestEvents(t)
	defer ts.Close()

	type args struct {
		ctx   context.Context
		appID string
		envID string
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		wantNbEvents int
	}{
		{"ExistingApp", args{context.Background(), "existingApp", "existingEnv"}, false, 1},
		{"UnknownApp", args{context.Background(), "unknownApp", "unknownEnv"}, true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			evService := &eventService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}

			_, nbEvents, err := evService.GetEventsForApplicationEnvironment(tt.args.ctx, tt.args.envID, 0, 10)
			if err != nil && !tt.wantErr {
				t.Errorf("eventService.GetEventsOfApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantNbEvents, nbEvents, "Unexpected number of events for app env %s", tt.args.envID)

		})
	}
}

func newHTTPServerTestEvents(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/deployments/existingEnv/events`).Match([]byte(r.URL.Path)):
			var res struct {
				Data struct {
					Data         []Event `json:"data"`
					From         int     `json:"from"`
					To           int     `json:"to"`
					TotalResults int     `json:"totalResults"`
				} `json:"data"`
			}
			event := Event{
				DeploymentID:     "testDeployement",
				DeploymentStatus: "DEPLOYED",
			}
			res.Data.Data = []Event{event}
			res.Data.To = 1
			res.Data.TotalResults = 1
			b, err := json.Marshal(&res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(b)
			return
		case regexp.MustCompile(`.*/deployments/.*/events`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"code": 504,"message":"Deployment does not exist"}}`))
			return
		}
		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))
}
