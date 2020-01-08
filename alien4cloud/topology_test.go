package alien4cloud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func Test_topologyService_CreateAndDeleteWorkflow(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/editor/.*/execute`).Match([]byte(r.URL.Path)):
			var resExec struct {
				Data struct {
					LastOperationIndex int `json:"lastOperationIndex"`
					Operations         []struct {
						PreviousOperationID string `json:"id"`
					} `json:"operations"`
				} `json:"data"`
			}
			resExec.Data.LastOperationIndex = 0
			resExec.Data.Operations = []struct {
				PreviousOperationID string "json:\"id\""
			}{
				struct {
					PreviousOperationID string "json:\"id\""
				}{PreviousOperationID: "0"},
			}
			resExec.Data.Operations = nil

			b, err := json.Marshal(&resExec)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(b)
			return
		case regexp.MustCompile(`.*/applications/.*/environments/.*/topology`).Match([]byte(r.URL.Path)):
			var res struct {
				Data string `json:"data"`
			}
			res.Data = "tid"
			b, err := json.Marshal(&res)
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
	defer ts.Close()

	type args struct {
		ctx           context.Context
		editorContext *TopologyEditorContext
		workflowName  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"DefaultCreate", args{context.Background(), &TopologyEditorContext{AppID: "test", EnvID: "test", TopologyID: "tid"}, "wfName"}, false},
		{"NoTopologyID", args{context.Background(), &TopologyEditorContext{AppID: "test", EnvID: "test"}, "wfName"}, false},
		{"NilEditorContext", args{context.Background(), nil, "wfName"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tSrv := &topologyService{
				client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
			}
			err := tSrv.CreateWorkflow(tt.args.ctx, tt.args.editorContext, tt.args.workflowName)
			if (err != nil) != tt.wantErr {
				t.Errorf("catalogService.CreateWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tSrv := &topologyService{
				client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
			}
			err := tSrv.DeleteWorkflow(tt.args.ctx, tt.args.editorContext, tt.args.workflowName)
			if (err != nil) != tt.wantErr {
				t.Errorf("catalogService.DeleteWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}
