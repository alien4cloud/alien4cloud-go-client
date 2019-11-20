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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/goware/urlx"
	"github.com/pkg/errors"
)

// Client is the client interface to Alien4cloud service
type Client interface {
	Login() error
	Logout() error
	// Create an application from a template and return its ID
	CreateAppli(appName string, appTemplate string) (string, error)
	// Return the Alien4Cloud environment ID from a given application ID and environment name
	GetEnvironmentIDbyName(appID string, envName string) (string, error)
	// Return true if the application with the given ID exists
	IsApplicationExist(applicationID string) (bool, error)
	GetApplicationsID(filter string) ([]string, error)
	GetApplicationByID(id string) (*Application, error)
	// delete an application
	DeleteApplication(appID string) error
	SetTagToApplication(applicationID string, tagKey string, tagValue string) error
	GetApplicationTag(applicationID string, tagKey string) (string, error)
	// Get matching locations where a given application can be deployed
	GetLocationsMatching(topologyID string, envID string) ([]LocationMatch, error)
	// Deploy the given application in the given environment using the given orchestrator
	// if location is empty, the first matching location will be used
	DeployApplication(appID string, envID string, location string) error
	GetDeploymentList(appID string, envID string) ([]Deployment, error)
	// Undeploy an application
	UndeployApplication(appID string, envID string) error
	// Wait until the state of an Alien4Cloud application is the one given as parameter.
	WaintUntilStateIs(appID string, envID string, status string) error
	GetDeploymentStatus(applicationID string, environmentID string) (string, error)
	GetCurrentDeploymentID(applicationID string, environmentID string) (string, error)
	DisplayNodeStatus(applicationID string, environmentID string, nodeName string)
	GetNodeStatus(applicationID string, environmentID string, nodeName string) (string, error)
	GetOutputAttributes(applicationID string, environmentID string) (map[string][]string, error)
	GetAttributesValue(applicationID string, environmentID string, nodeName string, requestedAttributesName []string) (map[string]string, error)
	// Update the property value (type string) of a component of an application
	UpdateComponentProperty(a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue string) error
	// Update the property value (type tosca complex) of a component of an application
	UpdateComponentPropertyComplexType(a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue map[string]interface{}) error
	// Update the property value of a capability related to a component of an application
	UpdateCapabilityProperty(a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue string, capabilityName string) error
	// Add a new node in the A4C topology
	AddNodeInA4CTopology(a4cCtx *TopologyEditorContext, nodeTypeID string, nodeName string) error
	// Add a new relationship in the A4C topology
	AddRelationship(a4cCtx *TopologyEditorContext, sourceNodeName string, targetNodeName string, relType string) error
	SaveA4CTopology(a4cCtx *TopologyEditorContext) error
	GetOrchestratorLocations(orchestratorID string) ([]Location, error)
	// Return the Alien4Cloud orchestrator ID from a given orchestator name
	GetOrchestratorIDbyName(orchestratorName string) (string, error)
	GetLogsOfApplication(applicationID string, environmentID string, filters LogFilter, fromIndex int) ([]Log, int, error)
	RunWorkflow(a4cAppID string, a4cEnvID string, workflowName string) (*WorkflowExecution, error)
	GetLastWorkflowExecution(applicationID string, environmentID string) (*WorkflowExecution, error)
}

const (
	// DefaultEnvironmentName is the default name of the environment created by
	// Alien4Cloud for an application
	DefaultEnvironmentName = "Environment"
	// ApplicationDeploymentInProgress a4c status
	ApplicationDeploymentInProgress = "deployment_in_progress"
	// ApplicationDeployed a4c status
	ApplicationDeployed = "deployed"
	// ApplicationUndeploymentInProgress a4c status
	ApplicationUndeploymentInProgress = "undeployment_in_progress"
	// ApplicationUndeployed a4c status
	ApplicationUndeployed = "undeployed"
	// ApplicationError a4c status
	ApplicationError = "failure"

	// WorkflowSucceeded workflow a4c status
	WorkflowSucceeded = "SUCCEEDED"
	// WorkflowRunning workflow a4c status
	WorkflowRunning = "RUNNING"
	// WorkflowFailed workflow a4c status
	WorkflowFailed = "FAILED"

	// NodeStart node a4c status
	NodeStart = "initial"
	// NodeSubmitting node a4c status
	NodeSubmitting = "submitting"
	// NodeSubmitted node  a4c status
	NodeSubmitted = "submitted"
	// NodePending node  a4c status
	NodePending = "pending"
	// NodeRunning node  a4c status
	NodeRunning = "running"
	// NodeExecuting node  a4c status
	NodeExecuting = "executing"
	// NodeExecuted node  a4c status
	NodeExecuted = "executed"
	// NodeEnd node  a4c status
	NodeEnd = "end"
	// NodeError node  a4c status
	NodeError = "error"
	// NodeFailed node  a4c status
	NodeFailed = "failed"
	// NodeStart node  a4c status
)

const (
	// a4CRestAPIPrefix a4c rest api prefix
	a4CRestAPIPrefix string = "/rest/latest"

	// a4cUpdateNodePropertyValueOperationJavaClassName a4c class name to update node property value operation
	a4cUpdateNodePropertyValueOperationJavaClassName = "org.alien4cloud.tosca.editor.operations.nodetemplate.UpdateNodePropertyValueOperation"

	// a4cUpdateNodePropertyValueSlurmJobOptions yorc struct name for slurm JobOptions
	a4cUpdateNodePropertyValueSlurmJobOptions = "yorc.datatypes.slurm.JobOptions"

	// a4cUpdateCapabilityPropertyValueOperationJavaClassName a4c class name to update capability value operation
	a4cUpdateCapabilityPropertyValueOperationJavaClassName = "org.alien4cloud.tosca.editor.operations.nodetemplate.UpdateCapabilityPropertyValueOperation"

	// a4cAddNodeOperationJavaClassName a4c class name to add node operation
	a4cAddNodeOperationJavaClassName = "org.alien4cloud.tosca.editor.operations.nodetemplate.AddNodeOperation"

	// a4cAddRelationshipOperationJavaClassName a4c class name to add relationship operation
	a4cAddRelationshipOperationJavaClassName = "org.alien4cloud.tosca.editor.operations.relationshiptemplate.AddRelationshipOperation"
)

// a4Client holds properties of an a4c client
type a4cClient struct {
	*http.Client
	baseURL        string
	username       string
	password       string
	checkWfTimeout int
}

// NewClient instanciates and returns Client
func NewClient(address string, user string, password string, wfTimeout int, caFile string, skipSecure bool) (Client, error) {
	a4cAPI := strings.TrimRight(address, "/")

	if m, _ := regexp.Match("^http[s]?://.*", []byte(a4cAPI)); !m {
		a4cAPI = "http://" + a4cAPI
	}

	var useTLS = true
	if m, _ := regexp.Match("^http://.*", []byte(a4cAPI)); m {
		useTLS = false
	}

	url, err := urlx.Parse(a4cAPI)
	if err != nil {
		return nil, errors.Wrapf(err, "Malformed alien4cloud URL: %s", a4cAPI)
	}

	a4chost, _, err := urlx.SplitHostPort(url)
	if err != nil {
		return nil, errors.Wrapf(err, "Malformed alien4cloud URL %s", url)
	}

	tlsConfig := &tls.Config{ServerName: a4chost}

	if useTLS {
		if caFile == "" || skipSecure {
			if skipSecure {
				log.Printf("WARNING: Skipping TLS verification to connect to a4c API. It should not been used in production")
				tlsConfig.InsecureSkipVerify = true
			} else {
				return nil, errors.Errorf("You must provide a certificate authority file in TLS verify mode")
			}
		}

		if !skipSecure {
			certPool := x509.NewCertPool()
			caCert, err := ioutil.ReadFile(caFile)
			if err != nil {
				return nil, errors.Wrapf(err, "Failed to read certificate authority file")
			}
			if !certPool.AppendCertsFromPEM(caCert) {
				return nil, errors.Errorf("%q is not a valid certificate authority.", caCert)
			}
			tlsConfig.RootCAs = certPool
		}
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     tlsConfig,
	}

	return &a4cClient{
		Client: &http.Client{
			Transport:     tr,
			CheckRedirect: nil,
			Jar:           NewJar(),
			Timeout:       0},
		baseURL:        a4cAPI,
		username:       user,
		password:       password,
		checkWfTimeout: wfTimeout,
	}, nil
}

// do requests the alien4cloud rest api
func (c *a4cClient) do(method string, path string, body []byte, headers []Header) (*http.Response, error) {

	bodyBytes := bytes.NewBuffer(body)

	// Create the request
	request, err := http.NewRequest(method, c.baseURL+path, bodyBytes)
	if err != nil {
		return nil, err
	}

	// Add header
	for _, header := range headers {
		request.Header.Add(header.Key, header.Value)
	}

	response, err := c.Client.Do(request)
	if err != nil {
		return nil, err
	}

	// Cookie can potentially be expired. If we are unauthorized to send a request, we should try to login again.
	if response.StatusCode == http.StatusForbidden {
		err = c.Login()
		if err != nil {
			return nil, err
		}

		bodyBytes = bytes.NewBuffer(body)

		request, err := http.NewRequest(method, c.baseURL+path, bodyBytes)
		if err != nil {
			return nil, err
		}

		for _, header := range headers {
			request.Header.Add(header.Key, header.Value)
		}

		response, err := c.Client.Do(request)
		if err != nil {
			return nil, err
		}

		return response, nil
	}

	return response, nil
}

// Login login to alien4cloud
func (c *a4cClient) Login() error {
	//	request, err := http.NewRequest("POST", fmt.Sprintf("%s/login?username=%s&password=%s&submit=Login", c.baseURL, c.username, c.password), nil)
	values := url.Values{}
	values.Set("username", c.username)
	values.Set("password", c.password)
	values.Set("submit", "Login")
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/login", c.baseURL),
		strings.NewReader(values.Encode()))
	if err != nil {
		log.Panic(err)
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.Client.Do(request)

	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return getError(response.Body)
	}

	return nil
}

// Logout log out from alien4cloud
func (c *a4cClient) Logout() error {
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/logout", c.baseURL), nil)
	if err != nil {
		log.Panic(err)
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Set("Connection", "close")

	request.Close = true

	response, err := c.Client.Do(request)

	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return getError(response.Body)
	}

	return nil
}

//////////////////////////////////////
// Topology template related method //
//////////////////////////////////////

// getTopologyTemplateIDByName return the topology template ID for the given topologyName
func (c *a4cClient) getTopologyTemplateIDByName(topologyName string) (string, error) {

	toposSearchBody, err := json.Marshal(searchRequest{topologyName, "0", "1"})

	if err != nil {
		return "", errors.Wrap(err, "Cannot marshal an searchRequest structure")
	}

	response, err := c.do(
		"POST",
		fmt.Sprintf("%s/catalog/topologies/search", a4CRestAPIPrefix),
		[]byte(string(toposSearchBody)),
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return "", errors.Wrap(err, "Cannot send a request to get the topology id")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", errors.Wrap(err, "Cannot read the body of the request when getting topology id")
	}

	var res struct {
		Data struct {
			Types []string `json:"types"`
			Data  []struct {
				ID          string `json:"id"`
				ArchiveName string `json:"name"`
			} `json:"data"`
			TotalResults int `json:"totalResults"`
		} `json:"data"`
	}

	if err = json.Unmarshal([]byte(responseBody), &res); err != nil {
		return "", errors.Wrap(err, "Cannot unmarshal the request to get topology id")
	}

	if res.Data.TotalResults <= 0 {
		return "", fmt.Errorf("'%s' topology template does not exist", topologyName)
	}

	templateID := res.Data.Data[0].ID

	return templateID, nil
}

///////////////////////////////////////////////
// Methods related to application management //
///////////////////////////////////////////////

// Application represent fields of an application returned by A4C
// TODO represent more fields of an application returned by A4C
type Application struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Tags []Tag
}

// CreateAppli Create an application from a template and return its ID
func (c *a4cClient) CreateAppli(appName string, appTemplate string) (string, error) {

	topologyTemplateID, err := c.getTopologyTemplateIDByName(appTemplate)

	var appID string
	if err != nil {
		return appID, errors.Wrapf(err, "Unable to get the topology template id of template '%s'", appTemplate)
	}

	appliCreateJSON, err := json.Marshal(
		ApplicationCreateRequest{
			appName,
			appName,
			topologyTemplateID,
		},
	)

	if err != nil {
		return appID, errors.Wrap(err, "Cannot marshal an a4cAppliCreateRequestIn structure")
	}

	response, err := c.do(
		"POST",
		fmt.Sprintf("%s/applications", a4CRestAPIPrefix),
		[]byte(string(appliCreateJSON)),
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return appID, errors.Wrap(err, "Cannot send a request to create an application")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		return appID, getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return appID, errors.Wrap(err, "Cannot read the body of the result of the application creation")
	}

	var appStruct struct {
		Data string `json:"data"`
	}

	err = json.Unmarshal(responseBody, &appStruct)
	if err != nil {
		return appID, errors.Wrap(err, "Cannot unmarshal the reponse of the application creation")
	}

	appID = appStruct.Data

	return appID, err
}

// GetEnvironmentIDbyName Return the Alien4Cloud environment ID from a given application ID and environment name
func (c *a4cClient) GetEnvironmentIDbyName(appID string, envName string) (string, error) {

	envsSearchBody, err := json.Marshal(
		environmentsSearchRequest{
			"0",
			"20",
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "Cannot marshal an environmentsSearchRequest structure")
	}

	response, err := c.do(
		"POST",
		fmt.Sprintf("%s/applications/%s/environments/search", a4CRestAPIPrefix, appID),
		[]byte(string(envsSearchBody)),
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return "", errors.Wrap(err, "Unable to send request to get environment ID from its name of an application")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", errors.Wrapf(err, "Cannot read the body of the search for '%s' environment", envName)
	}

	var res struct {
		Data struct {
			Types []string `json:"types"`
			Data  []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"data"`
		} `json:"data"`
	}

	if err = json.Unmarshal([]byte(responseBody), &res); err != nil {
		return "", errors.Wrapf(err, "Cannot convert the body of the search for '%s' environment", envName)
	}

	var envID string
	for i := range res.Data.Data {
		if res.Data.Data[i].Name == envName {
			envID = res.Data.Data[i].ID
			break
		}
	}

	if envID == "" {
		return envID, fmt.Errorf("'%s' environment for application '%s' not found", envName, appID)
	}
	return envID, nil
}

// IsApplicationExist Return true if the application with the given ID exists
func (c *a4cClient) IsApplicationExist(applicationID string) (bool, error) {

	response, err := c.do(
		"GET",
		fmt.Sprintf("%s/applications/%s", a4CRestAPIPrefix, applicationID),
		nil,
		nil,
	)

	if err != nil {
		return false, errors.Wrap(err, "Cannot send a request to ensure an application exists")
	}
	defer response.Body.Close()

	switch response.StatusCode {

	case http.StatusOK:
		return true, nil

	case http.StatusNotFound:
		return false, nil

	default:
		return false, getError(response.Body)
	}
}

// GetApplicationsID returns the application ID using the given filter
func (c *a4cClient) GetApplicationsID(filter string) ([]string, error) {

	appsSearchBody, err := json.Marshal(
		searchRequest{
			filter,
			"0",
			"",
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Cannot marshal an searchRequest structure")
	}

	response, err := c.do(
		"POST",
		fmt.Sprintf("%s/applications/search", a4CRestAPIPrefix),
		[]byte(string(appsSearchBody)),
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to send request to search A4C application")
	}
	defer response.Body.Close()

	switch response.StatusCode {
	default:
		return nil, getError(response.Body)

	case http.StatusNotFound:
		// No application with this filter have been found
		return nil, nil

	case http.StatusOK:

		responseBody, err := ioutil.ReadAll(response.Body)

		if err != nil {
			return nil, errors.Wrap(err, "Unable to read the response of A4C application list request")
		}

		var res struct {
			Data struct {
				Types []string `json:"types"`
				Data  []struct {
					ID          string `json:"id"`
					ArchiveName string `json:"name"`
				} `json:"data"`
				TotalResults int `json:"totalResults"`
			} `json:"data"`
			Error Error `json:"error"`
		}

		if err = json.Unmarshal([]byte(responseBody), &res); err != nil {
			return nil, errors.Wrap(err, "Unable to unmarshal the response of A4C application list request")
		}

		if res.Data.TotalResults <= 0 {
			// No result have been returned
			return nil, nil
		}

		applicationIds := []string{}

		for _, application := range res.Data.Data {
			applicationIds = append(applicationIds, application.ID)
		}

		return applicationIds, nil
	}

}

// GetApplicationByID returns the application for the given ID
func (c *a4cClient) GetApplicationByID(id string) (*Application, error) {

	appsSearchBody, err := json.Marshal(
		searchRequest{
			id,
			"0",
			"1",
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Cannot marshal an searchRequest structure")
	}

	response, err := c.do(
		"POST",
		fmt.Sprintf("%s/applications/search", a4CRestAPIPrefix),
		[]byte(string(appsSearchBody)),
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to send request to search A4C application")
	}
	defer response.Body.Close()

	switch response.StatusCode {
	default:
		return nil, getError(response.Body)

	case http.StatusNotFound:
		// No application with this filter have been found
		return nil, nil

	case http.StatusOK:

		responseBody, err := ioutil.ReadAll(response.Body)

		if err != nil {
			return nil, errors.Wrap(err, "Unable to read the response of A4C application request")
		}

		var res struct {
			Data struct {
				Types        []string      `json:"types"`
				Data         []Application `json:"data"`
				TotalResults int           `json:"totalResults"`
			} `json:"data"`
			Error Error `json:"error"`
		}

		if err = json.Unmarshal([]byte(responseBody), &res); err != nil {
			return nil, errors.Wrap(err, "Unable to unmarshal the response of A4C application request")
		}

		if res.Data.TotalResults <= 0 {
			// No result have been returned
			return nil, nil
		}

		if res.Data.Data != nil && len(res.Data.Data) > 0 {
			return &res.Data.Data[0], nil
		}
		return nil, errors.New("Unable to access the response Data (nil or empty)")
	}

}

// DeleteApplication delete an application
func (c *a4cClient) DeleteApplication(appID string) error {

	response, err := c.do(
		"DELETE",
		fmt.Sprintf("%s/applications/%s", a4CRestAPIPrefix, appID),
		nil,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send request to delete A4C application")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return getError(response.Body)
	}

	return nil
}

///////////////////////////////////////////////////
// Methods related to application tag management //
///////////////////////////////////////////////////

// Tag tag key/value json mapping
type Tag struct {
	Key   string `json:"name"`
	Value string `json:"value"`
}

// SetTagToApplication set tag tagKey/tagValue to application
func (c *a4cClient) SetTagToApplication(applicationID string, tagKey string, tagValue string) error {

	type tagToSet struct {
		Key   string `json:"tagKey"`
		Value string `json:"tagValue"`
	}

	tag, err := json.Marshal(tagToSet{
		Key:   tagKey,
		Value: tagValue,
	})

	if err != nil {
		return errors.Wrap(err, "Unable to marshal struct to set a tag")
	}

	response, err := c.do(
		"POST",
		fmt.Sprintf("%s/applications/%s/tags", a4CRestAPIPrefix, applicationID),
		[]byte(string(tag)),
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send request to set a tag to an application")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return getError(response.Body)
	}

	return nil
}

// GetApplicationTag returns the application tag for the given applicationID and tagKey
func (c *a4cClient) GetApplicationTag(applicationID string, tagKey string) (string, error) {

	application, err := c.GetApplicationByID(applicationID)

	if err != nil {
		return "", errors.Wrap(err, "Unable to get application")
	}

	if application == nil {
		return "", errors.New("Unable to get tag from an unknown application")
	}

	for _, tag := range application.Tags {
		if tag.Key == tagKey {
			return tag.Value, nil
		}
	}

	// If we get here, no tags with such key has been found.
	return "", fmt.Errorf("no tag with key '%s'", tagKey)

}

//////////////////////////////////////////////
// Methods related to deployment management //
//////////////////////////////////////////////

// Get matching locations where a given application can be deployed
func (c *a4cClient) GetLocationsMatching(topologyID string, envID string) ([]LocationMatch, error) {
	response, err := c.do(
		"GET",
		fmt.Sprintf("%s/topologies/%s/locations?environmentId=%s", a4CRestAPIPrefix, topologyID, envID),
		nil,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get locations matching topology for application '%s' in '%s' environment",
			topologyID, envID)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, getError(response.Body)
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
func (c *a4cClient) DeployApplication(appID string, envID string, location string) error {

	// get locations matching this application
	topologyID, err := c.getA4CTopologyID(appID, envID)
	if err != nil {
		return errors.Wrapf(err, "Unable to get application topology for app %s and env %s",
			appID, envID)
	}

	locationsMatch, err := c.GetLocationsMatching(topologyID, envID)
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
	response, err := c.do(
		"POST",
		fmt.Sprintf("%s/applications/%s/environments/%s/deployment-topology/location-policies", a4CRestAPIPrefix, appID, envID),
		[]byte(string(body)),
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send a request to set the location in order to deploy an application")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return getError(response.Body)
	}

	// Deploy the application a4cApplicationDeployhRequestIn
	appDeployBody, err := json.Marshal(
		ApplicationDeployRequest{
			envID,
			appID,
		},
	)
	response, err = c.do(
		"POST",
		fmt.Sprintf("%s/applications/deployment", a4CRestAPIPrefix),
		[]byte(string(appDeployBody)),
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send a request to deploy the application")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return getError(response.Body)
	}

	return nil
}

// GetDeploymentList returns the deployment list for the given appID and envID
func (c *a4cClient) GetDeploymentList(appID string, envID string) ([]Deployment, error) {

	response, err := c.do(
		"GET",
		fmt.Sprintf("%s/deployments/search?environmentId=%s&from=0&query=", a4CRestAPIPrefix, envID),
		nil,
		[]Header{},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to send request to get deployment list")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, getError(response.Body)
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
func (c *a4cClient) UndeployApplication(appID string, envID string) error {

	response, err := c.do(
		"DELETE",
		fmt.Sprintf("%s/applications/%s/environments/%s/deployment", a4CRestAPIPrefix, appID, envID),
		nil,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send request to undeploy A4C application")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return getError(response.Body)
	}

	return nil
}

// WaintUntilStateIs Wait until the state of an Alien4Cloud application is the one given as parameter.
func (c *a4cClient) WaintUntilStateIs(appID string, envID string, status string) error {
	for {
		a4cStatus, err := c.GetDeploymentStatus(appID, envID)

		if err != nil {
			return errors.Wrapf(err, "Unable to get status from application %s", appID)
		}

		if a4cStatus == status {
			return nil
		}

		time.Sleep(time.Second)
	}
}

// GetDeploymentStatus returns current deployment status for the given applicationID and environmentID
func (c *a4cClient) GetDeploymentStatus(applicationID string, environmentID string) (string, error) {

	body := []byte(fmt.Sprintf(`["%s"]`, applicationID))
	response, err := c.do(
		"POST",
		fmt.Sprintf("%s/applications/statuses", a4CRestAPIPrefix),
		body,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return "", errors.Wrap(err, "Cannot send a request to get the deployment status")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", errors.Wrapf(err, "Cannot read the body when getting deployment status")
	}

	var statusResponse struct {
		Data map[string]map[string]struct {
			EnvironmentName   string
			EnvironmentStatus string
		} `json:"data"`
		Error Error `json:"error"`
	}

	err = json.Unmarshal(responseBody, &statusResponse)

	if err != nil {
		return "", errors.Wrapf(err, "Unable to unmarshal the deployment status")
	}

	for _, application := range statusResponse.Data {
		for _, environment := range application {
			alienEnvironmentID, err := c.GetEnvironmentIDbyName(applicationID, environment.EnvironmentName)

			if err != nil {
				return "", err
			}

			if alienEnvironmentID == environmentID {
				return strings.ToLower(environment.EnvironmentStatus), nil
			}
		}
	}

	return "", errors.New("unable to get the deployment status")

}

// GetCurrentDeploymentID returns current deployment ID for the given applicationID and environmentID
func (c *a4cClient) GetCurrentDeploymentID(applicationID string, environmentID string) (string, error) {

	response, err := c.do(
		"GET",
		fmt.Sprintf("%s/applications/%s/environments/%s/active-deployment-monitored", a4CRestAPIPrefix, applicationID, environmentID),
		nil,
		[]Header{
			{
				"Accept",
				"application/json",
			},
		},
	)

	if err != nil {
		return "", errors.Wrapf(err, "Unable to retrieve the current deployment ID for app '%s'", applicationID)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", errors.Wrap(err, "Cannot read the body of the active deployment monitored request")
	}

	var res struct {
		Data struct {
			Deployment struct {
				ID string `json:"id"`
			} `json:"deployment"`
		} `json:"data"`
	}

	err = json.Unmarshal(responseBody, &res)

	if err != nil {
		return "", errors.Wrap(err, "Unable to unmarshal content of the get deployment monitored request")
	}

	return res.Data.Deployment.ID, nil

}

////////////////////////////////////////////////////////
// Methods related to application topology management //
////////////////////////////////////////////////////////

// editA4CTopology Edit the topology of an application
func (c *a4cClient) editA4CTopology(a4cCtx *TopologyEditorContext, a4cTopoEditorExecute TopologyEditor) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	if a4cCtx.TopologyID == "" {
		var err error
		a4cCtx.TopologyID, err = c.getA4CTopologyID(a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}

	topoEditorExecuteBody, err := json.Marshal(a4cTopoEditorExecute)

	if err != nil {
		return errors.Wrap(err, "Cannot marshal an a4cTopoEditorExecuteRequestIn structure")
	}

	response, err := c.do(
		"POST",
		fmt.Sprintf("%s/editor/%s/execute", a4CRestAPIPrefix, a4cCtx.TopologyID),
		[]byte(string(topoEditorExecuteBody)),
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
			{
				"Accept",
				"application/json",
			},
		},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send the request edit an A4C topology")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return getError(response.Body)
	}
	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return errors.Wrap(err, "Unable to read the content of a topology edition request")
	}

	var resExec struct {
		Data struct {
			LastOperationIndex int `json:"lastOperationIndex"`
			Operations         []struct {
				PreviousOperationID string `json:"id"`
			} `json:"operations"`
		} `json:"data"`
	}

	if err = json.Unmarshal([]byte(responseBody), &resExec); err != nil {
		return errors.Wrap(err, "Unable to unmarshal a topology edition response")
	}

	lastOperationIndex := resExec.Data.LastOperationIndex
	a4cCtx.PreviousOperationID = resExec.Data.Operations[lastOperationIndex].PreviousOperationID

	return nil
}

// getA4CTopologyID method returns the A4C topology ID on a given application and environment
func (c *a4cClient) getA4CTopologyID(appID string, envID string) (string, error) {

	response, err := c.do(
		"GET",
		fmt.Sprintf("%s/applications/%s/environments/%s/topology", a4CRestAPIPrefix, appID, envID),
		nil,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return "", errors.Wrapf(err, "Cannot send a request in order to find the topology for application '%s' in '%s' environment", appID, envID)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", errors.Wrapf(err, "Cannot read the body of the topology get data for application '%s' in '%s' environment", appID, envID)
	}
	var res struct {
		Data string `json:"data"`
	}

	if err = json.Unmarshal([]byte(responseBody), &res); err != nil {
		return "", errors.Wrapf(err, "Cannot convert the body of topology get data for application '%s' in '%s' environment", appID, envID)
	}

	return res.Data, nil
}

// getA4CTopology method returns the A4C topology on a given application and environment
func (c *a4cClient) getA4CTopology(appID string, envID string) (*Topology, error) {

	a4cTopologyID, err := c.getA4CTopologyID(appID, envID)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", appID, envID)
	}

	response, err := c.do(
		"GET",
		fmt.Sprintf("%s/topologies/%s", a4CRestAPIPrefix, a4cTopologyID),
		nil,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get the topology content for application '%s' in '%s' environment", appID, envID)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, errors.Wrapf(err, "Cannot read the body of the topology get data for application '%s' in '%s' environment", appID, envID)
	}

	res := new(Topology)

	if err = json.Unmarshal([]byte(responseBody), res); err != nil {
		return nil, errors.Wrapf(err, "Cannot convert the body of topology get data for application '%s' in '%s' environment", appID, envID)
	}

	return res, nil
}

// DisplayNodeStatus displays the node status for the given applicationID and environmentID and nodeName if debug mode
func (c *a4cClient) DisplayNodeStatus(a4cApplicationID string, a4cEnvironmentID string, nodeName string) {

	var dsp bool
	dsp = false

	switch strings.ToUpper(os.Getenv("FMLE_DSP_NODE")) {
	case "DEBUG":
		dsp = true
	case "1":
		dsp = true
	}
	if dsp {
		nodeStatus, err := c.GetNodeStatus(a4cApplicationID, a4cEnvironmentID, nodeName)
		if err != nil {
			fmt.Printf("Unable to get A4C node status of app '%s' and env='%s' and node = %s: %v\n ", a4cApplicationID, a4cEnvironmentID, nodeName, err)
			//return
		} else {
			fmt.Printf("status node %s of \n\tapp '%s' \n\tenv='%s' \n\t\t==> nodeStatus=%s\n", nodeName, a4cApplicationID, a4cEnvironmentID, nodeStatus)
		}
	}
}

// GetNodeStatus returns the node status for the given applicationID and environmentID and nodeName
func (c *a4cClient) GetNodeStatus(applicationID string, environmentID string, nodeName string) (string, error) {

	response, err := c.do(
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
		return "", getError(response.Body)
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
func (c *a4cClient) GetOutputAttributes(applicationID string, environmentID string) (map[string][]string, error) {

	response, err := c.do(
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
		return nil, getError(response.Body)
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
func (c *a4cClient) GetAttributesValue(applicationID string, environmentID string, nodeName string, requestedAttributesName []string) (map[string]string, error) {

	response, err := c.do(
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
		return nil, getError(response.Body)
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

// UpdateComponentPropertyComplexType Update the property value of a component of an application when propertyValue is not a simple type (map, array..)
func (c *a4cClient) UpdateComponentPropertyComplexType(a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue map[string]interface{}) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	mapProp := propertyValue

	topoEditorExecute := TopologyEditorUpdateNodePropertyComplexType{
		TopologyEditorExecuteRequest: TopologyEditorExecuteRequest{
			NodeName:            componentName,
			PreviousOperationID: a4cCtx.PreviousOperationID,
			OperationType:       a4cUpdateNodePropertyValueOperationJavaClassName,
		},
		PropertyName:  propertyName,
		PropertyValue: mapProp,
	}

	if a4cCtx.TopologyID == "" {
		var err error
		a4cCtx.TopologyID, err = c.getA4CTopologyID(a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s\n", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}
	err := c.editA4CTopology(a4cCtx, topoEditorExecute)
	if err != nil {
		return errors.Wrapf(err, "UpdateComponentProperty : Unable to edit the topology of application '%s' and environment '%s'\n", a4cCtx.AppID, a4cCtx.EnvID)
	}

	return nil
}

// UpdateComponentProperty Update the property value of a component of an application
func (c *a4cClient) UpdateComponentProperty(a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue string) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	topoEditorExecute := TopologyEditorUpdateNodeProperty{
		TopologyEditorExecuteRequest: TopologyEditorExecuteRequest{
			NodeName:            componentName,
			PreviousOperationID: a4cCtx.PreviousOperationID,
			OperationType:       a4cUpdateNodePropertyValueOperationJavaClassName,
		},
		PropertyName:  propertyName,
		PropertyValue: propertyValue,
	}

	if a4cCtx.TopologyID == "" {
		var err error
		a4cCtx.TopologyID, err = c.getA4CTopologyID(a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s\n", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}
	err := c.editA4CTopology(a4cCtx, topoEditorExecute)

	if err != nil {
		return errors.Wrapf(err, "UpdateComponentProperty : Unable to edit the topology of application '%s' and environment '%s'\n", a4cCtx.AppID, a4cCtx.EnvID)
	}

	return nil
}

// UpdateCapabilityProperty Update the property value of a capability related to a component of an application
func (c *a4cClient) UpdateCapabilityProperty(a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue string, capabilityName string) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	topoEditorExecute := TopologyEditorUpdateCapabilityProperty{
		TopologyEditorExecuteRequest: TopologyEditorExecuteRequest{
			NodeName:            componentName,
			PreviousOperationID: a4cCtx.PreviousOperationID,
			OperationType:       a4cUpdateCapabilityPropertyValueOperationJavaClassName,
		},
		PropertyName:   propertyName,
		PropertyValue:  propertyValue,
		CapabilityName: capabilityName,
	}

	if a4cCtx.TopologyID == "" {
		var err error
		a4cCtx.TopologyID, err = c.getA4CTopologyID(a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}

	err := c.editA4CTopology(a4cCtx, topoEditorExecute)

	if err != nil {
		return errors.Wrapf(err, "Unable to edit the topology of application '%s' and environment '%s'", a4cCtx.AppID, a4cCtx.EnvID)
	}

	return nil
}

// AddNodeInA4CTopology Add a new node in the A4C topology
// not used any more
func (c *a4cClient) AddNodeInA4CTopology(a4cCtx *TopologyEditorContext, NodeTypeID string, nodeName string) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	a4cTopology, err := c.getA4CTopology(a4cCtx.AppID, a4cCtx.EnvID)

	if err != nil {
		return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
	}

	var nodeTypeVersion string

	for _, node := range a4cTopology.Data.NodeTypes {
		if NodeTypeID == node.ElementID {
			nodeTypeVersion = node.ArchiveVersion
		}
	}

	if reflect.DeepEqual(nodeTypeVersion, reflect.Zero(reflect.TypeOf(nodeTypeVersion)).Interface()) {
		return errors.Wrapf(err, "Unable to get archive version for node '%s' from A4C application topology for app %s and env %s", NodeTypeID, a4cCtx.AppID, a4cCtx.EnvID)
	}

	topoEditorExecute := TopologyEditorAddNode{
		TopologyEditorExecuteRequest: TopologyEditorExecuteRequest{
			NodeName:            nodeName,
			PreviousOperationID: a4cCtx.PreviousOperationID,
			OperationType:       a4cAddNodeOperationJavaClassName,
		},
		NodeTypeID: NodeTypeID + ":" + nodeTypeVersion,
	}

	if a4cCtx.TopologyID == "" {
		a4cCtx.TopologyID, err = c.getA4CTopologyID(a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}

	err = c.editA4CTopology(a4cCtx, topoEditorExecute)

	if err != nil {
		return errors.Wrapf(err, "Unable to edit the topology of application '%s' and environment '%s'", a4cCtx.AppID, a4cCtx.EnvID)
	}

	return nil
}

// AddRelationship Add a new relationship in the A4C topology
// not used any more
func (c *a4cClient) AddRelationship(a4cCtx *TopologyEditorContext, sourceNodeName string, targetNodeName string, relType string) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	var sourceNodeDef nodeType
	var targetNodeDef nodeType
	var requirementDef componentRequirement
	var relationshipDef relationshipType
	var capabilityDef componentCapability

	a4cTopology, err := c.getA4CTopology(a4cCtx.AppID, a4cCtx.EnvID)

	if err != nil {
		return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
	}

	for _, node := range a4cTopology.Data.Topology.NodeTemplates {

		if sourceNodeName == node.Name {
			for _, nodeDef := range a4cTopology.Data.NodeTypes {
				if node.Type == nodeDef.ElementID {
					sourceNodeDef = nodeDef
					break
				}
			}
		}

		if targetNodeName == node.Name {
			for _, nodeDef := range a4cTopology.Data.NodeTypes {
				if node.Type == nodeDef.ElementID {
					targetNodeDef = nodeDef
					break
				}
			}
		}

	}

	if reflect.DeepEqual(sourceNodeDef, reflect.Zero(reflect.TypeOf(sourceNodeDef)).Interface()) {
		return errors.New("Missing relationship source node attribute")
	}

	if reflect.DeepEqual(targetNodeDef, reflect.Zero(reflect.TypeOf(targetNodeDef)).Interface()) {
		return errors.New("Missing relationship target node attribute")
	}

	for _, req := range sourceNodeDef.Requirements {
		if relType == req.RelationshipType {
			requirementDef = req
		}
	}

	if reflect.DeepEqual(requirementDef, reflect.Zero(reflect.TypeOf(requirementDef)).Interface()) {
		return errors.New("Missing relationship requirement attribute")
	}

	for _, rel := range a4cTopology.Data.RelationshipTypes {
		if relType == rel.ElementID {
			relationshipDef = rel
		}
	}

	if reflect.DeepEqual(relationshipDef, reflect.Zero(reflect.TypeOf(relationshipDef)).Interface()) {
		return errors.New("Missing relationship type")
	}

	for _, c := range targetNodeDef.Capabilities {
		if requirementDef.Type == c.Type {
			capabilityDef = c
		}
	}

	if reflect.DeepEqual(capabilityDef, reflect.Zero(reflect.TypeOf(capabilityDef)).Interface()) {
		return errors.New("Missing relationship capability type")
	}

	relTmp := strings.Split(relType, ".")
	relationshipName := sourceNodeName + strings.Title(relTmp[len(relTmp)-1]) + strings.Title(targetNodeName)

	topoEditorExecute := TopologyEditorAddRelationships{
		TopologyEditorExecuteRequest: TopologyEditorExecuteRequest{
			NodeName:            sourceNodeName,
			OperationType:       a4cAddRelationshipOperationJavaClassName,
			PreviousOperationID: a4cCtx.PreviousOperationID,
		},
		RelationshipName:       relationshipName,
		RelationshipType:       relType,
		RelationshipVersion:    relationshipDef.ArchiveVersion,
		RequirementName:        requirementDef.ID,
		RequirementType:        requirementDef.Type,
		Target:                 targetNodeName,
		TargetedCapabilityName: capabilityDef.ID,
	}

	if a4cCtx.TopologyID == "" {
		a4cCtx.TopologyID, err = c.getA4CTopologyID(a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}

	err = c.editA4CTopology(a4cCtx, topoEditorExecute)

	if err != nil {
		return errors.Wrapf(err, "Unable to edit the topology of application '%s' and environment '%s'", a4cCtx.AppID, a4cCtx.EnvID)
	}

	return nil
}

// SaveA4CTopology saves the topology context
func (c *a4cClient) SaveA4CTopology(a4cCtx *TopologyEditorContext) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	if a4cCtx.TopologyID == "" {
		var err error
		a4cCtx.TopologyID, err = c.getA4CTopologyID(a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}

	response, err := c.do(
		"POST",
		fmt.Sprintf("%s/editor/%s?lastOperationId=%s", a4CRestAPIPrefix, a4cCtx.TopologyID, a4cCtx.PreviousOperationID),
		nil,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
			{
				"Accept",
				"application/json",
			},
		},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send the request to save an A4C topology")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return getError(response.Body)
	}

	// After saving topology, get come back to a clear state.
	a4cCtx.PreviousOperationID = ""

	return nil
}

/////////////////////////////////////////////////
// Methods related to orchestrators management //
/////////////////////////////////////////////////

// GetOrchestratorLocations returns the a4c location for orchestratorID
func (c *a4cClient) GetOrchestratorLocations(orchestratorID string) ([]Location, error) {
	// Get orchestrator location
	response, err := c.do(
		"GET",
		fmt.Sprintf("%s/orchestrators/%s/locations", a4CRestAPIPrefix, orchestratorID),
		nil,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to send request to get orchestrator location for orchestrator '%s'", orchestratorID)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to read the content of orchestrator locations request for orchestrator '%s'", orchestratorID)
	}

	var loc Location
	var locationstoreturn []Location

	var res struct {
		Data []struct {
			Location struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"location"`
		} `json:"data"`
	}
	if err = json.Unmarshal([]byte(responseBody), &res); err != nil {
		return nil, errors.Wrapf(err, "Cannot convert the body of the get '%s' orchestrator location", orchestratorID)
	}

	for _, orchestrator := range res.Data {
		loc.ID = orchestrator.Location.ID
		loc.Name = orchestrator.Location.Name

		locationstoreturn = append(locationstoreturn, loc)
	}

	return locationstoreturn, nil
}

// GetOrchestratorIDbyName Return the Alien4Cloud orchestrator ID from a given orchestator name
func (c *a4cClient) GetOrchestratorIDbyName(orchestratorName string) (string, error) {

	orchestratorsSearchBody, err := json.Marshal(searchRequest{orchestratorName, "0", "1"})

	if err != nil {
		return "", errors.Wrap(err, "Cannot marshal an searchRequest structure")
	}

	response, err := c.do(
		"GET",
		fmt.Sprintf("%s/orchestrators", a4CRestAPIPrefix),
		[]byte(string(orchestratorsSearchBody)),
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return "", errors.Wrap(err, "Unable to send request to get orchestrator ID from its name")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", errors.Wrapf(err, "Cannot read the body of the search for '%s' orchestrator", orchestratorName)
	}

	var res struct {
		Data struct {
			Data []struct {
				ID               string `json:"id"`
				OrchestratorName string `json:"name"`
			} `json:"data"`
			TotalResults int `json:"totalResults"`
		} `json:"data"`
	}

	if err = json.Unmarshal([]byte(responseBody), &res); err != nil {
		return "", errors.Wrapf(err, "Cannot convert the body of the search for '%s' orchestrator", orchestratorName)
	}
	if res.Data.TotalResults <= 0 {
		return "", fmt.Errorf("'%s' orchestrator name does not exist", orchestratorName)
	}

	orchestratorID := res.Data.Data[0].ID
	if orchestratorID == "" {
		return orchestratorID, fmt.Errorf("no ID for '%s' orchestrator", orchestratorName)
	}
	return orchestratorID, nil

}

///////////////////////////////////////
// Methods related to log management //
///////////////////////////////////////

// GetLogsOfApplication Returns the logs of the application and environment filtered
func (c *a4cClient) GetLogsOfApplication(applicationID string, environmentID string,
	filters LogFilter, fromIndex int) ([]Log, int, error) {

	deployments, err := c.GetDeploymentList(applicationID, environmentID)

	if err != nil {
		return nil, 0, errors.Wrapf(err, "Unable to get deployment list for app '%s' and env '%s'", applicationID, environmentID)
	}

	if len(deployments) <= 0 {
		return nil, 0, errors.New("The list of deployments item is empty. Unable to get logs from")
	}

	// The first step allow us to get the number of logs available. We will re-use the TotalResults parameters in order to generate the second request.

	logsFilter := logsSearchRequest{
		From: fromIndex,
		Size: 1,
		Filters: struct {
			LogFilter
			DeploymentID []string `json:"deploymentId,omitempty"`
		}{LogFilter: filters, DeploymentID: []string{deployments[0].ID}},
	}

	body, err := json.Marshal(logsFilter)

	if err != nil {
		return nil, 0, errors.Wrap(err, "Unable to marshal log filters in order to get the number of logs available for this deployment.")
	}

	response, err := c.do(
		"POST",
		fmt.Sprintf("%s/deployment/logs/search", a4CRestAPIPrefix),
		body,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return nil, 0, errors.Wrapf(err, "Cannot send a request to get number of logs from application '%s' and environment '%s'", applicationID, environmentID)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, 0, getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, 0, errors.Wrapf(err, "Cannot read the body of the log query response '%s' in '%s' environment", applicationID, environmentID)
	}

	var res struct {
		Data struct {
			Data         []Log `json:"data"`
			From         int   `json:"from"`
			To           int   `json:"to"`
			TotalResults int   `json:"totalResults"`
		} `json:"data"`
	}

	err = json.Unmarshal(responseBody, &res)

	if err != nil {
		return nil, 0, errors.Wrap(err, "Unable to unmarshal logs from orchestrator")
	}

	// Then we send the resquest to get all the logs returned for this deployment.

	logsFilter = logsSearchRequest{
		From: fromIndex,
		Size: res.Data.TotalResults,
		Filters: struct {
			LogFilter
			DeploymentID []string `json:"deploymentId,omitempty"`
		}{LogFilter: filters, DeploymentID: []string{deployments[0].ID}},
		SortConfiguration: struct {
			Ascending bool   `json:"ascending"`
			SortBy    string `json:"sortBy"`
		}{Ascending: true, SortBy: "timestamp"},
	}

	body, err = json.Marshal(logsFilter)

	if err != nil {
		return nil, 0, errors.Wrap(err, "Unable to marshal log filters to get logs for the deployment.")
	}

	response, err = c.do(
		"POST",
		fmt.Sprintf("%s/deployment/logs/search", a4CRestAPIPrefix),
		body,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return nil, 0, errors.Wrapf(err, "Cannot send a request to get logs from application '%s' and environment '%s'", applicationID, environmentID)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, 0, getError(response.Body)
	}

	responseBody, err = ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, 0, errors.Wrapf(err, "Cannot read the body of the log query response '%s' in '%s' environment", applicationID, environmentID)
	}

	err = json.Unmarshal(responseBody, &res)

	if err != nil {
		return nil, 0, errors.Wrap(err, "Unable to unmarshal logs from orchestrator")
	}

	return res.Data.Data, len(res.Data.Data), nil

}

////////////////////////////////////////////
// Methods related to workflow management //
////////////////////////////////////////////

// RunWorkflow runs a4c workflowName workflow for the given a4cAppID and a4cEnvID
func (c *a4cClient) RunWorkflow(a4cAppID string, a4cEnvID string, workflowName string) (*WorkflowExecution, error) {

	// The Alien4Cloud endpoint to start a workflow in Alien4Cloud is synchronous and for now, never finishes (Alien4Cloud 2.1.0-SM7).
	go func() {
		response, err := c.do(
			"POST",
			fmt.Sprintf("%s/applications/%s/environments/%s/workflows/%s", a4CRestAPIPrefix, a4cAppID, a4cEnvID, workflowName),
			nil,
			[]Header{
				{
					"Accept",
					"application/json",
				},
			},
		)
		if err == nil {
			response.Body.Close()
		}
	}()

	t1 := time.Now()

	for i := 0; i < c.checkWfTimeout; i++ {

		t2 := time.Now()
		t3 := t2.Sub(t1)
		if t3.Seconds() > float64(c.checkWfTimeout) {
			break
		}
		// We try to get which workflow is executing. If its name is equal to the one we tried to launch, we consider, it's been launched.

		workflowExecution, err := c.GetLastWorkflowExecution(a4cAppID, a4cEnvID)

		if err != nil {
			return workflowExecution, errors.Wrapf(err, "Unable to ensure the workflow '%s' has been executed on app '%s'", workflowName, a4cAppID)
		}

		if workflowExecution.DisplayWorkflowName == workflowName {
			return workflowExecution, err
		}
		time.Sleep(time.Second)
	}

	return nil, errors.Errorf("Timeout while trying to launch the workflow '%s' for app '%s'", workflowName, a4cAppID)

}

// GetLastWorkflowExecution return a4c workflow execution for the given applicationID and environmentID
func (c *a4cClient) GetLastWorkflowExecution(applicationID string, environmentID string) (*WorkflowExecution, error) {

	deploymentID, err := c.GetCurrentDeploymentID(applicationID, environmentID)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to get current deployment ID")
	}

	response, err := c.do(
		"GET",
		fmt.Sprintf("%s/workflow_execution/%s", a4CRestAPIPrefix, deploymentID),
		nil,
		[]Header{
			{
				"Accept",
				"application/json",
			},
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get workflow status of application '%s'", applicationID)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, getError(response.Body)
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
