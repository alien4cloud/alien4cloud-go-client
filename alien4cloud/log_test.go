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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"gotest.tools/v3/assert"
)

func Test_deploymentService_GetLogsOfApplication(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/deployments/search`).Match([]byte(r.URL.Path)):
			envID := r.URL.Query().Get("environmentId")
			var deploymentListResponse struct {
				Data struct {
					Data []struct {
						Deployment Deployment
					}
					TotalResults int `json:"totalResults"`
				} `json:"data"`
			}
			deploymentListResponse.Data.TotalResults = 1
			deploymentListResponse.Data.Data = []struct {
				Deployment Deployment
			}{
				{
					Deployment{
						ID: envID,
					},
				},
			}

			b, err := json.Marshal(&deploymentListResponse)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(b)
		case regexp.MustCompile(`.*/deployment/logs/search`).Match([]byte(r.URL.Path)):

			var lsr logsSearchRequest
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = json.Unmarshal(b, &lsr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if lsr.Filters.DeploymentID[0] == "error" {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var res struct {
				Data struct {
					Data         []Log `json:"data"`
					From         int   `json:"from"`
					To           int   `json:"to"`
					TotalResults int   `json:"totalResults"`
				} `json:"data"`
			}
			res.Data.TotalResults = 3
			res.Data.Data = []Log{
				{
					Content: "somelog",
					ID:      "1",
				},
				{
					Content: "somemorelog",
					ID:      "2",
				},
				{
					Content: "someotherlog",
					ID:      "3",
				},
			}
			b, err = json.Marshal(&res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(b)
		}

	}))

	type args struct {
		ctx        context.Context
		appID      string
		envID      string
		logFilters LogFilter
		fromIndex  int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"GetLogsOfApplicationOK", args{context.Background(), "normal", "envID", LogFilter{}, 0}, false},
		{"GetLogsOfApplicationError", args{context.Background(), "error", "error", LogFilter{}, 0}, true},
	}
	client, err := NewClient(ts.URL, "", "", "", true)
	assert.NilError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, _, err := client.LogService().GetLogsOfApplication(tt.args.ctx, tt.args.appID, tt.args.envID, tt.args.logFilters, tt.args.fromIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.GetLogsOfApplication() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}
