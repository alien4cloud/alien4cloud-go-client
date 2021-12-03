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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"regexp"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func Test_deploymentService_GetDeployment(t *testing.T) {
	closeCh := make(chan struct{})
	defer close(closeCh)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		deploymentID := path.Base(r.URL.Path)

		if deploymentID == "NotFound" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var resDep struct {
			Data struct {
				Deployment `json:"deployment"`
			} `json:"data"`
		}
		resDep.Data.Deployment = Deployment{
			ID:                       deploymentID,
			EnvironmentID:            "envID",
			StartDate:                mustParseTime(t, "2020-01-13 21:58:27.377 +0100 CET"),
			EndDate:                  mustParseTime(t, "2020-01-13 21:58:45.749 +0100 CET"),
			DeploymentUsername:       "user",
			LocationIds:              []string{"loc1", "loc2"},
			OrchestratorDeploymentID: "orchestratorDeploymentID",
			OrchestratorID:           "orchestratorID",
			SourceID:                 "sourceID",
			SourceName:               "sourceName",
			SourceType:               "sourceType",
			VersionID:                "versionID",
			WorkflowExecutions:       map[string]string{"workflow1": "execution1", "workflow2": "execution2"},
		}

		b, err := json.Marshal(&resDep)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)

	}))

	type args struct {
		ctx          context.Context
		deploymentID string
	}
	tests := []struct {
		name    string
		args    args
		want    Deployment
		wantErr bool
	}{
		{"success",
			args{context.Background(), "depID51"},
			Deployment{ID: "depID51", EnvironmentID: "envID", StartDate: mustParseTime(t, "2020-01-13 21:58:27.377 +0100 CET"), EndDate: mustParseTime(t, "2020-01-13 21:58:45.749 +0100 CET"), DeploymentUsername: "user", LocationIds: []string{"loc1", "loc2"}, OrchestratorDeploymentID: "orchestratorDeploymentID", OrchestratorID: "orchestratorID", SourceID: "sourceID", SourceName: "sourceName", SourceType: "sourceType", VersionID: "versionID", WorkflowExecutions: map[string]string{"workflow1": "execution1", "workflow2": "execution2"}},
			false,
		},
		{"failure",
			args{context.Background(), "NotFound"},
			Deployment{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &deploymentService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}
			got, err := d.GetDeployment(tt.args.ctx, tt.args.deploymentID)
			if (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.GetDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.DeepEqual(t, got, tt.want)
		})
	}

}

func Test_deploymentService_DeployApplication(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/topologies/TopologyID/locations.*`).Match([]byte(r.URL.Path)):
			var res struct {
				Data []LocationMatch `json:"data"`
			}
			res.Data = []LocationMatch{
				{
					Location: LocationConfiguration{
						Name:           "location",
						ID:             "locationID",
						OrchestratorID: "orchestratorID",
					},
				},
			}
			b, err := json.Marshal(&res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(b)
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

		case regexp.MustCompile(`.*/applications/.*/environments/.*/deployment-topology/location-policies`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			return
		case regexp.MustCompile(`.*/applications/deployment`).Match([]byte(r.URL.Path)):
			b, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			appDeployRequest := new(ApplicationDeployRequest)
			err = json.Unmarshal(b, appDeployRequest)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if appDeployRequest.ApplicationID == "error" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))

	type args struct {
		ctx      context.Context
		appID    string
		envID    string
		location string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"NormalDeploy", args{context.Background(), "normal", "envID", "location"}, false},
		{"DeployError", args{context.Background(), "error", "envID", "location"}, true},
	}
	client, err := NewClient(ts.URL, "", "", "", false)
	assert.NilError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			d := &deploymentService{
				client: client.(*a4cClient),
			}

			if err := d.DeployApplication(tt.args.ctx, tt.args.appID, tt.args.envID, tt.args.location); (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.DeployApplication() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_deploymentService_UpdateApplication(t *testing.T) {
	closeCh := make(chan struct{})
	defer close(closeCh)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/applications/error/environments/.*/update-deployment`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusInternalServerError)
			return
		case regexp.MustCompile(`.*/applications/cancel/environments/.*/update-deployment`).Match([]byte(r.URL.Path)):
			// wait until test are finish to simulate long running op that will be cancelled
			<-closeCh
			w.WriteHeader(http.StatusOK)
			return
		case regexp.MustCompile(`.*/applications/.*/environments/.*/update-deployment`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			return
		}

		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))

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
		{"NormalUpdate", args{context.Background(), "normal", "envID"}, false},
		{"UpdateError", args{context.Background(), "error", "envID"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			d := &deploymentService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}

			if err := d.UpdateApplication(tt.args.ctx, tt.args.appID, tt.args.envID); (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.UpdateApplication() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	cancelableCtx, cancelFn := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancelFn()
	d := &deploymentService{
		client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
	}

	if err := d.UpdateApplication(cancelableCtx, "cancel", "envID"); err == nil {
		t.Error("deploymentService.UpdateApplication() expecting an error")
	}

}

func Test_deploymentService_GetDeploymentList(t *testing.T) {
	mt := &Time{time.Now()}
	b, err := json.Marshal(mt)
	assert.NilError(t, err)
	err = json.Unmarshal(b, mt)
	assert.NilError(t, err)
	expectedResult := []Deployment{
		{
			ID:            "D1",
			EnvironmentID: "E1",
			StartDate:     *mt,
			EndDate:       *mt,
		},
		{
			ID:            "D2",
			EnvironmentID: "E2",
			StartDate:     *mt,
			EndDate:       *mt,
		},
		{
			ID:            "D3",
			EnvironmentID: "E3",
			StartDate:     *mt,
			EndDate:       *mt,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		envID := r.URL.Query().Get("environmentId")
		if envID == "error" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
			return
		}
		var deploymentListResponse struct {
			Data struct {
				Data []struct {
					Deployment Deployment
				}
				TotalResults int `json:"totalResults"`
			} `json:"data"`
		}
		deploymentListResponse.Data.TotalResults = len(expectedResult)
		deploymentListResponse.Data.Data = make([]struct{ Deployment Deployment }, 0, len(expectedResult))
		for _, er := range expectedResult {
			deploymentListResponse.Data.Data = append(deploymentListResponse.Data.Data, struct{ Deployment Deployment }{er})
		}
		b, err := json.Marshal(&deploymentListResponse)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}))

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
		{"GetDeploymentListOK", args{context.Background(), "normal", "envID"}, false},
		{"GetDeploymentListError", args{context.Background(), "error", "error"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			d := &deploymentService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}
			got, err := d.GetDeploymentList(tt.args.ctx, tt.args.appID, tt.args.envID)
			if (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.GetDeploymentList() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				assert.DeepEqual(t, got, expectedResult)
			}
		})
	}
}

func Test_deploymentService_GetAttributesValue(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/applications/error/environments/.*/deployment/informations`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
			return
		case regexp.MustCompile(`.*/applications/noresult/environments/.*/deployment/informations`).Match([]byte(r.URL.Path)):
			info := new(Informations)
			b, err := json.Marshal(info)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(b)
			w.WriteHeader(http.StatusOK)
			return
		case regexp.MustCompile(`.*/applications/.*/environments/.*/deployment/informations`).Match([]byte(r.URL.Path)):
			info := new(Informations)
			info.Data = map[string]map[string]struct {
				State      string            "json:\"state\""
				Attributes map[string]string "json:\"attributes\""
			}{
				"node1": {
					"0": {
						Attributes: map[string]string{
							"attr1": "val1",
							"attr2": "val2",
							"attr3": "val3",
						},
					},
					"1": {
						Attributes: map[string]string{
							"attr1": "val11",
							"attr2": "val12",
							"attr3": "val13",
						},
					},
				},
			}

			b, err := json.Marshal(info)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(b)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))

	type args struct {
		ctx                 context.Context
		appID               string
		envID               string
		nodeName            string
		requestedAttributes []string
	}
	tests := []struct {
		name               string
		args               args
		wantErr            bool
		expectedAttributes map[string]string
	}{
		{"GetAttributesValueOK", args{context.Background(), "normal", "envID", "node1", []string{"attr1", "attr3"}}, false, map[string]string{"attr1": "val1", "attr3": "val3"}},
		{"GetAttributesValueNoResult", args{context.Background(), "noresult", "envID", "node1", nil}, false, nil},
		{"GetAttributesValueError", args{context.Background(), "error", "envID", "node1", nil}, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			d := &deploymentService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}
			attributes, err := d.GetAttributesValue(tt.args.ctx, tt.args.appID, tt.args.envID, tt.args.nodeName, tt.args.requestedAttributes)
			if (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.GetAttributesValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.DeepEqual(t, attributes, tt.expectedAttributes)
		})
	}
}

func Test_deploymentService_GetNodeStatus(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/applications/error/environments/.*/deployment/informations`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
			return
		case regexp.MustCompile(`.*/applications/noresult/environments/.*/deployment/informations`).Match([]byte(r.URL.Path)):
			info := new(Informations)
			b, err := json.Marshal(info)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(b)
			w.WriteHeader(http.StatusOK)
			return
		case regexp.MustCompile(`.*/applications/.*/environments/.*/deployment/informations`).Match([]byte(r.URL.Path)):
			info := new(Informations)
			info.Data = map[string]map[string]struct {
				State      string            "json:\"state\""
				Attributes map[string]string "json:\"attributes\""
			}{
				"node1": {
					"0": {
						State: "STARTED",
					},
					"1": {
						State: "ERROR",
					},
				},
			}

			b, err := json.Marshal(info)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(b)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))

	type args struct {
		ctx      context.Context
		appID    string
		envID    string
		nodeName string
	}
	tests := []struct {
		name           string
		args           args
		wantErr        bool
		expectedStatus string
	}{
		{"GetNodeStatusOK", args{context.Background(), "normal", "envID", "node1"}, false, "STARTED"},
		{"GetNodeStatusNoResult", args{context.Background(), "noresult", "envID", "node1"}, false, ""},
		{"GetNodeStatusError", args{context.Background(), "error", "envID", "node1"}, true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			d := &deploymentService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}
			status, err := d.GetNodeStatus(tt.args.ctx, tt.args.appID, tt.args.envID, tt.args.nodeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.GetNodeStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, status, tt.expectedStatus)
		})
	}
}

func Test_deploymentService_GetOutputAttributes(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/runtime/error/environment/.*/topology`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
			return
		case regexp.MustCompile(`.*/runtime/noresult/environment/.*/topology`).Match([]byte(r.URL.Path)):
			info := new(RuntimeTopology)
			b, err := json.Marshal(info)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(b)
			w.WriteHeader(http.StatusOK)
			return
		case regexp.MustCompile(`.*/runtime/.*/environment/.*/topology`).Match([]byte(r.URL.Path)):
			info := new(RuntimeTopology)
			info.Data.Topology.OutputAttributes = map[string][]string{"output1": {"v11", "v12"}, "output2": {"v21", "v22"}}

			b, err := json.Marshal(info)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(b)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))

	type args struct {
		ctx   context.Context
		appID string
		envID string
	}
	tests := []struct {
		name            string
		args            args
		wantErr         bool
		expectedOutputs map[string][]string
	}{
		{"GetOutputAttributesOK", args{context.Background(), "normal", "envID"}, false, map[string][]string{"output1": {"v11", "v12"}, "output2": {"v21", "v22"}}},
		{"GetOutputAttributesNoResult", args{context.Background(), "noresult", "envID"}, false, nil},
		{"GetOutputAttributesError", args{context.Background(), "error", "envID"}, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			d := &deploymentService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}
			status, err := d.GetOutputAttributes(tt.args.ctx, tt.args.appID, tt.args.envID)
			if (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.GetOutputAttributes() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.DeepEqual(t, status, tt.expectedOutputs)
		})
	}
}

func Test_deploymentService_undeployApplication(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/applications/error/environments/.*/deployment`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
			return
		case regexp.MustCompile(`.*/applications/.*/environments/.*/deployment`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			return
		}

		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))

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
		{"UndeployApplicationOK", args{context.Background(), "normal", "envID"}, false},
		{"UndeployApplicationError", args{context.Background(), "error", "envID"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			d := &deploymentService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}

			if err := d.UndeployApplication(tt.args.ctx, tt.args.appID, tt.args.envID); (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.UndeployApplication() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_deploymentService_WaitUntilStateIs(t *testing.T) {
	closeCh := make(chan struct{})
	defer close(closeCh)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/applications/err/environments/.*/active-deployment-monitored`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"deployment":{"id":"err"}}}`))
			return
		case regexp.MustCompile(`.*/applications/.*/environments/.*/active-deployment-monitored`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"deployment":{"id":"myID"}}}`))
			return
		case regexp.MustCompile(`.*/deployments/err/status`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
			return
		case regexp.MustCompile(`.*/deployments/.*/status`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":"%s"}`, ApplicationDeployed)))
			return

		}

		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))

	type args struct {
		ctx      context.Context
		appID    string
		envID    string
		statuses []string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"MissingStatues", args{context.Background(), "app", "env", nil}, "", true},
		{"DeployedStatus", args{context.Background(), "app", "env", []string{ApplicationDeployed}}, ApplicationDeployed, false},
		{"DeployedWithOtherStatuses", args{context.Background(), "app", "env", []string{ApplicationError, ApplicationUndeployed, ApplicationDeployed}}, ApplicationDeployed, false},
		{"ErrorNotFound", args{context.Background(), "err", "env", []string{ApplicationError, ApplicationUndeployed, ApplicationDeployed}}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &deploymentService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}
			got, err := d.WaitUntilStateIs(tt.args.ctx, tt.args.appID, tt.args.envID, tt.args.statuses...)
			if (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.WaitUntilStateIs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("deploymentService.WaitUntilStateIs() = %v, want %v", got, tt.want)
			}
		})
	}

	cancelableCtx, cancelFn := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancelFn()
	d := &deploymentService{
		client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
	}

	if _, err := d.WaitUntilStateIs(cancelableCtx, "cancel", "envID", ApplicationUpdated); err == nil {
		t.Error("deploymentService.WaitUntilStateIs() expecting an error")
	}
}

func Test_deploymentService_GetDeploymentStatus(t *testing.T) {
	closeCh := make(chan struct{})
	defer close(closeCh)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/applications/UndeployedApp/environments/.*/active-deployment-monitored`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{}}`))
			return
		case regexp.MustCompile(`.*/applications/UnknownApp/environments/.*/active-deployment-monitored`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
			return
		case regexp.MustCompile(`.*/applications/.*/environments/.*/active-deployment-monitored`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"deployment":{"id":"myID"}}}`))
			return
		case regexp.MustCompile(`.*/deployments//status`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
			return
		case regexp.MustCompile(`.*/deployments/.*/status`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":"%s"}`, ApplicationDeployed)))
			return

		}

		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))

	type args struct {
		ctx   context.Context
		appID string
		envID string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"UndeployedStatus", args{context.Background(), "UndeployedApp", "env"}, ApplicationUndeployed, false},
		{"DeployedStatus", args{context.Background(), "app", "env"}, ApplicationDeployed, false},
		{"ErrorNotFound", args{context.Background(), "UnknownApp", "env"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &deploymentService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}
			got, err := d.GetDeploymentStatus(tt.args.ctx, tt.args.appID, tt.args.envID)
			if (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.GetDeploymentStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("deploymentService.GetDeploymentStatus() = %v, want %v", got, tt.want)
			}
		})
	}

	cancelableCtx, cancelFn := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancelFn()
	d := &deploymentService{
		client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
	}

	if _, err := d.WaitUntilStateIs(cancelableCtx, "cancel", "envID", ApplicationUpdated); err == nil {
		t.Error("deploymentService.WaitUntilStateIs() expecting an error")
	}
}

func Test_deploymentService_RunWorkflow(t *testing.T) {
	closeCh := make(chan struct{})
	defer close(closeCh)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/applications/app/environments/env/active-deployment-monitored`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"deployment":{"id":"4186a188-24a4-4910-9d7b-207ca09f98e3"}}}`))
			return
		case regexp.MustCompile(`.*/applications/BadExecID/environments/env/active-deployment-monitored`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusInternalServerError)
			return
		case regexp.MustCompile(`.*/applications/app/environments/env/workflows/wf`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":"execID"}`))
			return
		case regexp.MustCompile(`.*/applications/error/environments/.*/workflows/.*`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusInternalServerError)
			return
		case regexp.MustCompile(`.*/applications/testcancel/environments/.*/workflows/.*`).Match([]byte(r.URL.Path)):
			// wait until test are finish to simulate long running op that will be cancelled
			<-closeCh
			w.WriteHeader(http.StatusOK)
			return
		case regexp.MustCompile(`.*/applications/.*/environments/.*/workflows/.*`).Match([]byte(r.URL.Path)):
			matches := regexp.MustCompile(`.*/applications/(.*)/environments/.*/workflows/.*`).FindStringSubmatch(r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":"%s"}`, matches[1])))
			return
		case regexp.MustCompile(`.*/executions/execID`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"id":"execID","deploymentId":"4186a188-24a4-4910-9d7b-207ca09f98e3","workflowId":"wf","workflowName":"wf","displayWorkflowName":"wf","startDate":1578949107377,"endDate":1578949125749,"status":"SUCCEEDED","hasFailedTasks":false}}`))
			return
		case regexp.MustCompile(`.*/executions/execCancel`).Match([]byte(r.URL.Path)) && r.URL.Query().Get("query") == "execCancel":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"id":"7459ca00-f98f-47f1-a7e8-4d779d65253a","deploymentId":"4186a188-24a4-4910-9d7b-207ca09f98e3","workflowId":"execCancel","workflowName":"execCancel","displayWorkflowName":"execCancel","startDate":1578949107377,"endDate":1578949125749,"status":"RUNNING","hasFailedTasks":false}}`))
			return
		case regexp.MustCompile(`.*/executions/.*`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"id":"7459ca00-f98f-47f1-a7e8-4d779d65253a","deploymentId":"4186a188-24a4-4910-9d7b-207ca09f98e3","workflowId":"stopWebServer","workflowName":"stopWebServer","displayWorkflowName":"stopWebServer","startDate":1578949107377,"endDate":1578949125749,"status":"SUCCEEDED","hasFailedTasks":false}}`))
			return
		}

		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))

	type args struct {
		ctx          context.Context
		a4cAppID     string
		a4cEnvID     string
		workflowName string
		timeout      time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    *Execution
		wantErr bool
	}{
		{"Normal", args{context.Background(), "app", "env", "wf", 5 * time.Minute},
			&Execution{ID: "execID", DeploymentID: "4186a188-24a4-4910-9d7b-207ca09f98e3", WorkflowID: "wf", WorkflowName: "wf", DisplayWorkflowName: "wf", Status: "SUCCEEDED", HasFailedTasks: false, StartDate: mustParseTime(t, "2020-01-13 21:58:27.377 +0100 CET"), EndDate: mustParseTime(t, "2020-01-13 21:58:45.749 +0100 CET")},
			false,
		},
		{"BadExecID", args{context.Background(), "error", "env", "wf", 5 * time.Minute}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &deploymentService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}
			got, err := d.RunWorkflow(tt.args.ctx, tt.args.a4cAppID, tt.args.a4cEnvID, tt.args.workflowName, tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.RunWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.DeepEqual(t, got, tt.want)
		})
	}

	cancelableCtx, cancelFn := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancelFn()
	d := &deploymentService{
		client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
	}

	_, err := d.RunWorkflow(cancelableCtx, "app", "env", "cancelWf", 500*time.Millisecond)
	assert.ErrorContains(t, err, "context deadline exceeded")

	cancelableCtx, cancelFn = context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancelFn()
	_, err = d.RunWorkflow(cancelableCtx, "app", "env", "cancelWf", 50*time.Millisecond)
	assert.ErrorContains(t, err, "context deadline exceeded")

	cancelableCtx, cancelFn = context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancelFn()
	_, err = d.RunWorkflow(cancelableCtx, "app", "env", "execCancelWf", 50*time.Millisecond)
	assert.ErrorContains(t, err, "context deadline exceeded")

	cancelableCtx, cancelFn = context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancelFn()
	_, err = d.RunWorkflow(cancelableCtx, "app", "env", "execCancelWf", 50*time.Millisecond)
	assert.ErrorContains(t, err, "context deadline exceeded")
}

func Test_deploymentService_UpdateDeploymentSetup(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/applications/error/environments/.*/deployment-topology`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusInternalServerError)
			return
		case regexp.MustCompile(`.*/applications/.*/environments/.*/deployment-topology`).Match([]byte(r.URL.Path)):
			var req UpdateDeploymentTopologyRequest
			rb, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Errorf("Failed to read request body %+v", r)
			}
			defer r.Body.Close()
			s := string(rb)
			t.Logf("request: %s", s)

			err = json.Unmarshal(rb, &req)
			if err != nil {
				t.Errorf("Failed to unmarshal request body %+v", r)
			}
			assert.Equal(t, req.InputProperties["testInputProp"], "testValue")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":""}`))
			return
		}

		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))

	defer ts.Close()

	type args struct {
		ctx                context.Context
		appID              string
		envID              string
		inputPropertyName  string
		inputPropertyValue string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"UpdateInput", args{context.Background(), "normal", "envID", "testInputProp", "testValue"}, false},
		{"UpdateError", args{context.Background(), "error", "envID", "inputPropErr", "valErr"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			d := &deploymentService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}

			err := d.UpdateDeploymentTopology(tt.args.ctx, tt.args.appID, tt.args.envID,
				UpdateDeploymentTopologyRequest{
					InputProperties: map[string]interface{}{
						tt.args.inputPropertyName: tt.args.inputPropertyValue,
					},
				})
			if err != nil && !tt.wantErr {
				t.Errorf("deploymentService.UpdateDeploymentTopology() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func Test_deploymentService_UploadDeploymentInputArtifact(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/applications/error/environments/.*/deployment-topology/inputArtifacts/.*/upload`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusInternalServerError)
			return
		case regexp.MustCompile(`.*/applications/.*/environments/.*/deployment-topology/inputArtifacts/.*/upload`).Match([]byte(r.URL.Path)):
			rb, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Errorf("Failed to read request body %+v", r)
			}
			defer r.Body.Close()
			s := string(rb)
			t.Logf("request: %s", s)

			if !strings.Contains(s, "testContent") {
				t.Errorf("Failed to find expected content in uploaded file")
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":""}`))
			return
		}

		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))

	defer ts.Close()

	type args struct {
		ctx               context.Context
		appID             string
		envID             string
		inputArtifactName string
		content           string
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		wrongFile bool
	}{
		{"TestUpdateInputArtifact", args{context.Background(), "normal", "envID",
			"testArtifact", "testContent"}, false, false},
		{"TestUpdateInputArtifactError", args{context.Background(), "error", "envID",
			"testArtifact", "testError"}, true, false},
		{"TestUpdateInputArtifactWrongPath", args{context.Background(), "error", "envID",
			"testArtifact", "testError"}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			d := &deploymentService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}

			f, err := ioutil.TempFile("", tt.name)
			artifactPath := f.Name()
			if tt.wrongFile {
				artifactPath = "badFile"
			}

			assert.NilError(t, err, "Failed to create a file to upload")
			_, err = f.Write([]byte(tt.args.content))
			_ = f.Sync()
			_ = f.Close()
			defer os.Remove(f.Name())
			assert.NilError(t, err, "Failed to write to file to upload")

			err = d.UploadDeploymentInputArtifact(tt.args.ctx, tt.args.appID, tt.args.envID,
				tt.args.inputArtifactName, artifactPath)
			if err != nil && !tt.wantErr {
				t.Errorf("deploymentService.UpdateDeploymentTopology() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func Test_deploymentService_GetLastWorkflowExecution(t *testing.T) {
	endDate := mustParseTime(t, "2021-05-10 17:18:41.608 +0200 CEST")
	startDate := mustParseTime(t, "2021-05-10 16:18:41.608 +0200 CEST")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/applications/.*/environments/.*/active-deployment-monitored`).Match([]byte(r.URL.Path)):
			appName := regexp.MustCompile(`.*/applications/(.*)/environments/.*/active-deployment-monitored`).FindStringSubmatch(r.URL.Path)[1]
			var res struct {
				Data struct {
					Deployment struct {
						ID string `json:"id"`
					} `json:"deployment"`
				} `json:"data"`
			}
			res.Data.Deployment.ID = appName
			b, err := json.Marshal(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(b)
			return
		case regexp.MustCompile(`.*/workflow_execution/error`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
			return
		case regexp.MustCompile(`.*/workflow_execution/.*`).Match([]byte(r.URL.Path)):
			wfExec := &struct {
				Data WorkflowExecution `json:"data"`
			}{
				WorkflowExecution{Execution: Execution{ID: "1", StartDate: startDate, EndDate: endDate}},
			}

			b, err := json.Marshal(wfExec)
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

	type args struct {
		ctx                 context.Context
		appID               string
		envID               string
		nodeName            string
		requestedAttributes []string
	}
	tests := []struct {
		name           string
		args           args
		wantErr        bool
		expectedWfExec *WorkflowExecution
	}{
		{"GetLastWorkflowExecutionOK", args{context.Background(), "normal", "envID", "node1", []string{"attr1", "attr3"}}, false, &WorkflowExecution{Execution: Execution{ID: "1", StartDate: startDate, EndDate: endDate}}},
		{"GetLastWorkflowExecutionError", args{context.Background(), "error", "envID", "node1", nil}, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			d := &deploymentService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}
			wfExec, err := d.GetLastWorkflowExecution(tt.args.ctx, tt.args.appID, tt.args.envID)
			if (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.GetLastWorkflowExecution() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				assert.DeepEqual(t, wfExec, tt.expectedWfExec)
			}
		})
	}
}
