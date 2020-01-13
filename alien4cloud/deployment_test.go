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
	"time"
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
			w.Write([]byte(`{"data":"deployed"}`))
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
