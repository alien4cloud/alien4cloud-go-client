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

	"github.com/alien4cloud/alien4cloud-go-client/v2/alien4cloud"
	"github.com/fatih/color"
	"github.com/pkg/errors"
)

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

	// Get workflow description
	topo, err := client.ApplicationService().GetDeploymentTopology(ctx, appName, envID)
	if err != nil {
		log.Panic(err)
	}

	wf, found := topo.Data.Topology.Workflows[workflow]
	if !found {
		log.Panicf("Found no workflow %s in application %s", appName, workflow)
	}
	// Identify initial steps
	var initialSteps []alien4cloud.WorkflowStep
	for _, wfStep := range wf.Steps {
		if len(wfStep.PrecedingSteps) == 0 {
			initialSteps = append(initialSteps, wfStep)
		}
	}

	// Get workflow steps status
	wfExec, err := client.DeploymentService().GetLastWorkflowExecution(ctx, appName, envID)
	if err != nil {
		log.Panic(err)
	}

	// Print workflow steps status
	for _, step := range initialSteps {
		description, err := getStepDescription(&step)
		if err != nil {
			log.Panic(err)
		}
		status, _ := getStepStatus(&step, wfExec)
		printStep(description, status)
		if len(step.OnSuccess) > 0 {
			err := printNextSteps(step.OnSuccess, wf, wfExec, len(description))
			if err != nil {
				log.Panic(err)
			}
		} else {
			fmt.Println("")
		}

	}
}

func getStep(stepName string, wf alien4cloud.Workflow) (*alien4cloud.WorkflowStep, error) {
	for _, wfStep := range wf.Steps {
		if wfStep.Name == stepName {
			return &wfStep, nil
		}
	}

	return nil, errors.Errorf("No such step %s in workflow %s", stepName, wf.Name)
}

func getStepDescription(step *alien4cloud.WorkflowStep) (string, error) {
	var description string
	var err error
	if len(step.Activities) != 1 {
		return description, errors.Errorf("This example expects 1 workflow step activity, got %d for step %s",
			len(step.Activities), step.Name)
	}
	activity := step.Activities[0]
	switch activity.Type {
	case alien4cloud.CallOperationWorkflowActivityType:
		description = fmt.Sprintf("%s %s", step.Target, activity.OperationName)
	case alien4cloud.DelegateWorkflowActivity:
		description = fmt.Sprintf("%s %s", step.Target, activity.Delegate)
	case alien4cloud.SetStateWorkflowActivityType:
		description = fmt.Sprintf("%s %s", step.Target, activity.StateName)
	case alien4cloud.InlineWorkflowActivityType:
		description = fmt.Sprintf("Workflow %s", activity.Inline)
	default:
		err = errors.Errorf("Unexpected activity type %s for step %s", activity.Type, step.Name)
	}
	return description, err
}

func getStepStatus(step *alien4cloud.WorkflowStep, wfExec *alien4cloud.WorkflowExecution) (string, bool) {
	status, found := wfExec.StepStatus[step.Name]
	return status, found
}

func printStep(description, status string) {

	switch status {
	case alien4cloud.StepCompletedSuccessfull:
		color.New(color.FgGreen).Printf("%s", description)
	case alien4cloud.StepCompletedWithError:
		color.New(color.FgRed).Printf("%s", description)
	case alien4cloud.StepStarted:
		color.New(color.FgBlue).Printf("%s", description)
	default:
		fmt.Printf("%s", description)
	}
}

func printNextSteps(stepNames []string, wf alien4cloud.Workflow, wfExec *alien4cloud.WorkflowExecution, indent int) error {

	for i, stepName := range stepNames {
		if i != 0 {
			fmt.Printf("%s", strings.Repeat(" ", indent))
		}
		step, err := getStep(stepName, wf)
		if err != nil {
			return err
		}

		description, err := getStepDescription(step)
		if err != nil {
			return err
		}
		status, _ := getStepStatus(step, wfExec)
		if err != nil {
			log.Panic(err)
		}

		var newIdentation int
		if status == "" && step.Activities[0].Type == alien4cloud.SetStateWorkflowActivityType {
			// Skipping steps setting a state
			// The status of such steps is not available in the workflow execution
			newIdentation = 0
		} else {
			fmt.Printf(" -> ")
			printStep(description, status)
			newIdentation = len(description) + 4
		}

		if len(step.OnSuccess) > 0 {
			err := printNextSteps(step.OnSuccess, wf, wfExec, indent+newIdentation)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("")
		}
	}
	return nil
}
