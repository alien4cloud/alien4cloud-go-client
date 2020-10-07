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

func Test_topologyService_GetTopology(t *testing.T) {
	ts := newHTTPServerTestTopology(t)
	defer ts.Close()

	type args struct {
		ctx   context.Context
		appID string
		envID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"ExistingApp", args{context.Background(), "existingApp", "existingEnv"}, false},
		{"UnknownApp", args{context.Background(), "unknownApp", "unknownEnv"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			topoService := &topologyService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}

			_, err := topoService.GetTopology(tt.args.ctx, tt.args.appID, tt.args.envID)
			if err != nil && !tt.wantErr {
				t.Errorf("topologyService.GetTopology() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_topologyService_GetTopologies(t *testing.T) {
	ts := newHTTPServerTestTopology(t)
	defer ts.Close()

	topoService := &topologyService{
		client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
	}
	allTopo, err := topoService.GetTopologies(context.Background(), "")
	if err != nil {
		t.Errorf("topologyService.GetTopologies() error = %v", err)
		return
	}
	assert.Equal(t, len(allTopo), 1, "Unexpected number of results for GetTopologies")
	assert.Equal(t, allTopo[0].ArchiveName, "testArchive", "Unexpected archive name in GetTopologies result")

}

func newHTTPServerTestTopology(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/applications/unknownApp/environments/.*/topology`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
			return
		case regexp.MustCompile(`.*/applications/.*/environments/.*/topology`).Match([]byte(r.URL.Path)):
			var res struct {
				Data string `json:"data"`
			}
			res.Data = "TopologyID"
			b, err := json.Marshal(&res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(b)
			return
		case regexp.MustCompile(`.*/topologies/search`).Match([]byte(r.URL.Path)):
			type DataStruct struct {
				ArchiveName string `json:"archiveName"`
				Workspace   string `json:"workspace"`
				ID          string `json:"id"`
			}
			var res struct {
				Data struct {
					Types []string     `json:"types"`
					Data  []DataStruct `json:"data"`
				} `json:"data"`
			}
			res.Data.Data = []DataStruct{DataStruct{ArchiveName: "testArchive"}}
			b, err := json.Marshal(&res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(b)
			return
		case regexp.MustCompile(`.*/topologies/.*`).Match([]byte(r.URL.Path)):
			var res Topology
			res.Data.Topology.ArchiveName = "myArchive"
			b, err := json.Marshal(&res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(b)
			return
		}
		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))
}
