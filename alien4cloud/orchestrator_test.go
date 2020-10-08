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

func newHTTPServerTestOrchestrator(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/orchestrators/error/locations`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
			return
		case regexp.MustCompile(`.*/orchestrators/.*/locations`).Match([]byte(r.URL.Path)):
			var res struct {
				Data []struct {
					Location struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"location"`
				} `json:"data"`
			}
			res.Data = []struct {
				Location struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"location"`
			}{
				{
					Location: struct {
						ID   string "json:\"id\""
						Name string "json:\"name\""
					}{
						ID:   "1",
						Name: "location1",
					},
				},
				{
					Location: struct {
						ID   string "json:\"id\""
						Name string "json:\"name\""
					}{
						ID:   "2",
						Name: "location2",
					},
				},
			}
			b, err := json.Marshal(&res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(b)
			return
		case regexp.MustCompile(`.*/orchestrators`).Match([]byte(r.URL.Path)):
			sr := new(SearchRequest)
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = json.Unmarshal(b, sr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if sr.Query == "error" {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
				return
			}

			var res struct {
				Data struct {
					Data []struct {
						ID               string `json:"id"`
						OrchestratorName string `json:"name"`
					} `json:"data"`
					TotalResults int `json:"totalResults"`
				} `json:"data"`
			}

			if sr.Query != "noresults" {
				res.Data.TotalResults = 1
				res.Data.Data = make([]struct {
					ID               string "json:\"id\""
					OrchestratorName string "json:\"name\""
				}, 0)
				res.Data.Data = append(res.Data.Data, struct {
					ID               string "json:\"id\""
					OrchestratorName string "json:\"name\""
				}{
					ID:               "orchID1",
					OrchestratorName: "orch1",
				})
				if sr.Query == "multiresults" {
					res.Data.Data = append(res.Data.Data, struct {
						ID               string "json:\"id\""
						OrchestratorName string "json:\"name\""
					}{
						ID:               "orchID2",
						OrchestratorName: "orch2",
					})
					res.Data.TotalResults = 2
				}
			}
			b, err = json.Marshal(&res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(b)
			return

		}
		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))
}

func Test_orchestratorService_GetOrchestratorLocations(t *testing.T) {
	ts := newHTTPServerTestOrchestrator(t)
	defer ts.Close()
	type args struct {
		orchestratorID string
	}
	tests := []struct {
		name    string
		args    args
		want    []Location
		wantErr bool
	}{
		{"GetOrchestratorLocationsOK", args{"normal"}, []Location{{ID: "1", Name: "location1"}, {ID: "2", Name: "location2"}}, false},
		{"GetOrchestratorLocationsError", args{"error"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &orchestratorService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}
			got, err := o.GetOrchestratorLocations(context.Background(), tt.args.orchestratorID)
			if (err != nil) != tt.wantErr {
				t.Errorf("orchestratorService.GetOrchestratorLocations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.DeepEqual(t, got, tt.want)

		})
	}
}

func Test_orchestratorService_GetOrchestratorIDbyName(t *testing.T) {
	ts := newHTTPServerTestOrchestrator(t)
	defer ts.Close()

	type args struct {
		orchestratorName string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"GetOrchestratorIDbyNameSingle", args{"normal"}, "orchID1", false},
		{"GetOrchestratorIDbyNameMultiResults", args{"multiresults"}, "orchID1", false},
		{"GetOrchestratorIDbyNameNoResults", args{"noresults"}, "", true},
		{"GetOrchestratorIDbyNameError", args{"error"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &orchestratorService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}
			got, err := o.GetOrchestratorIDbyName(context.Background(), tt.args.orchestratorName)
			if (err != nil) != tt.wantErr {
				t.Errorf("orchestratorService.GetOrchestratorIDbyName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("orchestratorService.GetOrchestratorIDbyName() = %v, want %v", got, tt.want)
			}
		})
	}
}
