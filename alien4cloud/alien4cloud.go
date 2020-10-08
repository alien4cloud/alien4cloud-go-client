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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
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

	// NewRequest allows to create a custom request to be sent to Alien4Cloud
	// given a Context, method, url path and optional body.
	//
	// If the provided body is also an io.Closer, the Client Do function will automatically
	// close it.
	// The body needs to be a ReadSeeker in order to rewind request on retries.
	//
	// NewRequestWithContext returns a Request suitable for use with Client.Do
	//
	// If body is of type *bytes.Reader or *strings.Reader, the returned
	// request's ContentLength is set to its
	// exact value (instead of -1)
	NewRequest(ctx context.Context, method, urlStr string, body io.ReadSeeker) (*http.Request, error)

	// Do sends an HTTP request and returns an HTTP response
	//
	// If the returned error is nil, the Response will contain a non-nil
	// Body which the user is expected to close. If the Body is not both
	// read to EOF and closed, the Client's underlying RoundTripper
	// (typically Transport) may not be able to re-use a persistent TCP
	// connection to the server for a subsequent "keep-alive" request.
	// ReadA4CResponse() helper function is typically used to do this.
	//
	// The request Body, if non-nil, will be closed by the Do function
	// even on errors.
	//
	// Optional Retry functions may be provided. Those functions are executed sequentially to determine
	// if and how a request should be retried. See Retry documentation for more details.
	// Note: a special Retry function is always added at the end of the retries list. It will
	// automatically retry 403 Forbidden errors by trying to call Client.Login first.
	// This is for backward compatibility.
	Do(req *http.Request, retries ...Retry) (*http.Response, error)
}

// Retry is a function called after sending a request.
// It allows to perform actions based on the given response before re-sending a request.
// A typical usecase is to automatically call the Client.Login() function when receiving a 403 Forbidden response.
//
// It is possible to alter the request to be sent by returning an updated request. But in most cases
// the given original request can safely be returned as it. This framework take care of rewinding the request body
// before giving it to retry functions.
//
// The retry algorithm is:
// - If a retry function returns an error the retry process is stopped and this error is returned
// - If a retry function returns a nil request the retry process continue and consider the next available retry function
// - If a retry function returns a non-nil request this request is used in a Client.Do() call
//
// Note: It is critical that if the response body is read in a retry function it should not be closed
// and somehow rewind to the begining.
type Retry func(client Client, request *http.Request, response *http.Response) (*http.Request, error)

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

// a4Client holds properties of an a4c client
type a4cClient struct {
	client   *http.Client
	baseURL  string
	username string
	password string

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

	c := &a4cClient{
		client: &http.Client{
			Transport:     tr,
			CheckRedirect: nil,
			Jar:           newJar(),
			Timeout:       0},

		baseURL:  a4cAPI,
		username: user,
		password: password,
	}

	c.applicationService = &applicationService{c}
	c.deploymentService = &deploymentService{c}
	c.eventService = &eventService{c}
	c.logService = &logService{c}
	c.orchestratorService = &orchestratorService{c}
	c.topologyService = &topologyService{c}
	c.catalogService = &catalogService{c}
	c.userService = &userService{c}
	return c, nil
}

// Login login to alien4cloud
func (c *a4cClient) Login(ctx context.Context) error {
	values := url.Values{}
	values.Set("username", c.username)
	values.Set("password", c.password)
	values.Set("submit", "Login")
	request, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/login", c.baseURL),
		strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	// Replace default content-type
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.client.Do(request)

	if err != nil {
		return err
	}
	return ReadA4CResponse(response, nil)
}

// Logout log out from alien4cloud
func (c *a4cClient) Logout(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/logout", c.baseURL), nil)
	if err != nil {
		return err
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Set("Connection", "close")

	request.Close = true

	response, err := c.client.Do(request)

	if err != nil {
		return err
	}
	return ReadA4CResponse(response, nil)
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
