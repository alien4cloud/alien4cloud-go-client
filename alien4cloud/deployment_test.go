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
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

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
				client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
			}

			if err := d.UpdateApplication(tt.args.ctx, tt.args.appID, tt.args.envID); (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.UpdateApplication() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	cancelableCtx, cancelFn := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancelFn()
	d := &deploymentService{
		client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
	}

	if err := d.UpdateApplication(cancelableCtx, "cancel", "envID"); err == nil {
		t.Error("deploymentService.UpdateApplication() expecting an error")
	}

}

func Test_deploymentService_WaitUntilStateIs(t *testing.T) {
	closeCh := make(chan struct{})
	defer close(closeCh)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/applications/err/environments/.*/active-deployment-monitored`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{"deployment":{"id":"err"}}}`))
			return
		case regexp.MustCompile(`.*/applications/.*/environments/.*/active-deployment-monitored`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{"deployment":{"id":"myID"}}}`))
			return
		case regexp.MustCompile(`.*/deployments/err/status`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
			return
		case regexp.MustCompile(`.*/deployments/.*/status`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"data":"%s"}`, ApplicationDeployed)))
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
				client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
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
		client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
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
		case regexp.MustCompile(`.*/applications/error/environments/.*/workflows/.*`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusInternalServerError)
			return
		case regexp.MustCompile(`.*/applications/cancel/environments/.*/workflows/.*`).Match([]byte(r.URL.Path)):
			// wait until test are finish to simulate long running op that will be cancelled
			<-closeCh
			w.WriteHeader(http.StatusOK)
			return
		case regexp.MustCompile(`.*/applications/emptyExecID/environments/.*/workflows/.*`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":""}`))
			return
		case regexp.MustCompile(`.*/applications/badExecID/environments/.*/workflows/.*`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":`))
			return
		case regexp.MustCompile(`.*/applications/.*/environments/.*/workflows/.*`).Match([]byte(r.URL.Path)):
			matches := regexp.MustCompile(`.*/applications/(.*)/environments/.*/workflows/.*`).FindStringSubmatch(r.URL.Path)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"data":"%s"}`, matches[1])))
			return
		case regexp.MustCompile(`.*/executions/search`).Match([]byte(r.URL.Path)) && r.URL.Query().Get("query") == "execCancel":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{"types":["execution"],"data":[{"id":"7459ca00-f98f-47f1-a7e8-4d779d65253a","deploymentId":"4186a188-24a4-4910-9d7b-207ca09f98e3","workflowId":"stopWebServer","workflowName":"stopWebServer","displayWorkflowName":"stopWebServer","startDate":1578949107377,"endDate":1578949125749,"status":"RUNNING","hasFailedTasks":false}],"queryDuration":1,"totalResults":3,"from":1,"to":1,"facets":null},"error":null}`))
			return
		case regexp.MustCompile(`.*/executions/search`).Match([]byte(r.URL.Path)) && r.URL.Query().Get("query") == "badExecSearch":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{"types":["execution"],"data":[{"i`))
			return
		case regexp.MustCompile(`.*/executions/search`).Match([]byte(r.URL.Path)) && r.URL.Query().Get("query") == "noExecSearch":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{"types":[],"data":[],"queryDuration":1,"totalResults":0,"from":0,"to":0,"facets":null},"error":null}`))
			return
		case regexp.MustCompile(`.*/executions/search`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{"types":["execution"],"data":[{"id":"7459ca00-f98f-47f1-a7e8-4d779d65253a","deploymentId":"4186a188-24a4-4910-9d7b-207ca09f98e3","workflowId":"stopWebServer","workflowName":"stopWebServer","displayWorkflowName":"stopWebServer","startDate":1578949107377,"endDate":1578949125749,"status":"SUCCEEDED","hasFailedTasks":false}],"queryDuration":1,"totalResults":3,"from":1,"to":1,"facets":null},"error":null}`))
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
		want    *WorkflowExecution
		wantErr bool
	}{
		{"Normal", args{context.Background(), "app", "env", "wf", 5 * time.Minute},
			&WorkflowExecution{ID: "7459ca00-f98f-47f1-a7e8-4d779d65253a", DeploymentID: "4186a188-24a4-4910-9d7b-207ca09f98e3", WorkflowID: "stopWebServer", WorkflowName: "stopWebServer", DisplayWorkflowName: "stopWebServer", Status: "SUCCEEDED", HasFailedTasks: false},
			false,
		},
		{"EmptyExecID", args{context.Background(), "emptyExecID", "env", "wf", 5 * time.Minute}, nil, true},
		{"BadExecID", args{context.Background(), "badExecID", "env", "wf", 5 * time.Minute}, nil, true},
		{"BadExecSearch", args{context.Background(), "badExecSearch", "env", "wf", 5 * time.Minute}, nil, true},
		{"NoExecSearch", args{context.Background(), "noExecSearch", "env", "wf", 5 * time.Minute}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &deploymentService{
				client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
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
		client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
	}

	_, err := d.RunWorkflow(cancelableCtx, "cancel", "envID", "wf", 500*time.Millisecond)
	assert.ErrorContains(t, err, "context deadline exceeded")

	cancelableCtx, cancelFn = context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancelFn()
	_, err = d.RunWorkflow(cancelableCtx, "cancel", "envID", "wf", 50*time.Millisecond)
	assert.ErrorContains(t, err, "context deadline exceeded")

	cancelableCtx, cancelFn = context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancelFn()
	_, err = d.RunWorkflow(cancelableCtx, "execCancel", "envID", "wf", 50*time.Millisecond)
	assert.ErrorContains(t, err, "context deadline exceeded")

	cancelableCtx, cancelFn = context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancelFn()
	_, err = d.RunWorkflow(cancelableCtx, "execCancel", "envID", "wf", 50*time.Millisecond)
	assert.ErrorContains(t, err, "context deadline exceeded")
}
