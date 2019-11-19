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
	"flag"
	"log"

	"github.com/alien4cloud/alien4cloud-go-client/v2/alien4cloud"
)

// Command arguments
var url, user, password, appName, appTemplate, orchestratorName, locationName string

func init() {
	// Initialize command arguments
	flag.StringVar(&url, "url", "http://localhost:8080", "Alien4Cloud URL")
	flag.StringVar(&user, "user", "admin", "User")
	flag.StringVar(&password, "password", "changeme", "Password")
	flag.StringVar(&appName, "app", "", "Name of the application to create")
	flag.StringVar(&appTemplate, "template", "", "Name of the topology template to use")
	flag.StringVar(&locationName, "location", "", "Name of the location where to deploy the application")
}

func main() {

	// Parsing command arguments
	flag.Parse()

	// Check required parameters
	if appName == "" {
		log.Panic("Mandatory argument 'app' missing (Name of the application to create)")
	}
	if appTemplate == "" {
		log.Panic("Mandatory argument 'template' missing (Name of the topology template to use)")
	}

	client, err := alien4cloud.NewClient(url, user, password, 0, "", true)
	if err != nil {
		log.Panic(err)
	}

	err = client.Login()
	if err != nil {
		log.Panic(err)
	}

	appID, err := client.CreateAppli(appName, appTemplate)
	if err != nil {
		log.Panic(err)
	}

	envID, err := client.GetEnvironmentIDbyName(appID, alien4cloud.DefaultEnvironmentName)
	if err != nil {
		log.Panic(err)
	}

	err = client.DeployApplication(appID, envID, locationName)
	if err != nil {
		log.Panic(err)
	}
}
