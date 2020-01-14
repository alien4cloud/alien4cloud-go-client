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
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// DeploymentService is the interface to the service managing deployments
type DeploymentService interface {
	// Gets matching locations where a given application can be deployed
	GetLocationsMatching(ctx context.Context, topologyID string, envID string) ([]LocationMatch, error)
	// Deploys the given application in the given environment using the given orchestrator
	// if location is empty, the first matching location will be used
	DeployApplication(ctx context.Context, appID string, envID string, location string) error
	// Updates an application with the latest topology version
	UpdateApplication(ctx context.Context, appID, envID string) error
	// Returns the deployment list for the given appID and envID
	GetDeploymentList(ctx context.Context, appID string, envID string) ([]Deployment, error)
	// Undeploys an application
	UndeployApplication(ctx context.Context, appID string, envID string) error
	// WaitUntilStateIs Waits until the state of an Alien4Cloud application is one of the given statuses as parameter and returns the actual status.
	WaitUntilStateIs(ctx context.Context, appID string, envID string, statuses ...string) (string, error)
	// Returns current deployment status for the given applicationID and environmentID
	GetDeploymentStatus(ctx context.Context, applicationID string, environmentID string) (string, error)
	// Returns current deployment ID for the given applicationID and environmentID
	GetCurrentDeploymentID(ctx context.Context, applicationID string, environmentID string) (string, error)
	// Returns the node status for the given applicationID and environmentID and nodeName
	GetNodeStatus(ctx context.Context, applicationID string, environmentID string, nodeName string) (string, error)
	// Returns the output attributes of nodes in the given applicationID and environmentID
	GetOutputAttributes(ctx context.Context, applicationID string, environmentID string) (map[string][]string, error)
	// Returns the application deployment attributes
	GetAttributesValue(ctx context.Context, applicationID string, environmentID string, nodeName string, requestedAttributesName []string) (map[string]string, error)
	// Runs Alien4Cloud workflowName workflow for the given a4cAppID and a4cEnvID
	RunWorkflow(ctx context.Context, a4cAppID string, a4cEnvID string, workflowName string, timeout time.Duration) (*WorkflowExecution, error)
	// Runs a workflow asynchronously, results will be notified using the ExecutionCallback function.
	// Cancelling the context cancels the function that monitor the execution
	RunWorkflowAsync(ctx context.Context, a4cAppID string, a4cEnvID string, workflowName string, callback ExecutionCallback) error
	// Returns the workflow execution for the given applicationID and environmentID
	GetLastWorkflowExecution(ctx context.Context, applicationID string, environmentID string) (*WorkflowExecution, error)

	// Returns executions for a given deployment
	// query allow to search a specific execution but may be empty
	// from and size allows to paginate results
	GetExecutions(ctx context.Context, deploymentID, query string, from, size int) ([]WorkflowExecution, FacetedSearchResult, error)
}

// ExecutionCallback is a function call by asynchronous operations when an execution reaches a terminal state
type ExecutionCallback func(*WorkflowExecution, error)

type deploymentService struct {
	client             restClient
	applicationService *applicationService
	topologyService    *topologyService
}

// Get matching locations where a given application can be deployed
func (d *deploymentService) GetLocationsMatching(ctx context.Context, topologyID string, envID string) ([]LocationMatch, error) {
	response, err := d.client.doWithContext(ctx,
		"GET",
		fmt.Sprintf("%s/topologies/%s/locations?environmentId=%s", a4CRestAPIPrefix, topologyID, envID),
		nil,
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get locations matching topology for application '%s' in '%s' environment",
			topologyID, envID)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, getError(response)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, errors.Wrapf(err, "Cannot read response to request on locations matching topology '%s' in '%s' environment",
			topologyID, envID)
	}
	var res struct {
		Data []LocationMatch `json:"data"`
	}

	if err = json.Unmarshal([]byte(responseBody), &res); err != nil {
		return nil, errors.Wrapf(err, "Cannot convert the body response to request on locations matching topology '%s' in '%s' environment",
			topologyID, envID)
	}

	return res.Data, err
}

// DeployApplication Deploy the given application in the given environment using the given orchestrator
// if location is empty, the first matching location will be used
func (d *deploymentService) DeployApplication(ctx context.Context, appID string, envID string, location string) error {

	// get locations matching this application
	topologyID, err := d.topologyService.GetTopologyID(ctx, appID, envID)
	if err != nil {
		return errors.Wrapf(err, "Unable to get application topology for app %s and env %s",
			appID, envID)
	}

	locationsMatch, err := d.GetLocationsMatching(ctx, topologyID, envID)
	if err != nil {
		return errors.Wrapf(err, "Failed to get locations matching app %s env %s",
			appID, envID)
	}

	locationID := ""
	orchestratorID := ""
	for _, locationMatch := range locationsMatch {
		if location == "" || locationMatch.Location.Name == location {
			locationID = locationMatch.Location.ID
			orchestratorID = locationMatch.Location.OrchestratorID
			break
		}
	}
	if locationID == "" {
		// Return the list of possible locations names
		var locationNames []string
		for _, locationMatch := range locationsMatch {
			locationNames = append(locationNames, locationMatch.Location.Name)
		}
		return errors.Errorf("Location %q not found in list of matching locations: %+v", location, locationNames)
	}
	// Set location policy for deployment
	var locationPolicies LocationPoliciesPostRequestIn
	locationPolicies.GroupsToLocations.A4CAll = locationID
	locationPolicies.OrchestratorID = orchestratorID

	body, err := json.Marshal(locationPolicies)
	if err != nil {
		return errors.Wrap(err, "Cannot marshal an a4cLocationPoliciesPostRequestIn structure")
	}
	response, err := d.client.doWithContext(ctx,
		"POST",
		fmt.Sprintf("%s/applications/%s/environments/%s/deployment-topology/location-policies", a4CRestAPIPrefix, appID, envID),
		[]byte(string(body)),
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send a request to set the location in order to deploy an application")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return getError(response)
	}

	// Deploy the application a4cApplicationDeployhRequestIn
	appDeployBody, err := json.Marshal(
		ApplicationDeployRequest{
			envID,
			appID,
		},
	)
	response, err = d.client.doWithContext(ctx,
		"POST",
		fmt.Sprintf("%s/applications/deployment", a4CRestAPIPrefix),
		[]byte(string(appDeployBody)),
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send a request to deploy the application")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return getError(response)
	}

	return nil
}

// UpdateApplication updates an application with the latest topology version
func (d *deploymentService) UpdateApplication(ctx context.Context, appID, envID string) error {

	response, err := d.client.doWithContext(ctx, "POST",
		fmt.Sprintf("%s/applications/%s/environments/%s/update-deployment", a4CRestAPIPrefix, appID, envID),
		[]byte("{}"),
		[]Header{contentTypeAppJSONHeader, acceptAppJSONHeader},
	)

	if err != nil {
		return errors.Wrapf(err, "Unable to send a request to update application %s", appID)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return getError(response)
	}

	return nil
}

// GetDeploymentList returns the deployment list for the given appID and envID
func (d *deploymentService) GetDeploymentList(ctx context.Context, appID string, envID string) ([]Deployment, error) {

	response, err := d.client.doWithContext(ctx,
		"GET",
		fmt.Sprintf("%s/deployments/search?environmentId=%s&from=0&query=", a4CRestAPIPrefix, envID),
		nil,
		[]Header{acceptAppJSONHeader},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to send request to get deployment list")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, getError(response)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, errors.Wrapf(err, "Cannot read the body when getting deployment list")
	}

	var deploymentListResponse struct {
		Data struct {
			Data []struct {
				Deployment Deployment
			}
			TotalResults int `json:"totalResults"`
		} `json:"data"`
	}

	err = json.Unmarshal(responseBody, &deploymentListResponse)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to unmarshal the deployment list")
	}

	var deploymentList []Deployment

	for _, dListData := range deploymentListResponse.Data.Data {
		deploymentList = append(deploymentList, dListData.Deployment)
	}

	return deploymentList, nil
}

// UndeployApplication Undeploy an application
func (d *deploymentService) UndeployApplication(ctx context.Context, appID string, envID string) error {

	response, err := d.client.doWithContext(ctx,
		"DELETE",
		fmt.Sprintf("%s/applications/%s/environments/%s/deployment", a4CRestAPIPrefix, appID, envID),
		nil,
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send request to undeploy A4C application")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return getError(response)
	}

	return nil
}

// WaitUntilStateIs Waits until the state of an Alien4Cloud application is one of the given statuses as parameter and returns the actual status.
func (d *deploymentService) WaitUntilStateIs(ctx context.Context, appID string, envID string, statuses ...string) (string, error) {
	if len(statuses) == 0 {
		return "", errors.New("at least one status should be given")
	}
	for {
		a4cStatus, err := d.GetDeploymentStatus(ctx, appID, envID)

		if err != nil {
			return "", errors.Wrapf(err, "Unable to get status from application %s", appID)
		}

		for _, status := range statuses {
			if a4cStatus == status {
				return a4cStatus, nil
			}
		}

		select {
		case <-ctx.Done():
			return "", errors.Wrapf(ctx.Err(), "Unable to get status from application %s", appID)
		case <-time.After(time.Second):
		}
	}
}

// GetDeploymentStatus returns current deployment status for the given applicationID and environmentID
func (d *deploymentService) GetDeploymentStatus(ctx context.Context, applicationID string, environmentID string) (string, error) {

	deploymentID, err := d.GetCurrentDeploymentID(ctx, applicationID, environmentID)
	if err != nil {
		return "", err
	}

	response, err := d.client.doWithContext(ctx,
		"GET",
		fmt.Sprintf("%s/deployments/%s/status", a4CRestAPIPrefix, deploymentID),
		nil,
		[]Header{acceptAppJSONHeader},
	)

	if err != nil {
		return "", errors.Wrap(err, "Cannot send a request to get the deployment status")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", getError(response)
	}

	var statusResponse struct {
		Data  string `json:"data"`
		Error *Error `json:"error,omitempty"`
	}

	err = readCloseResponseBody(response, &statusResponse)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to unmarshal the deployment status")
	}

	if statusResponse.Error != nil {
		return "", errors.New(statusResponse.Error.Message)
	}

	return statusResponse.Data, nil

}

// GetCurrentDeploymentID returns current deployment ID for the given applicationID and environmentID
func (d *deploymentService) GetCurrentDeploymentID(ctx context.Context, applicationID string, environmentID string) (string, error) {

	response, err := d.client.doWithContext(ctx,
		"GET",
		fmt.Sprintf("%s/applications/%s/environments/%s/active-deployment-monitored", a4CRestAPIPrefix, applicationID, environmentID),
		nil,
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return "", errors.Wrapf(err, "Unable to retrieve the current deployment ID for app '%s'", applicationID)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", getError(response)
	}

	var res struct {
		Data struct {
			Deployment struct {
				ID string `json:"id"`
			} `json:"deployment"`
		} `json:"data"`
	}

	err = readCloseResponseBody(response, &res)

	if err != nil {
		return "", errors.Wrap(err, "Unable to unmarshal content of the get deployment monitored request")
	}

	return res.Data.Deployment.ID, nil

}

// GetNodeStatus returns the node status for the given applicationID and environmentID and nodeName
func (d *deploymentService) GetNodeStatus(ctx context.Context, applicationID string, environmentID string, nodeName string) (string, error) {

	response, err := d.client.doWithContext(ctx,
		"GET",
		fmt.Sprintf("%s/applications/%s/environments/%s/deployment/informations", a4CRestAPIPrefix, applicationID, environmentID),
		nil,
		nil,
	)

	if err != nil {
		return "", errors.Wrapf(err, "Cannot send a request to get node status of node '%s'", nodeName)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", getError(response)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", errors.Wrapf(err, "Cannot read the body of the node status for node '%s'", nodeName)
	}

	var nodeStatusResponse Informations

	err = json.Unmarshal(responseBody, &nodeStatusResponse)

	if err != nil {
		return "", errors.Wrapf(err, "Unable to unmarshal node status for node '%s'", nodeName)
	}

	if len(nodeStatusResponse.Data) == 0 {
		return "", nil
	}

	for alienNodeName, node := range nodeStatusResponse.Data {
		if alienNodeName == nodeName {
			return node["0"].State, nil
		}
	}

	return "", fmt.Errorf("unable to get status of node '%s'", nodeName)

}

// GetOutputAttributes return the output attributes of nodes in the given applicationID and environmentID
func (d *deploymentService) GetOutputAttributes(ctx context.Context, applicationID string, environmentID string) (map[string][]string, error) {

	response, err := d.client.doWithContext(ctx,
		"GET",
		fmt.Sprintf("%s/runtime/%s/environment/%s/topology", a4CRestAPIPrefix, applicationID, environmentID),
		nil,
		nil,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Cannot send a request to get output properties")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, getError(response)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, errors.Wrap(err, "Cannot read the body of the output properties")
	}

	var outputPropertiesResponse RuntimeTopology

	err = json.Unmarshal(responseBody, &outputPropertiesResponse)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal output properties")
	}

	return outputPropertiesResponse.Data.Topology.OutputAttributes, nil

}

// GetAttributesValue returns the application deployment attributes
func (d *deploymentService) GetAttributesValue(ctx context.Context, applicationID string, environmentID string, nodeName string, requestedAttributesName []string) (map[string]string, error) {

	response, err := d.client.doWithContext(ctx,
		"GET",
		fmt.Sprintf("%s/applications/%s/environments/%s/deployment/informations", a4CRestAPIPrefix, applicationID, environmentID),
		nil,
		nil,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Cannot send a request to get attributes value")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, getError(response)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, errors.Wrapf(err, "Cannot read the body of the attributes value response '%s' in '%s' environment", applicationID, environmentID)
	}

	var nodeStatusResponse Informations

	err = json.Unmarshal(responseBody, &nodeStatusResponse)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal attributes value")
	}

	if len(nodeStatusResponse.Data) == 0 {
		return nil, nil
	}

	attributesValue := map[string]string{}

	// Iterate over the data returned by A4C in order to get values of requested attributes (they can have multiple).
	// This script just take the attribute value of the first instance of the given node.

	for alienNodeName, node := range nodeStatusResponse.Data {
		if alienNodeName == nodeName {
			for _, attributeName := range requestedAttributesName {
				for alienAttributeName, attributeValue := range node["0"].Attributes {
					if attributeName == alienAttributeName {
						attributesValue[attributeName] = attributeValue
						// Just to improve performances
						delete(node["0"].Attributes, alienAttributeName)
						break
					}
				}
			}

			break
		}
	}

	return attributesValue, nil
}

// Runs a workflow asynchronously, results will be notified using the ExecutionCallback function.
// Cancelling the context cancels the function that monitor the execution
func (d *deploymentService) RunWorkflowAsync(ctx context.Context, a4cAppID string, a4cEnvID string, workflowName string, callback ExecutionCallback) error {
	response, err := d.client.doWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/applications/%s/environments/%s/workflows/%s", a4CRestAPIPrefix, a4cAppID, a4cEnvID, workflowName),
		nil,
		[]Header{acceptAppJSONHeader},
	)
	if err != nil {
		return errors.Wrapf(err, "failed to run workflow %q on application %q, environment %q", workflowName, a4cAppID, a4cEnvID)
	}
	var res struct {
		Data string `json:"data"`
	}

	err = readCloseResponseBody(response, &res)
	if err != nil {
		return errors.Wrapf(err, "failed to read response on running workflow %q on application %q, environment %q", workflowName, a4cAppID, a4cEnvID)
	}

	if res.Data == "" {
		return errors.Errorf("no execution id returned on run workflow %q on application %q, environment %q", workflowName, a4cAppID, a4cEnvID)
	}

	deploymentID, err := d.GetCurrentDeploymentID(ctx, a4cAppID, a4cEnvID)
	if err != nil {
		return errors.Wrapf(err, "failed to get deployment id on running workflow %q on application %q, environment %q", workflowName, a4cAppID, a4cEnvID)
	}

	// now monitor workflow execution
	go func() {
		for {
			executions, _, err := d.GetExecutions(ctx, deploymentID, res.Data, 0, 1)
			if err != nil {
				callback(nil, err)
				return
			}
			if len(executions) != 1 {
				callback(nil,
					errors.Errorf("expecting 1 execution on monitoring execution id %q for workflow %q on application %q, environment %q, but actually got %d executions", res.Data, workflowName, a4cAppID, a4cEnvID, len(executions)))
				return
			}

			switch executions[0].Status {
			case "SUCCEEDED", "CANCELLED", "FAILED":
				callback(&executions[0], nil)
				return
			default:
			}

			select {
			case <-ctx.Done():
				callback(nil, ctx.Err())
				return
			case <-time.After(5 * time.Second):
			}
		}
	}()

	return nil
}

// RunWorkflow runs a4c workflowName workflow for the given a4cAppID and a4cEnvID
func (d *deploymentService) RunWorkflow(ctx context.Context, a4cAppID string, a4cEnvID string, workflowName string, timeout time.Duration) (*WorkflowExecution, error) {
	ctx, cancelFunc := context.WithTimeout(ctx, timeout)
	defer cancelFunc()

	var wfExec *WorkflowExecution
	doneCh := make(chan struct{})
	var cbErr error
	err := d.RunWorkflowAsync(ctx, a4cAppID, a4cEnvID, workflowName, func(exec *WorkflowExecution, e error) {
		wfExec = exec
		cbErr = e
		close(doneCh)
	})
	if err != nil {
		return nil, err
	}

	<-doneCh
	return wfExec, cbErr
}

// GetLastWorkflowExecution return a4c workflow execution for the given applicationID and environmentID
func (d *deploymentService) GetLastWorkflowExecution(ctx context.Context, applicationID string, environmentID string) (*WorkflowExecution, error) {

	deploymentID, err := d.GetCurrentDeploymentID(ctx, applicationID, environmentID)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to get current deployment ID")
	}

	response, err := d.client.doWithContext(ctx,
		"GET",
		fmt.Sprintf("%s/workflow_execution/%s", a4CRestAPIPrefix, deploymentID),
		nil,
		[]Header{acceptAppJSONHeader},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get workflow status of application '%s'", applicationID)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, getError(response)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, errors.Wrap(err, "Cannot read the response from Alien4Cloud")
	}

	var res struct {
		Data struct {
			Execution WorkflowExecution `json:"execution"`
		} `json:"data"`
	}

	err = json.Unmarshal(responseBody, &res)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal content of the execution status response")
	}

	return &res.Data.Execution, nil

}
