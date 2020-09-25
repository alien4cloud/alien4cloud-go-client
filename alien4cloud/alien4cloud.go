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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/goware/urlx"
	"github.com/pkg/errors"
)

// Client is the client interface to Alien4cloud
type Client interface {
	Login(ctx context.Context) error
	Logout(ctx context.Context) error

	ApplicationService() ApplicationService
	DeploymentService() DeploymentService
	EventService() EventService
	LogService() LogService
	OrchestratorService() OrchestratorService
	TopologyService() TopologyService
	CatalogService() CatalogService
	UserService() UserService
}

const (
	// DefaultEnvironmentName is the default name of the environment created by
	// Alien4Cloud for an application
	DefaultEnvironmentName = "Environment"
	// ApplicationDeploymentInProgress a4c status
	ApplicationDeploymentInProgress = "DEPLOYMENT_IN_PROGRESS"
	// ApplicationDeployed a4c status
	ApplicationDeployed = "DEPLOYED"
	// ApplicationUndeploymentInProgress a4c status
	ApplicationUndeploymentInProgress = "UNDEPLOYMENT_IN_PROGRESS"
	// ApplicationUndeployed a4c status
	ApplicationUndeployed = "UNDEPLOYED"
	// ApplicationError a4c status
	ApplicationError = "FAILURE"
	// ApplicationUpdateError a4c status
	ApplicationUpdateError = "UPDATE_FAILURE"
	// ApplicationUpdated a4c status
	ApplicationUpdated = "UPDATED"
	// ApplicationUpdateInProgress a4c status
	ApplicationUpdateInProgress = "UPDATE_IN_PROGRESS"

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

	// FunctionConcat is a function used in attribute/property values to concatenate strings
	FunctionConcat = "concat"
	// FunctionGetInput is a function used in attribute/property values to reference an input property
	FunctionGetInput = "get_input"

	// ROLE_ADMIN is the adminstrator role
	ROLE_ADMIN = "ADMIN"
	// ROLE_COMPONENTS_MANAGER allows to define packages on how to install, configure, start and connect components (mapped as node types)
	ROLE_COMPONENTS_MANAGER = "COMPONENTS_MANAGER"
	// ROLE_ARCHITECT allows to define application templates (topologies) by reusing building blocks (node types defined by components managers)
	ROLE_ARCHITECT = "ARCHITECT"
	// ROLE_APPLICATIONS_MANAGER allows to define applications with it’s own topologies that can be linked to a global topology from architects and that can reuse components defined by the components managers
	ROLE_APPLICATIONS_MANAGER = "APPLICATIONS_MANAGER"
)

const (
	// a4CRestAPIPrefix a4c rest api prefix
	a4CRestAPIPrefix string = "/rest/latest"
)

type restClient struct {
	*http.Client
	baseURL  string
	username string
	password string
}

// a4Client holds properties of an a4c client
type a4cClient struct {
	client              restClient
	applicationService  *applicationService
	deploymentService   *deploymentService
	eventService        *eventService
	logService          *logService
	orchestratorService *orchestratorService
	topologyService     *topologyService
	catalogService      *catalogService
	userService         *userService
}

// NewClient instanciates and returns Client
func NewClient(address string, user string, password string, caFile string, skipSecure bool) (Client, error) {
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

	restClient := restClient{
		Client: &http.Client{
			Transport:     tr,
			CheckRedirect: nil,
			Jar:           newJar(),
			Timeout:       0},
		baseURL:  a4cAPI,
		username: user,
		password: password,
	}
	topoService := topologyService{restClient}
	eventService := eventService{restClient}
	catService := catalogService{restClient}
	appService := applicationService{restClient, &topoService}
	deployService := deploymentService{restClient, &appService, &topoService}
	userService := userService{restClient}
	return &a4cClient{
		client:              restClient,
		applicationService:  &appService,
		deploymentService:   &deployService,
		eventService:        &eventService,
		logService:          &logService{restClient, &deployService},
		orchestratorService: &orchestratorService{restClient},
		topologyService:     &topoService,
		catalogService:      &catService,
		userService:         &userService,
	}, nil
}

// Login login to alien4cloud
func (c *a4cClient) Login(ctx context.Context) error {
	return c.client.login(ctx)
}

// Logout log out from alien4cloud
func (c *a4cClient) Logout(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/logout", c.client.baseURL), nil)
	if err != nil {
		return err
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Set("Connection", "close")

	request.Close = true

	response, err := c.client.Client.Do(request)

	if err != nil {
		return err
	}
	return processA4CResponse(response, nil, http.StatusOK)
}

// ApplicationService retrieves the Application Service
func (c *a4cClient) ApplicationService() ApplicationService {
	return c.applicationService
}

// DeploymentService retrieves the Deployment Service
func (c *a4cClient) DeploymentService() DeploymentService {
	return c.deploymentService
}

// Event retrieves the Event Service
func (c *a4cClient) EventService() EventService {
	return c.eventService
}

// LogService retrieves the Log Service
func (c *a4cClient) LogService() LogService {
	return c.logService
}

// OrchestratorService retrieves the Orchestrator Service
func (c *a4cClient) OrchestratorService() OrchestratorService {
	return c.orchestratorService
}

// TopologyService retrieves the Topology Service
func (c *a4cClient) TopologyService() TopologyService {
	return c.topologyService
}

// CatalogService retrieves the Catalog Service
func (c *a4cClient) CatalogService() CatalogService {
	return c.catalogService
}

// UserService retrieves the User Service
func (c *a4cClient) UserService() UserService {
	return c.userService
}

// do requests the alien4cloud rest api with a Context that can be canceled
func (r *restClient) doWithContext(ctx context.Context, method string, path string, body []byte, headers []Header) (*http.Response, error) {

	bodyBytes := bytes.NewBuffer(body)

	request, err := http.NewRequestWithContext(ctx, method, r.baseURL+path, bodyBytes)

	if err != nil {
		return nil, err
	}

	// Add header
	for _, header := range headers {
		request.Header.Add(header.Key, header.Value)
	}

	response, err := r.Client.Do(request)
	if err != nil {
		return nil, err
	}

	// Cookie can potentially be expired. If we are unauthorized to send a request, we should try to login again.
	if response.StatusCode == http.StatusForbidden {
		err = r.login(ctx)
		if err != nil {
			return nil, err
		}

		bodyBytes = bytes.NewBuffer(body)

		request, err := http.NewRequest(method, r.baseURL+path, bodyBytes)
		if err != nil {
			return nil, err
		}

		for _, header := range headers {
			request.Header.Add(header.Key, header.Value)
		}

		response, err := r.Client.Do(request)
		if err != nil {
			return nil, err
		}

		return response, nil
	}

	return response, nil
}

// login to alien4cloud
func (r *restClient) login(ctx context.Context) error {
	values := url.Values{}
	values.Set("username", r.username)
	values.Set("password", r.password)
	values.Set("submit", "Login")
	request, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/login", r.baseURL),
		strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := r.Client.Do(request)

	if err != nil {
		return err
	}
	return processA4CResponse(response, nil, http.StatusOK)
}
