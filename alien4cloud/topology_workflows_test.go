package alien4cloud

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func Test_topologyService_AddWorkflowActivity(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/editor/.*/execute`).Match([]byte(r.URL.Path)):
			var awaReq addWorkflowActivityReq
			rb, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Errorf("Failed to read request body %+v", r)
			}
			defer r.Body.Close()
			s := string(rb)
			t.Logf("request: %s", s)

			err = json.Unmarshal(rb, &awaReq)
			if err != nil {
				t.Errorf("Failed to unmarshal request body %+v", r)
			}

			switch awaReq.Activity.Type {
			case InlineWorkflowActivityType:
				if awaReq.Activity.Inline == "" {
					t.Error("Missing inline workflow name")
				}
			case SetStateWorkflowActivityType:
				if awaReq.Activity.StateName == "" {
					t.Error("Missing inline State name")
				}
				if awaReq.Target == "" {
					t.Error("Missing target name")
				}
			case CallOperationWorkflowActivityType:
				if awaReq.Activity.InterfaceName == "" {
					t.Error("Missing inline interface name")
				}
				if awaReq.Activity.OperationName == "" {
					t.Error("Missing inline operation name")
				}
				if awaReq.Target == "" {
					t.Error("Missing target name")
				}
			}
			if awaReq.RelatedStepID != "" && awaReq.Before == nil {
				t.Error("Missing before switch for related step ID")
			}
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
			_, _ = w.Write(b)
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
			_, _ = w.Write(b)
			return
		}

		// Should not go there
		t.Errorf("Unexpected call for request %+v", r)
	}))

	wrongActivity := WorkflowActivity{
		activitytype: "WrongActivity",
	}
	type args struct {
		ctx          context.Context
		a4cCtx       *TopologyEditorContext
		workflowName string
		activity     *WorkflowActivity
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"AddInlineWorkflow", args{context.Background(),
			&TopologyEditorContext{AppID: "test", EnvID: "test", TopologyID: "tid"}, "wf",
			newWfActivity().InlineWorkflow("inlineWF")}, false},
		{"AddSetState", args{context.Background(),
			&TopologyEditorContext{AppID: "test", EnvID: "test", TopologyID: "tid"}, "wf",
			newWfActivity().SetState("mynode", "myState").AppendAfter("myotherStep")}, false},
		{"AddCallOp", args{context.Background(),
			&TopologyEditorContext{AppID: "test", EnvID: "test", TopologyID: "tid"}, "wf",
			newWfActivity().OperationCall("mynode", "rel", "ifce", "opName").InsertBefore("myotherStep")}, false},
		{"AddWrongActivity", args{context.Background(),
			&TopologyEditorContext{AppID: "test", EnvID: "test", TopologyID: "tid"}, "wf",
			wrongActivity.InsertBefore("myotherStep")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tSrv := &topologyService{
				client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
			}
			if err := tSrv.AddWorkflowActivity(tt.args.ctx, tt.args.a4cCtx, tt.args.workflowName, tt.args.activity); (err != nil) != tt.wantErr {
				t.Errorf("topologyService.AddWorkflowActivity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func newWfActivity() *WorkflowActivity {
	return &WorkflowActivity{}
}

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
			_, _ = w.Write(b)
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
			_, _ = w.Write(b)
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
