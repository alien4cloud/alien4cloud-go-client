package alien4cloud

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func Test_deploymentService_GetExecutions(t *testing.T) {
	closeCh := make(chan struct{})
	defer close(closeCh)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			_, _ = w.Write([]byte(`{"data":{"types":["execution","execution","execution"],"data":[{"id":"d9f63781-5245-4cd0-a24c-b83d4c4842f1","deploymentId":"4186a188-24a4-4910-9d7b-207ca09f98e3","workflowId":"startWebServer","workflowName":"startWebServer","displayWorkflowName":"startWebServer","startDate":1578951354540,"endDate":1578951378035,"status":"SUCCEEDED","hasFailedTasks":false},{"id":"7459ca00-f98f-47f1-a7e8-4d779d65253a","deploymentId":"4186a188-24a4-4910-9d7b-207ca09f98e3","workflowId":"stopWebServer","workflowName":"stopWebServer","displayWorkflowName":"stopWebServer","startDate":1578949107377,"endDate":1578949125749,"status":"SUCCEEDED","hasFailedTasks":false},{"id":"e8cbb5bd-5f85-408e-9190-caee179d0581","deploymentId":"4186a188-24a4-4910-9d7b-207ca09f98e3","workflowId":"install","workflowName":"install","displayWorkflowName":"install","startDate":1578933372461,"endDate":1578933443757,"status":"SUCCEEDED","hasFailedTasks":false}],"queryDuration":1,"totalResults":3,"from":0,"to":2,"facets":null},"error":null}`))
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
				Execution{ID: "7459ca00-f98f-47f1-a7e8-4d779d65253a", DeploymentID: "4186a188-24a4-4910-9d7b-207ca09f98e3", WorkflowID: "stopWebServer", WorkflowName: "stopWebServer", DisplayWorkflowName: "stopWebServer", Status: "SUCCEEDED", HasFailedTasks: false},
			},
			FacetedSearchResult{TotalResults: 3, From: 1, To: 1},
			false,
		},
		{"query", args{context.Background(), "query", "7459ca00-f98f-47f1-a7e8-4d779d65253a", 0, 1},
			[]Execution{
				Execution{ID: "7459ca00-f98f-47f1-a7e8-4d779d65253a", DeploymentID: "4186a188-24a4-4910-9d7b-207ca09f98e3", WorkflowID: "stopWebServer", WorkflowName: "stopWebServer", DisplayWorkflowName: "stopWebServer", Status: "SUCCEEDED", HasFailedTasks: false},
			},
			FacetedSearchResult{TotalResults: 1, From: 0, To: 0},
			false,
		},
		{"multi", args{context.Background(), "multi", "", 0, 10},
			[]Execution{
				Execution{ID: "d9f63781-5245-4cd0-a24c-b83d4c4842f1", DeploymentID: "4186a188-24a4-4910-9d7b-207ca09f98e3", WorkflowID: "startWebServer", WorkflowName: "startWebServer", DisplayWorkflowName: "startWebServer", Status: "SUCCEEDED", HasFailedTasks: false},
				Execution{ID: "7459ca00-f98f-47f1-a7e8-4d779d65253a", DeploymentID: "4186a188-24a4-4910-9d7b-207ca09f98e3", WorkflowID: "stopWebServer", WorkflowName: "stopWebServer", DisplayWorkflowName: "stopWebServer", Status: "SUCCEEDED", HasFailedTasks: false},
				Execution{ID: "e8cbb5bd-5f85-408e-9190-caee179d0581", DeploymentID: "4186a188-24a4-4910-9d7b-207ca09f98e3", WorkflowID: "install", WorkflowName: "install", DisplayWorkflowName: "install", Status: "SUCCEEDED", HasFailedTasks: false},
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
