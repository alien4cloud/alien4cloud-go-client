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
	"strings"
	"time"

	"github.com/alien4cloud/alien4cloud-go-client/v3/alien4cloud"
)

// Command arguments
var url, user, password, appName string

func init() {
	// Initialize command arguments
	flag.StringVar(&url, "url", "http://localhost:8088", "Alien4Cloud URL")
	flag.StringVar(&user, "user", "admin", "User")
	flag.StringVar(&password, "password", "changeme", "Password")
	flag.StringVar(&appName, "app", "", "Name of the application to create")
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

	err = client.DeploymentService().UndeployApplication(ctx, appName, envID)
	if err != nil {
		log.Panic(err)
	}

	// Wait for the end of undeployment
	done := false
	log.Printf("Waiting for the end of undeployment...")
	var filters alien4cloud.LogFilter
	var deploymentStatus string
	logIndex := 0
	for !done {
		time.Sleep(5 * time.Second)

		a4cLogs, nbLogs, err := client.LogService().GetLogsOfApplication(ctx, appName, envID, filters, logIndex)
		if nbLogs > 0 {
			logIndex = logIndex + nbLogs
			for idx := 0; idx < nbLogs; idx++ {
				fmt.Printf("%s [%s][%s][%s][%s][%s][%s][%s] %s\n",
					a4cLogs[idx].Timestamp.Format(time.RFC3339),
					a4cLogs[idx].DeploymentPaaSID,
					a4cLogs[idx].Level,
					a4cLogs[idx].WorkflowID,
					a4cLogs[idx].NodeID,
					a4cLogs[idx].InstanceID,
					a4cLogs[idx].InterfaceName,
					a4cLogs[idx].OperationName,
					a4cLogs[idx].Content)
			}
		}

		status, err := client.DeploymentService().GetDeploymentStatus(ctx, appName, envID)
		if err != nil {
			log.Panic(err)
		}

		deploymentStatus = strings.ToLower(status)
		done = (deploymentStatus == alien4cloud.ApplicationUndeployed || deploymentStatus == alien4cloud.ApplicationError)
		if done {
			fmt.Printf("\nDeployment status: %s\n", status)
			break
		}
	}

	if deploymentStatus == alien4cloud.ApplicationUndeployed {
		// Now that the application is undeployed, deleting it
		err = client.ApplicationService().DeleteApplication(ctx, appName)
		if err != nil {
			log.Panic(err)
		}

		fmt.Printf("Application %s deleted\n", appName)
	}
}
