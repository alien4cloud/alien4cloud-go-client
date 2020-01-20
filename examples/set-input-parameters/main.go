// Copyright 2020 Bull S.A.S. Atos Technologies - Bull, Rue Jean Jaures, B.P.68, 78340, Les Clayes-sous-Bois, France.
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

	"github.com/alien4cloud/alien4cloud-go-client/v2/alien4cloud"
)

// Command arguments
var url, user, password, appName, propName, propValue, artifactName, artifactFilePath string

func init() {
	// Initialize command arguments
	flag.StringVar(&url, "url", "http://localhost:8088", "Alien4Cloud URL")
	flag.StringVar(&user, "user", "admin", "User")
	flag.StringVar(&password, "password", "changeme", "Password")
	flag.StringVar(&appName, "app", "", "Name of the application to create")
	flag.StringVar(&propName, "property", "", "Name of the input property to set")
	flag.StringVar(&propValue, "value", "", "Value of the input property to set")
	flag.StringVar(&artifactName, "artifact", "", "Name of the input artifact to set")
	flag.StringVar(&artifactFilePath, "file", "", "Path of the input artifact file")
}

func main() {

	// Parsing command arguments
	flag.Parse()

	// Check required parameter
	if appName == "" {
		log.Panic("Mandatory argument 'app' missing (Name of the application to delete)")
	}

	client, err := alien4cloud.NewClient(url, user, password, "", true)
	if err != nil {
		log.Panic(err)
	}

	// Timeout after one hour (this is optional you can use a context without timeout or cancelation)
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	err = client.Login(ctx)
	if err != nil {
		log.Panic(err)
	}

	envID, err := client.ApplicationService().GetEnvironmentIDbyName(ctx, appName, alien4cloud.DefaultEnvironmentName)
	if err != nil {
		log.Panic(err)
	}

	// Get input propoerties
	topology, err := client.TopologyService().GetTopology(ctx, appName, envID)
	if err != nil {
		log.Panic(err)
	}

	if propName != "" {
		inputProperties := topology.Data.Topology.Inputs

		if _, ok := inputProperties[propName]; !ok {
			log.Panicf("No such input property %s defined in application", propName)
		}

		updateRequest := alien4cloud.UpdateDeploymentTopologyRequest{
			InputProperties: map[string]interface{}{
				propName: propValue,
			},
		}

		err = client.DeploymentService().UpdateDeploymentTopology(ctx, appName, envID, updateRequest)
		if err != nil {
			log.Panic(err)
		}
	}

	if artifactName != "" {

		inputArtifacts := topology.Data.Topology.InputArtifacts

		if _, ok := inputArtifacts[artifactName]; !ok {
			log.Panicf("No such input artifact %s defined in application", artifactName)
		}

		err = client.DeploymentService().UploadDeploymentInputArtifact(ctx, appName, envID, artifactName, artifactFilePath)
		if err != nil {
			log.Panic(err)
		}
	}
}
