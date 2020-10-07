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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
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
	// Updates inputs of a deployment topology
	UpdateDeploymentTopology(ctx context.Context, appID, envID string, request UpdateDeploymentTopologyRequest) error
	// Uploads an input artifact
	UploadDeploymentInputArtifact(ctx context.Context, appID, envID, inputArtifact, filePath string) error
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
	// Returns the application deployment attributes for the first instance of a node name
	GetAttributesValue(ctx context.Context, applicationID string, environmentID string, nodeName string, requestedAttributesName []string) (map[string]string, error)
	// Returns the application deployment attributes for the specified instance of a node name
	GetInstanceAttributesValue(ctx context.Context, applicationID string, environmentID string, nodeName, instanceName string, requestedAttributesName []string) (map[string]string, error)
	// Runs Alien4Cloud workflowName workflow for the given a4cAppID and a4cEnvID
	RunWorkflow(ctx context.Context, a4cAppID string, a4cEnvID string, workflowName string, timeout time.Duration) (*Execution, error)
	// Runs a workflow asynchronously returning the execution id, results will be notified using the ExecutionCallback function.
	// Cancelling the context cancels the function that monitor the execution
	RunWorkflowAsync(ctx context.Context, a4cAppID string, a4cEnvID string, workflowName string, callback ExecutionCallback) (string, error)
	// Returns the workflow execution for the given applicationID and environmentID
	GetLastWorkflowExecution(ctx context.Context, applicationID string, environmentID string) (*WorkflowExecution, error)

	// Returns executions
	//
	// - deploymentID allows to search executions of a specific deployment but may be empty
	// - query allows to search a specific execution but may be empty
	// - from and size allows to paginate results
	GetExecutions(ctx context.Context, deploymentID, query string, from, size int) ([]Execution, FacetedSearchResult, error)
	// Cancels execution for given environmentID and executionID
	CancelExecution(ctx context.Context, environmentID string, executionID string) error
}

// ExecutionCallback is a function call by asynchronous operations when an execution reaches a terminal state
type ExecutionCallback func(*Execution, error)

type deploymentService struct {
	client *a4cClient
}

// Get matching locations where a given application can be deployed
func (d *deploymentService) GetLocationsMatching(ctx context.Context, topologyID string, envID string) ([]LocationMatch, error) {
	request, err := d.client.NewRequest(ctx,
		"GET",
		fmt.Sprintf("%s/topologies/%s/locations?environmentId=%s", a4CRestAPIPrefix, topologyID, envID),
		nil,
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get locations matching topology for application '%s' in '%s' environment",
			topologyID, envID)
	}
	var res struct {
		Data []LocationMatch `json:"data"`
	}
	response, err := d.client.Do(request)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get locations matching topology for application '%s' in '%s' environment",
			topologyID, envID)
	}
	err = ReadA4CResponse(response, &res)
	return res.Data, errors.Wrapf(err, "Cannot convert the body response to request on locations matching topology '%s' in '%s' environment",
		topologyID, envID)
}

// DeployApplication Deploy the given application in the given environment using the given orchestrator
// if location is empty, the first matching location will be used
func (d *deploymentService) DeployApplication(ctx context.Context, appID string, envID string, location string) error {

	// get locations matching this application
	topologyID, err := d.client.topologyService.GetTopologyID(ctx, appID, envID)
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
	request, err := d.client.NewRequest(ctx,
		"POST",
		fmt.Sprintf("%s/applications/%s/environments/%s/deployment-topology/location-policies", a4CRestAPIPrefix, appID, envID),
		bytes.NewReader(body),
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send a request to set the location in order to deploy an application")
	}
	response, err := d.client.Do(request)
	if err != nil {
		return errors.Wrap(err, "Unable to send a request to set the location in order to deploy an application")
	}

	err = ReadA4CResponse(response, nil)
	if err != nil {
		return errors.Wrap(err, "Unable to set the location in order to deploy an application")
	}

	// Deploy the application a4cApplicationDeployhRequestIn
	appDeployBody, err := json.Marshal(
		ApplicationDeployRequest{
			envID,
			appID,
		},
	)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal application deployment request")
	}

	request, err = d.client.NewRequest(ctx,
		"POST",
		fmt.Sprintf("%s/applications/deployment", a4CRestAPIPrefix),
		bytes.NewReader(appDeployBody),
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send a request to deploy the application")
	}
	response, err = d.client.Do(request)
	if err != nil {
		return errors.Wrap(err, "Unable to send a request to deploy the application")
	}
	err = ReadA4CResponse(response, nil)
	return errors.Wrap(err, "Unable to deploy the application")
}

// UpdateApplication updates an application with the latest topology version
func (d *deploymentService) UpdateApplication(ctx context.Context, appID, envID string) error {

	request, err := d.client.NewRequest(ctx, "POST",
		fmt.Sprintf("%s/applications/%s/environments/%s/update-deployment", a4CRestAPIPrefix, appID, envID),
		bytes.NewReader([]byte("{}")),
	)

	if err != nil {
		return errors.Wrapf(err, "Unable to send a request to update application %s", appID)
	}

	response, err := d.client.Do(request)
	if err != nil {
		return errors.Wrapf(err, "Unable to send a request to update application %s", appID)
	}
	err = ReadA4CResponse(response, err)
	return errors.Wrapf(err, "Unable to update application %s", appID)
}

// UpdateDeploymentTopology updates inputs of a deployment topology
func (d *deploymentService) UpdateDeploymentTopology(ctx context.Context, appID, envID string,
	upDepTopoRequest UpdateDeploymentTopologyRequest) error {

	requestBody, _ := json.Marshal(upDepTopoRequest)
	request, err := d.client.NewRequest(ctx, "PUT",
		fmt.Sprintf("%s/applications/%s/environments/%s/deployment-topology", a4CRestAPIPrefix, appID, envID),
		bytes.NewReader(requestBody),
	)

	if err != nil {
		return errors.Wrapf(err, "Unable to send a request to deployment topology for application %s", appID)
	}

	response, err := d.client.Do(request)
	if err != nil {
		return errors.Wrapf(err, "Unable to send a request to update deployment topology for application %s", appID)
	}
	err = ReadA4CResponse(response, nil)
	return errors.Wrapf(err, "Unable to update deployment topology for application %s", appID)
}

// Uploads an input artifact

func (d *deploymentService) UploadDeploymentInputArtifact(ctx context.Context,
	appID, envID, inputArtifact, filePath string) error {

	f, err := os.Open(filePath)
	if err != nil {
		return errors.Wrapf(err, "Failed to open file to upload %s", filePath)
	}
	defer f.Close()

	// TODO(loicalbertin) we may have an issue on large files as it will load the whole file in memory.
	// We should consider using io.Pipe() to create a synchronous in-memory pipe.
	// The tricky part will be to make it work with an expected io.ReadSeeker.
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	fName := filepath.Base(filePath)
	part, err := writer.CreateFormFile("file", fName)
	if err != nil {
		return errors.Wrapf(err, "Failed to create from file for %s", fName)
	}
	_, err = io.Copy(part, f)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	request, err := d.client.NewRequest(ctx, "POST",
		fmt.Sprintf("%s/applications/%s/environments/%s/deployment-topology/inputArtifacts/%s/upload",
			a4CRestAPIPrefix, appID, envID, inputArtifact),
		bytes.NewReader(body.Bytes()),
	)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	if err != nil {
		return errors.Wrapf(err, "Unable to send a request to deployment topology for application %s", appID)
	}

	response, err := d.client.Do(request)
	if err != nil {
		return errors.Wrapf(err, "Unable to send a request to deployment topology for application %s", appID)
	}
	err = ReadA4CResponse(response, nil)
	return errors.Wrapf(err, "Unable to deployment topology for application %s", appID)
}

// GetDeploymentList returns the deployment list for the given appID and envID
func (d *deploymentService) GetDeploymentList(ctx context.Context, appID string, envID string) ([]Deployment, error) {

	request, err := d.client.NewRequest(ctx,
		"GET",
		fmt.Sprintf("%s/deployments/search?environmentId=%s&from=0&query=", a4CRestAPIPrefix, envID),
		nil,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to send request to get deployment list")
	}

	var deploymentListResponse struct {
		Data struct {
			Data []struct {
				Deployment Deployment
			}
			TotalResults int `json:"totalResults"`
		} `json:"data"`
	}
	response, err := d.client.Do(request)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get deployment list response for application %q environment %q", appID, envID)
	}

	err = ReadA4CResponse(response, &deploymentListResponse)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get deployment list for application %q environment %q", appID, envID)
	}
	var deploymentList []Deployment

	for _, dListData := range deploymentListResponse.Data.Data {
		deploymentList = append(deploymentList, dListData.Deployment)
	}

	return deploymentList, nil
}

// UndeployApplication Undeploy an application
func (d *deploymentService) UndeployApplication(ctx context.Context, appID string, envID string) error {

	request, err := d.client.NewRequest(ctx,
		"DELETE",
		fmt.Sprintf("%s/applications/%s/environments/%s/deployment", a4CRestAPIPrefix, appID, envID),
		nil,
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send request to undeploy A4C application")
	}
	response, err := d.client.Do(request)
	if err != nil {
		return errors.Wrap(err, "Unable to send request to undeploy A4C application")
	}
	err = ReadA4CResponse(response, nil)
	return errors.Wrap(err, "Unable to undeploy A4C application")
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

	if deploymentID == "" {
		// Application is not deployed
		return ApplicationUndeployed, err
	}

	request, err := d.client.NewRequest(ctx,
		"GET",
		fmt.Sprintf("%s/deployments/%s/status", a4CRestAPIPrefix, deploymentID),
		nil,
	)

	if err != nil {
		return "", errors.Wrap(err, "Cannot send a request to get the deployment status")
	}

	var statusResponse struct {
		Data string `json:"data"`
	}

	response, err := d.client.Do(request)

	if err != nil {
		return "", errors.Wrap(err, "Cannot send a request to get the deployment status")
	}

	err = ReadA4CResponse(response, &statusResponse)
	return statusResponse.Data, errors.Wrapf(err, "Unable to get deployment status for application %q environment %q", applicationID, environmentID)

}

// GetCurrentDeploymentID returns current deployment ID for the given applicationID and environmentID
// Returns an empty string if the application is undeployed
func (d *deploymentService) GetCurrentDeploymentID(ctx context.Context, applicationID string, environmentID string) (string, error) {

	request, err := d.client.NewRequest(ctx,
		"GET",
		fmt.Sprintf("%s/applications/%s/environments/%s/active-deployment-monitored", a4CRestAPIPrefix, applicationID, environmentID),
		nil,
	)

	if err != nil {
		return "", errors.Wrapf(err, "Unable to retrieve the current deployment ID for app '%s'", applicationID)
	}

	var res struct {
		Data struct {
			Deployment struct {
				ID string `json:"id"`
			} `json:"deployment"`
		} `json:"data"`
	}

	response, err := d.client.Do(request)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to retrieve the current deployment ID for app '%s'", applicationID)
	}
	err = ReadA4CResponse(response, &res)
	return res.Data.Deployment.ID, errors.Wrap(err, "Unable to unmarshal content of the get deployment monitored request")

}

// GetNodeStatus returns the node status for the given applicationID and environmentID and nodeName
func (d *deploymentService) GetNodeStatus(ctx context.Context, applicationID string, environmentID string, nodeName string) (string, error) {

	request, err := d.client.NewRequest(ctx,
		"GET",
		fmt.Sprintf("%s/applications/%s/environments/%s/deployment/informations", a4CRestAPIPrefix, applicationID, environmentID),
		nil,
	)

	if err != nil {
		return "", errors.Wrapf(err, "Cannot send a request to get node status of node '%s'", nodeName)
	}

	var nodeStatusResponse Informations
	response, err := d.client.Do(request)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to unmarshal node status for node '%s'", nodeName)
	}

	err = ReadA4CResponse(response, &nodeStatusResponse)
	if err != nil {
		return "", errors.Wrapf(err, "unable to get status of node '%s'", nodeName)
	}

	if len(nodeStatusResponse.Data) == 0 {
		return "", nil
	}

	for alienNodeName, node := range nodeStatusResponse.Data {
		if alienNodeName == nodeName {
			return node["0"].State, nil
		}
	}

	return "", errors.Errorf("unable to get status of node '%s'", nodeName)

}

// GetOutputAttributes return the output attributes of nodes in the given applicationID and environmentID
func (d *deploymentService) GetOutputAttributes(ctx context.Context, applicationID string, environmentID string) (map[string][]string, error) {

	request, err := d.client.NewRequest(ctx,
		"GET",
		fmt.Sprintf("%s/runtime/%s/environment/%s/topology", a4CRestAPIPrefix, applicationID, environmentID),
		nil,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Cannot send a request to get output properties")
	}
	var outputPropertiesResponse RuntimeTopology
	response, err := d.client.Do(request)

	if err != nil {
		return nil, errors.Wrap(err, "Cannot send a request to get output properties")
	}
	err = ReadA4CResponse(response, &outputPropertiesResponse)
	return outputPropertiesResponse.Data.Topology.OutputAttributes, errors.Wrap(err, "Unable to get output properties")

}

// GetAttributesValue returns the application deployment attributes for the first instance of the specified nodeName
func (d *deploymentService) GetAttributesValue(ctx context.Context, applicationID string, environmentID string, nodeName string, requestedAttributesName []string) (map[string]string, error) {
	return d.getInstanceAttributesValue(ctx, applicationID, environmentID, nodeName, "0", requestedAttributesName)
}

// GetInstanceAttributesValue returns the application deployment attributes for a specified nodeName and instanceName
func (d *deploymentService) GetInstanceAttributesValue(ctx context.Context, applicationID string, environmentID string, nodeName, instanceName string, requestedAttributesName []string) (map[string]string, error) {
	return d.getInstanceAttributesValue(ctx, applicationID, environmentID, nodeName, instanceName, requestedAttributesName)
}

func (d *deploymentService) getInstanceAttributesValue(ctx context.Context, applicationID string, environmentID string, nodeName, instanceName string, requestedAttributesName []string) (map[string]string, error) {

	request, err := d.client.NewRequest(ctx,
		"GET",
		fmt.Sprintf("%s/applications/%s/environments/%s/deployment/informations", a4CRestAPIPrefix, applicationID, environmentID),
		nil,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Cannot send a request to get attributes value")
	}
	var nodeStatusResponse Informations
	response, err := d.client.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get attributes value")
	}
	err = ReadA4CResponse(response, &nodeStatusResponse)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get attributes value")
	}
	if len(nodeStatusResponse.Data) == 0 {
		return nil, nil
	}

	attributesValue := map[string]string{}

	// Iterate over the data returned by A4C in order to get values of requested attributes (they can have multiple).

	for alienNodeName, node := range nodeStatusResponse.Data {
		if alienNodeName == nodeName {
			for _, attributeName := range requestedAttributesName {
				for alienAttributeName, attributeValue := range node[instanceName].Attributes {
					if attributeName == alienAttributeName {
						attributesValue[attributeName] = attributeValue
						// Just to improve performances
						delete(node[instanceName].Attributes, alienAttributeName)
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
func (d *deploymentService) RunWorkflowAsync(ctx context.Context, a4cAppID string, a4cEnvID string, workflowName string, callback ExecutionCallback) (string, error) {
	request, err := d.client.NewRequest(
		ctx,
		"POST",
		fmt.Sprintf("%s/applications/%s/environments/%s/workflows/%s", a4CRestAPIPrefix, a4cAppID, a4cEnvID, workflowName),
		nil,
	)
	if err != nil {
		return "", errors.Wrapf(err, "failed to run workflow %q on application %q, environment %q", workflowName, a4cAppID, a4cEnvID)
	}
	var res struct {
		Data string `json:"data"`
	}
	response, err := d.client.Do(request)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read response on running workflow %q on application %q, environment %q", workflowName, a4cAppID, a4cEnvID)
	}
	err = ReadA4CResponse(response, &res)
	if err != nil {
		return "", errors.Wrapf(err, "failed to run workflow %q on application %q, environment %q", workflowName, a4cAppID, a4cEnvID)
	}

	if res.Data == "" {
		return "", errors.Errorf("no execution id returned on run workflow %q on application %q, environment %q", workflowName, a4cAppID, a4cEnvID)
	}
	// Let a4c time to register execution (500ms is not enough)
	<-time.After(time.Second)
	// now monitor workflow execution
	go func() {
		for {
			executions, _, err := d.GetExecutions(ctx, "", res.Data, 0, 1)
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

	return res.Data, nil
}

// RunWorkflow runs a4c workflowName workflow for the given a4cAppID and a4cEnvID
func (d *deploymentService) RunWorkflow(ctx context.Context, a4cAppID string, a4cEnvID string, workflowName string, timeout time.Duration) (*Execution, error) {
	ctx, cancelFunc := context.WithTimeout(ctx, timeout)
	defer cancelFunc()

	var execParam *Execution
	doneCh := make(chan struct{})
	var cbErr error
	_, err := d.RunWorkflowAsync(ctx, a4cAppID, a4cEnvID, workflowName, func(exec *Execution, e error) {
		execParam = exec
		cbErr = e
		close(doneCh)
	})
	if err != nil {
		return nil, err
	}

	<-doneCh
	return execParam, cbErr
}

// GetLastWorkflowExecution return a4c workflow execution for the given applicationID and environmentID
func (d *deploymentService) GetLastWorkflowExecution(ctx context.Context, applicationID string, environmentID string) (*WorkflowExecution, error) {

	deploymentID, err := d.GetCurrentDeploymentID(ctx, applicationID, environmentID)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to get current deployment ID")
	}

	request, err := d.client.NewRequest(ctx,
		"GET",
		fmt.Sprintf("%s/workflow_execution/%s", a4CRestAPIPrefix, deploymentID),
		nil,
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get workflow status of application '%s'", applicationID)
	}

	var res struct {
		Data WorkflowExecution `json:"data"`
	}

	response, err := d.client.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get content of the execution status response")
	}
	err = ReadA4CResponse(response, &res)
	return &res.Data, errors.Wrap(err, "Unable to get content of the execution status response")

}
