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
	"fmt"
	"log"
	"time"

	"github.com/alien4cloud/alien4cloud-go-client/v2/alien4cloud"
)

const workflowExecutionStartTimeoutInSeconds = 60

// Command arguments
var url, user, password, appName, workflow string

func init() {
	// Initialize command arguments
	flag.StringVar(&url, "url", "http://localhost:8088", "Alien4Cloud URL")
	flag.StringVar(&user, "user", "admin", "User")
	flag.StringVar(&password, "password", "changeme", "Password")
	flag.StringVar(&appName, "app", "", "Name of the application to create")
	flag.StringVar(&workflow, "workflow", "", "Name of the workflow to run")
}

func main() {

	// Parsing command arguments
	flag.Parse()

	// Check required parameters
	if appName == "" {
		log.Panic("Mandatory argument 'app' missing (Name of the application to create)")
	}
	if workflow == "" {
		log.Panic("Mandatory argument 'workflow' missing (Name of the workflow to run)")
	}

	client, err := alien4cloud.NewClient(url, user, password, "", true)
	if err != nil {
		log.Panic(err)
	}

	err = client.Login()
	if err != nil {
		log.Panic(err)
	}

	envID, err := client.GetEnvironmentIDbyName(appName, alien4cloud.DefaultEnvironmentName)
	if err != nil {
		log.Panic(err)
	}

	workflowExecution, err := client.RunWorkflow(appName, envID, workflow, workflowExecutionStartTimeoutInSeconds)
	if err != nil {
		log.Panic(err)
	}

	// Wait for the end of deployment
	done := false
	log.Printf("Waiting for the end of workflow execution...")
	filters := alien4cloud.LogFilter{
		ExecutionID: []string{workflowExecution.ID},
	}
	var workflowStatus string
	logIndex := 0
	for !done {
		time.Sleep(5 * time.Second)

		a4cLogs, nbLogs, err := client.GetLogsOfApplication(appName, envID, filters, logIndex)
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

		workflowExecution, err = client.GetLastWorkflowExecution(appName, envID)
		if err != nil {
			log.Panic(err)
		}
		workflowStatus = workflowExecution.Status
		done = (workflowStatus == alien4cloud.WorkflowSucceeded || workflowStatus == alien4cloud.WorkflowFailed)
		if done {
			fmt.Printf("\nWorkflow status: %s\n", workflowStatus)
			break
		}
	}
}
