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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"gotest.tools/v3/assert"
)

func Test_userService_TestCreateUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/users`).Match([]byte(r.URL.Path)):
			var req CreateUserRequest
			rb, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Errorf("Failed to read request body %+v", r)
			}
			defer r.Body.Close()

			err = json.Unmarshal(rb, &req)
			if err != nil {
				t.Errorf("Failed to unmarshal request body %+v", r)
			}
			if req.Username == "" {
				var res struct {
					Error Error `json:"error"`
				}
				res.Error.Code = http.StatusInternalServerError
				res.Error.Message = "Method argument is invalid"
				b, err := json.Marshal(&res)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.WriteHeader(http.StatusNotImplemented)
					_, _ = w.Write(b)
				}
			} else {
				w.WriteHeader(http.StatusOK)
			}
		default:
			t.Errorf("Unexpected request %s", r.URL.Path)
		}
	}))

	type args struct {
		ctx           context.Context
		createRequest CreateUserRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"UndefinedUserName", args{context.Background(),
			CreateUserRequest{Username: "", Password: "passwd"}}, true},
		{"DefinedUserName", args{context.Background(),
			CreateUserRequest{Username: "user1", Password: "passwd"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uServ := &userService{
				client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
			}
			if err := uServ.CreateUser(tt.args.ctx, tt.args.createRequest); (err != nil) != tt.wantErr {
				t.Errorf("userService.CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_userService_TestUpdateUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/users/wronguser`).Match([]byte(r.URL.Path)):
			defer r.Body.Close()

			var res struct {
				Error Error `json:"error"`
			}
			res.Error.Code = http.StatusGatewayTimeout
			res.Error.Message = "User [wronguser] cannot be found"
			b, err := json.Marshal(&res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusGatewayTimeout)
				_, _ = w.Write(b)
			}
		case regexp.MustCompile(`.*/users/.*`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("Unexpected request %s", r.URL.Path)
		}
	}))

	type args struct {
		ctx           context.Context
		username      string
		updateRequest UpdateUserRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"NotExistingUser", args{context.Background(), "wronguser",
			UpdateUserRequest{FirstName: "unknown", Password: "passwd"}}, true},
		{"ExistingUser", args{context.Background(), "user1",
			UpdateUserRequest{FirstName: "user1", Password: "passwd"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uServ := &userService{
				client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
			}
			if err := uServ.UpdateUser(tt.args.ctx, tt.args.username, tt.args.updateRequest); (err != nil) != tt.wantErr {
				t.Errorf("userService.CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_userService_TestGetUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case regexp.MustCompile(`.*/users/expectedUser`).Match([]byte(r.URL.Path)):
			defer r.Body.Close()

			var res struct {
				Data  User  `json:"data"`
				Error Error `json:"error"`
			}
			res.Data.Username = "expectedUser"
			b, err := json.Marshal(&res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(b)
			}
		case regexp.MustCompile(`.*/users/.*`).Match([]byte(r.URL.Path)):
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("Unexpected request %s", r.URL.Path)
		}
	}))

	type args struct {
		ctx      context.Context
		username string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"ExistingUser", args{context.Background(), "expectedUser"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uServ := &userService{
				client: restClient{Client: http.DefaultClient, baseURL: ts.URL},
			}
			userResp, err := uServ.GetUser(tt.args.ctx, tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("userService.GetUser() error = %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				assert.Equal(t, tt.args.username, userResp.Username, "Unexpected result for GetUser: %+v", userResp)
			}

		})
	}
}
