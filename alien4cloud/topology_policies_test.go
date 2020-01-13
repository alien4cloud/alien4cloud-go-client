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

func Test_topologyService_AddPolicy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/editor/.*/execute`).Match([]byte(r.URL.Path)):
			var tepReq topologyEditorPolicies
			rb, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Errorf("Failed to read request body %+v", r)
			}
			defer r.Body.Close()
			s := string(rb)
			t.Logf("request: %s", s)

			err = json.Unmarshal(rb, &tepReq)
			if err != nil {
				t.Errorf("Failed to unmarshal request body %+v", r)
			}
			assert.Equal(t, tepReq.getOperationType(), "org.alien4cloud.tosca.editor.operations.policies.AddPolicyOperation")
			assert.Assert(t, "" != tepReq.PolicyName)
			assert.Assert(t, "" != tepReq.PolicyTypeID)
			if tepReq.PolicyName == "policy1withid" {
				assert.Assert(t, "" != tepReq.getPreviousOperationID())
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
			w.Write(b)
			return
		case regexp.MustCompile(`.*/applications/notfound/environments/.*/topology`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
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

	type args struct {
		ctx          context.Context
		a4cCtx       *TopologyEditorContext
		policyName   string
		policyTypeID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"NormalCase", args{context.Background(), &TopologyEditorContext{AppID: "app", EnvID: "env"}, "policy1", "policy:type"}, false},
		{"WithPreviousOpID", args{context.Background(), &TopologyEditorContext{AppID: "app", EnvID: "env", PreviousOperationID: "someid"}, "policy1withid", "policy:type"}, false},
		{"TopoNotFound", args{context.Background(), &TopologyEditorContext{AppID: "notfound", EnvID: "env"}, "policy1", "policy:type"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tServ := &topologyService{
				client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
			}
			if err := tServ.AddPolicy(tt.args.ctx, tt.args.a4cCtx, tt.args.policyName, tt.args.policyTypeID); (err != nil) != tt.wantErr {
				t.Errorf("topologyService.AddPolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_topologyService_DeletePolicy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/editor/.*/execute`).Match([]byte(r.URL.Path)):
			var tepReq topologyEditorPolicies
			rb, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Errorf("Failed to read request body %+v", r)
			}
			defer r.Body.Close()
			s := string(rb)
			t.Logf("request: %s", s)

			err = json.Unmarshal(rb, &tepReq)
			if err != nil {
				t.Errorf("Failed to unmarshal request body %+v", r)
			}
			assert.Equal(t, tepReq.getOperationType(), "org.alien4cloud.tosca.editor.operations.policies.DeletePolicyOperation")
			assert.Assert(t, "" != tepReq.PolicyName)
			if tepReq.PolicyName == "policy1withid" {
				assert.Assert(t, "" != tepReq.getPreviousOperationID())
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
			w.Write(b)
			return
		case regexp.MustCompile(`.*/applications/notfound/environments/.*/topology`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"code": 404,"message":"not found"}}`))
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

	type args struct {
		ctx        context.Context
		a4cCtx     *TopologyEditorContext
		policyName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"NormalCase", args{context.Background(), &TopologyEditorContext{AppID: "app", EnvID: "env"}, "policy1"}, false},
		{"WithPreviousOpID", args{context.Background(), &TopologyEditorContext{AppID: "app", EnvID: "env", PreviousOperationID: "someid"}, "policy1withid"}, false},
		{"TopoNotFound", args{context.Background(), &TopologyEditorContext{AppID: "notfound", EnvID: "env"}, "policy1"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tServ := &topologyService{
				client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
			}
			if err := tServ.DeletePolicy(tt.args.ctx, tt.args.a4cCtx, tt.args.policyName); (err != nil) != tt.wantErr {
				t.Errorf("topologyService.DeletePolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
