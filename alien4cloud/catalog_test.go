package alien4cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func Test_catalogService_UploadCSAR(t *testing.T) {
	expectedParsingErrors := make(map[string][]ParsingError)
	expectedParsingErrors["types.yaml"] = []ParsingError{
		{ErrorLevel: "ERROR", ErrorCode: "SOMETHING_RUDE", Problem: "ExpectedError"},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !strings.HasPrefix(mediaType, "multipart/") {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		workspace := r.URL.Query().Get("workspace")
		perr := make(map[string][]ParsingError)
		switch workspace {
		case "error":
			w.WriteHeader(http.StatusInternalServerError)
			return
		case "parsing_error":
			perr = expectedParsingErrors
		default:
		}

		var res struct {
			Data struct {
				CSAR   CSAR                      `json:"csar,omitempty"`
				Errors map[string][]ParsingError `json:"errors,omitempty"`
			} `json:"data"`
		}

		res.Data.CSAR = CSAR{
			ID:        "mycsar",
			Workspace: workspace,
		}
		res.Data.Errors = perr

		b, err := json.Marshal(&res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(b)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}))
	defer ts.Close()

	type args struct {
		ctx       context.Context
		csar      io.Reader
		workspace string
	}
	tests := []struct {
		name    string
		args    args
		want    CSAR
		wantErr bool
	}{
		{"DefaultCase",
			args{context.Background(), &bytes.Reader{}, ""},
			CSAR{ID: "mycsar"},
			false},
		{"ParsingError",
			args{context.Background(), &bytes.Reader{}, "parsing_error"},
			CSAR{ID: "mycsar", Workspace: "parsing_error"},
			true},
		{"Error",
			args{context.Background(), &bytes.Reader{}, "error"},
			CSAR{},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &catalogService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}
			got, err := cs.UploadCSAR(tt.args.ctx, tt.args.csar, tt.args.workspace)
			if (err != nil) != tt.wantErr {
				t.Errorf("catalogService.UploadCSAR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("catalogService.UploadCSAR() = %v, want %v", got, tt.want)
			}
		})
	}
}
