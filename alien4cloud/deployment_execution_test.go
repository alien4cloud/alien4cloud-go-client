package alien4cloud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"regexp"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func mustParseTime(t *testing.T, timeStr string) Time {

	r, err := time.Parse("2006-01-02 15:04:05.000 -0700 MST", timeStr)
	assert.NilError(t, err, "failed to parse time")

	return Time{r}
}

func Test_deploymentService_GetExecutions(t *testing.T) {
	closeCh := make(chan struct{})
	defer close(closeCh)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch {
		case regexp.MustCompile(`.*/applications/app/environments/.*/active-deployment-monitored`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"deployment":{"id":"7459ca00-f98f-47f1-a7e8-4d779d65253a"}}}`))
			return
		case regexp.MustCompile(`.*/executions/search.*`).Match([]byte(r.URL.Path)):

			assert.Assert(t, "" != r.URL.Query().Get("deploymentId"))
			assert.Assert(t, "" != r.URL.Query().Get("from"))
			assert.Assert(t, "" != r.URL.Query().Get("size"))

			switch r.URL.Query().Get("deploymentId") {
			case "normal":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data":{"types":["execution"],"data":[{"id":"7459ca00-f98f-47f1-a7e8-4d779d65253a","deploymentId":"4186a188-24a4-4910-9d7b-207ca09f98e3","workflowId":"stopWebServer","workflowName":"stopWebServer","displayWorkflowName":"stopWebServer","startDate":1578949107377,"endDate":1578949125749,"status":"SUCCEEDED","hasFailedTasks":false}],"queryDuration":1,"totalResults":3,"from":1,"to":1,"facets":null},"error":null}`))
				return
			case "query":
				assert.Assert(t, "" != r.URL.Query().Get("query"))
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data":{"types":["execution"],"data":[{"id":"7459ca00-f98f-47f1-a7e8-4d779d65253a","deploymentId":"4186a188-24a4-4910-9d7b-207ca09f98e3","workflowId":"stopWebServer","workflowName":"stopWebServer","displayWorkflowName":"stopWebServer","startDate":1578949107377,"endDate":1578949125749,"status":"SUCCEEDED","hasFailedTasks":false}],"queryDuration":1,"totalResults":1,"from":0,"to":0,"facets":null},"error":null}`))
				return
			case "multi":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"data":{"types":["execution","execution","execution"],"data":[{"id":"d9f63781-5245-4cd0-a24c-b83d4c4842f1","deploymentId":"4186a188-24a4-4910-9d7b-207ca09f98e3","workflowId":"startWebServer","workflowName":"startWebServer","displayWorkflowName":"startWebServer","startDate":1578949107377,"endDate":1578949125749,"status":"SUCCEEDED","hasFailedTasks":false},{"id":"7459ca00-f98f-47f1-a7e8-4d779d65253a","deploymentId":"4186a188-24a4-4910-9d7b-207ca09f98e3","workflowId":"stopWebServer","workflowName":"stopWebServer","displayWorkflowName":"stopWebServer","startDate":1578949107377,"endDate":1578949125749,"status":"SUCCEEDED","hasFailedTasks":false},{"id":"e8cbb5bd-5f85-408e-9190-caee179d0581","deploymentId":"4186a188-24a4-4910-9d7b-207ca09f98e3","workflowId":"install","workflowName":"install","displayWorkflowName":"install","startDate":1578949107377,"endDate":1578949125749,"status":"SUCCEEDED","hasFailedTasks":false}],"queryDuration":1,"totalResults":3,"from":0,"to":2,"facets":null},"error":null}`))
				return
			case "error":
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
				return
			case "internalerror":
				<-closeCh
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}))

	type args struct {
		ctx          context.Context
		deploymentID string
		query        string
		from         int
		size         int
	}
	tests := []struct {
		name    string
		args    args
		want    []Execution
		want1   FacetedSearchResult
		wantErr bool
	}{
		{"normal", args{context.Background(), "normal", "", 1, 1},
			[]Execution{
				{ID: "7459ca00-f98f-47f1-a7e8-4d779d65253a", DeploymentID: "4186a188-24a4-4910-9d7b-207ca09f98e3", WorkflowID: "stopWebServer", WorkflowName: "stopWebServer", DisplayWorkflowName: "stopWebServer", Status: "SUCCEEDED", HasFailedTasks: false, StartDate: mustParseTime(t, "2020-01-13 21:58:27.377 +0100 CET"), EndDate: mustParseTime(t, "2020-01-13 21:58:45.749 +0100 CET")},
			},
			FacetedSearchResult{TotalResults: 3, From: 1, To: 1},
			false,
		},
		{"query", args{context.Background(), "query", "7459ca00-f98f-47f1-a7e8-4d779d65253a", 0, 1},
			[]Execution{
				{ID: "7459ca00-f98f-47f1-a7e8-4d779d65253a", DeploymentID: "4186a188-24a4-4910-9d7b-207ca09f98e3", WorkflowID: "stopWebServer", WorkflowName: "stopWebServer", DisplayWorkflowName: "stopWebServer", Status: "SUCCEEDED", HasFailedTasks: false, StartDate: mustParseTime(t, "2020-01-13 21:58:27.377 +0100 CET"), EndDate: mustParseTime(t, "2020-01-13 21:58:45.749 +0100 CET")},
			},
			FacetedSearchResult{TotalResults: 1, From: 0, To: 0},
			false,
		},
		{"multi", args{context.Background(), "multi", "", 0, 10},
			[]Execution{
				{ID: "d9f63781-5245-4cd0-a24c-b83d4c4842f1", DeploymentID: "4186a188-24a4-4910-9d7b-207ca09f98e3", WorkflowID: "startWebServer", WorkflowName: "startWebServer", DisplayWorkflowName: "startWebServer", Status: "SUCCEEDED", HasFailedTasks: false, StartDate: mustParseTime(t, "2020-01-13 21:58:27.377 +0100 CET"), EndDate: mustParseTime(t, "2020-01-13 21:58:45.749 +0100 CET")},
				{ID: "7459ca00-f98f-47f1-a7e8-4d779d65253a", DeploymentID: "4186a188-24a4-4910-9d7b-207ca09f98e3", WorkflowID: "stopWebServer", WorkflowName: "stopWebServer", DisplayWorkflowName: "stopWebServer", Status: "SUCCEEDED", HasFailedTasks: false, StartDate: mustParseTime(t, "2020-01-13 21:58:27.377 +0100 CET"), EndDate: mustParseTime(t, "2020-01-13 21:58:45.749 +0100 CET")},
				{ID: "e8cbb5bd-5f85-408e-9190-caee179d0581", DeploymentID: "4186a188-24a4-4910-9d7b-207ca09f98e3", WorkflowID: "install", WorkflowName: "install", DisplayWorkflowName: "install", Status: "SUCCEEDED", HasFailedTasks: false, StartDate: mustParseTime(t, "2020-01-13 21:58:27.377 +0100 CET"), EndDate: mustParseTime(t, "2020-01-13 21:58:45.749 +0100 CET")},
			},
			FacetedSearchResult{TotalResults: 3, From: 0, To: 2},
			false,
		},
		{"error", args{context.Background(), "error", "", 0, 10}, nil, FacetedSearchResult{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &deploymentService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}
			got, got1, err := d.GetExecutions(tt.args.ctx, tt.args.deploymentID, tt.args.query, tt.args.from, tt.args.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.GetExecutions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.DeepEqual(t, got, tt.want)
			assert.DeepEqual(t, got1, tt.want1)
		})
	}

	ctx, cf := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cf()
	d := &deploymentService{
		client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
	}
	got, got1, err := d.GetExecutions(ctx, "internalerror", "", 0, 10)
	assert.ErrorContains(t, err, "context deadline exceeded")
	assert.Assert(t, got == nil)
	assert.DeepEqual(t, got1, FacetedSearchResult{})
}

func Test_deploymentService_GetExecutionByID(t *testing.T) {
	closeCh := make(chan struct{})
	defer close(closeCh)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		execID := path.Base(r.URL.Path)

		if execID == "NotFound" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var resExec struct {
			Execution `json:"data"`
		}
		resExec.Execution = Execution{
			ID:                  execID,
			DeploymentID:        "depID",
			WorkflowID:          "wf",
			WorkflowName:        "wf",
			DisplayWorkflowName: "wf",
			Status:              "SUCCEEDED",
			HasFailedTasks:      false,
			StartDate:           mustParseTime(t, "2020-01-13 21:58:27.377 +0100 CET"),
			EndDate:             mustParseTime(t, "2020-01-13 21:58:45.749 +0100 CET"),
		}

		b, err := json.Marshal(&resExec)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)

	}))

	type args struct {
		ctx         context.Context
		executionID string
	}
	tests := []struct {
		name    string
		args    args
		want    Execution
		wantErr bool
	}{
		{"success",
			args{context.Background(), "depID51"},
			Execution{ID: "depID51", DeploymentID: "depID", WorkflowID: "wf", WorkflowName: "wf", DisplayWorkflowName: "wf", Status: "SUCCEEDED", HasFailedTasks: false, StartDate: mustParseTime(t, "2020-01-13 21:58:27.377 +0100 CET"), EndDate: mustParseTime(t, "2020-01-13 21:58:45.749 +0100 CET")},
			false,
		},
		{"failure",
			args{context.Background(), "NotFound"},
			Execution{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &deploymentService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}
			got, err := d.GetExecutionByID(tt.args.ctx, tt.args.executionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.GetExecutionByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.DeepEqual(t, got, tt.want)
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &deploymentService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}
			got, err := d.GetExecution(tt.args.ctx, "", "", tt.args.executionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("deploymentService.GetExecution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.DeepEqual(t, got, tt.want)
		})
	}
}
