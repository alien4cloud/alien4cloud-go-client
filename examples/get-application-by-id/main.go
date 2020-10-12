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
	"log"
	"time"

	"github.com/alien4cloud/alien4cloud-go-client/v3/alien4cloud"
)

// Command arguments
var url, user, password, appID string

func init() {
	// Initialize command arguments
	flag.StringVar(&url, "url", "http://localhost:8088", "Alien4Cloud URL")
	flag.StringVar(&user, "user", "admin", "User")
	flag.StringVar(&password, "password", "changeme", "Password")
	flag.StringVar(&appID, "id", "", "ID of the application to find")
}

func main() {

	// Parsing command arguments
	flag.Parse()

	// Check required parameters
	if appID == "" {
		log.Panic("Mandatory argument 'id' missing (ID of the application to find)")
	}

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

	app, err := client.ApplicationService().GetApplicationByID(ctx, appID)
	if err != nil {
		log.Panic(err)
	}

	// Wait for the end of deployment
	log.Printf("Found application:\n")
	log.Printf("  ID:   %s\n", app.ID)
	log.Printf("  Name: %s\n", app.Name)
	log.Printf("  Tags: %v\n", app.Tags)

}
