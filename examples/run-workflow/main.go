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

	"github.com/alien4cloud/alien4cloud-go-client/v2/alien4cloud"
)

// Command arguments
var url, user, password, appName, workflow string
var showEvents bool

func init() {
	// Initialize command arguments
	flag.StringVar(&url, "url", "http://localhost:8088", "Alien4Cloud URL")
	flag.StringVar(&user, "user", "admin", "User")
	flag.StringVar(&password, "password", "changeme", "Password")
	flag.StringVar(&appName, "app", "", "Name of the application to create")
	flag.StringVar(&workflow, "workflow", "", "Name of the workflow to run")
	flag.BoolVar(&showEvents, "events", false, "Show events")
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
	closeCh := make(chan struct{})
	var cb alien4cloud.ExecutionCallback = func(wfExec *alien4cloud.Execution, cbe error) {
		if wfExec != nil {
			log.Printf("Workflow ended with status: %s\n", wfExec.Status)
		}
		if cbe != nil {
			log.Printf("Workflow monitoring encountered an error: %v\n", cbe)
		}
		close(closeCh)
	}

	nbEvents := 0
	if showEvents {
		// get last number of events
		_, nbEvents, err = client.EventService().GetEventsForApplicationEnvironment(ctx, envID, 0, 1)
		if err != nil {
			log.Panic(err)
		}

	}
	execID, err := client.DeploymentService().RunWorkflowAsync(ctx, appName, envID, workflow, cb)
	if err != nil {
		log.Panic(err)
	}

	// Wait for the end of deployment
	log.Printf("Waiting for the end of workflow execution...")
	filters := alien4cloud.LogFilter{
		ExecutionID: []string{execID},
	}
	logIndex := 0
ExitLoop:
	for {
		select {
		case <-closeCh:
			break ExitLoop
		case <-time.After(5 * time.Second):
		}
		if showEvents {

			events, newNbEvents, err := client.EventService().GetEventsForApplicationEnvironment(ctx, envID, 0, nbEvents+100000)
			if err != nil {
				log.Panic(err)
			}
			// Results are sorted by date in descending order
			for idx := newNbEvents - nbEvents - 1; idx >= 0; idx-- {

				if events[idx].InstanceState != "" {
					// Printing a message like:
					// Event received: component Welcome instance 0 state stopping
					// Event received: component Welcome instance 0 state stopped
					log.Printf("Event received: component %s instance %s state %s",
						events[idx].NodeTemplateId, events[idx].InstanceId, events[idx].InstanceState)

				}
			}
			nbEvents = newNbEvents
			continue
		}

		// Else just display logs
		a4cLogs, nbLogs, err := client.LogService().GetLogsOfApplication(ctx, appName, envID, filters, logIndex)
		if err != nil {
			log.Panic(err)
		}
		if nbLogs > 0 {
			logIndex = logIndex + nbLogs
			for idx := 0; idx < nbLogs; idx++ {
				log.Printf("%s [%s][%s][%s][%s][%s][%s][%s] %s",
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
	}
}
