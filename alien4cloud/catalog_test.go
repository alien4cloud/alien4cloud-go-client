package alien4cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
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

func Test_catalogService_GetComplexTOSCAType(t *testing.T) {
	expectedParsingErrors := make(map[string][]ParsingError)
	expectedParsingErrors["types.yaml"] = []ParsingError{
		{ErrorLevel: "ERROR", ErrorCode: "SOMETHING_RUDE", Problem: "ExpectedError"},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if regexp.MustCompile(`.*/formdescriptor/complex-tosca-type`).Match([]byte(r.URL.Path)) {
			var req ComplexToscaTypeDescriptorRequest
			rb, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Errorf("Failed to read request body %+v", r)
			}
			defer r.Body.Close()
			err = json.Unmarshal(rb, &req)
			if err != nil {
				t.Errorf("Failed to unmarshal request body %+v", r)
			}

			statusCode := http.StatusOK
			var resStr string
			propType := req.PropertyDefinition.Type
			if propType == "UnknownType" {
				statusCode = http.StatusBadRequest
				resStr = "{\"error\": \"bad request\"}"
			} else {
				resStr =
					"{\n" +
						"	\"data\": {\n" +
						"		\"_propertyType\": {\n" +
						"			\"prop1\": {\n" +
						"				\"" + TYPE_DESCRIPTION_CONTENT_TYPE_KEY + "\": {\n" +
						"					\"" + TYPE_DESCRIPTION_TOSCA_DEFINITION_KEY + "\": {\n" +
						"						\"definition\": true,\n" +
						"						\"description\": \"prop1 description\",\n" +
						"						\"password\": false,\n" +
						"						\"required\": true,\n" +
						"						\"type\": \"string\"\n" +
						"					},\n" +
						"					\"" + TYPE_DESCRIPTION_TYPE_KEY + "\": \"" + TYPE_DESCRIPTION_TOSCA_TYPE + "\"\n" +
						"				},\n" +
						"				\"" + TYPE_DESCRIPTION_TYPE_KEY + "\": \"" + TYPE_DESCRIPTION_ARRAY_TYPE + "\"\n" +
						"			},\n" +
						"			\"prop2\": {\n" +
						"				\"" + TYPE_DESCRIPTION_TOSCA_DEFINITION_KEY + "\": {\n" +
						"					\"definition\": true,\n" +
						"					\"description\": \"Prop2 description\",\n" +
						"					\"password\": false,\n" +
						"					\"required\": false,\n" +
						"					\"type\": \"string\"\n" +
						"				},\n" +
						"				\"" + TYPE_DESCRIPTION_TYPE_KEY + "\": \"" + TYPE_DESCRIPTION_TOSCA_TYPE + "\"\n" +
						"			}\n" +
						"		},\n" +
						"		\"" + TYPE_DESCRIPTION_TYPE_KEY + "\": \"" + TYPE_DESCRIPTION_COMPLEX_TYPE + "\"\n" +
						"	},\n" +
						"	\"error\": null\n" +
						"}"
			}
			w.WriteHeader(statusCode)
			_, _ = w.Write([]byte(resStr))
			return
		} else {
			// Should not go there
			t.Errorf("Unexpected call for request %+v", r)
		}

	}))
	defer ts.Close()

	type args struct {
		ctx     context.Context
		request ComplexToscaTypeDescriptorRequest
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"TestOK",
			args{context.Background(),
				ComplexToscaTypeDescriptorRequest{PropertyDefinition: PropertyDefinition{Type: "myType"}}},
			"data",
			false},
		{"TestOK",
			args{context.Background(),
				ComplexToscaTypeDescriptorRequest{PropertyDefinition: PropertyDefinition{Type: "UnknownType"}}},
			"",
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &catalogService{
				client: &a4cClient{client: http.DefaultClient, baseURL: ts.URL},
			}
			got, err := cs.GetComplexTOSCAType(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("catalogService.GetComplexTOSCAType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != "" {
				_, ok := got[tt.want]
				if !ok {
					t.Errorf("catalogService.GetComplexTOSCAType() has no key %s, got map %+v", tt.want, got)
				}
			}
		})
	}
}
