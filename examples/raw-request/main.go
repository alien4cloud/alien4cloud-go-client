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

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/alien4cloud/alien4cloud-go-client/v2/alien4cloud"
)

// Command arguments
var url, user, password string

func init() {
	// Initialize command arguments
	flag.StringVar(&url, "url", "http://localhost:8088", "Alien4Cloud URL")
	flag.StringVar(&user, "user", "admin", "User")
	flag.StringVar(&password, "password", "changeme", "Password")

}

func main() {

	// Parsing command arguments
	flag.Parse()

	client, err := alien4cloud.NewClient(url, user, password, "", true)
	if err != nil {
		log.Panic(err)
	}

	// Timeout after one minute (this is optional you can use a context without timeout or cancelation)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err = client.Login(ctx)
	if err != nil {
		log.Panic(err)
	}

	request, err := client.NewRequest(ctx, "GET", "/rest/v1/auth/status", nil)
	if err != nil {
		log.Panic(err)
	}

	response, err := client.Do(request)

	var res struct {
		Data struct {
			AuthSystem     string   `json:"authSystem"`
			GithubUsername string   `json:"githubUsername"`
			Groups         []string `json:"groups"`
			IsLogged       bool     `json:"isLogged"`
			Roles          []string `json:"roles"`
			Username       string   `json:"username"`
		} `json:"data"`
	}

	err = alien4cloud.ReadA4CResponse(response, &res)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("User %s:\n", res.Data.Username)
	fmt.Printf("\tGithubUsername %q\n", res.Data.GithubUsername)
	fmt.Printf("\tAuthSystem %q\n", res.Data.AuthSystem)
	fmt.Printf("\tIsLogged \"%v\"\n", res.Data.IsLogged)
	if len(res.Data.Groups) > 0 {
		fmt.Printf("\tGroups:\n")
		for _, g := range res.Data.Groups {
			fmt.Printf("\t\t- %q\n", g)
		}
	}
	if len(res.Data.Roles) > 0 {
		fmt.Printf("\tRoles:\n")
		for _, r := range res.Data.Roles {
			fmt.Printf("\t\t- %q\n", r)
		}
	}

}
