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
	"net/http"

	"github.com/pkg/errors"
)

//go:generate mockgen -destination=../a4cmocks/${GOFILE} -package a4cmocks . ApplicationService

// ApplicationService is the interface to the service managing Applications
type ApplicationService interface {
	// Creates an application from a template and return its ID
	CreateAppli(ctx context.Context, appName string, appTemplate string) (string, error)
	// Returns the Alien4Cloud environment ID from a given application ID and environment name
	GetEnvironmentIDbyName(ctx context.Context, appID string, envName string) (string, error)
	// Returns true if the application with the given ID exists
	IsApplicationExist(ctx context.Context, applicationID string) (bool, error)
	// Returns the application ID using the given filter
	GetApplicationsID(ctx context.Context, filter string) ([]string, error)
	// Returns the application with the given ID
	GetApplicationByID(ctx context.Context, id string) (*Application, error)
	// Deletes an application
	DeleteApplication(ctx context.Context, appID string) error
	// Sets a tag tagKey/tagValue for the application
	SetTagToApplication(ctx context.Context, applicationID string, tagKey string, tagValue string) error
	// Returns the tag value for the given application ID and tag key
	GetApplicationTag(ctx context.Context, applicationID string, tagKey string) (string, error)
	// Returns the deployment topology for an application given an environment
	GetDeploymentTopology(ctx context.Context, appID string, envID string) (*Topology, error)
}

type applicationService struct {
	client *a4cClient
}

// CreateAppli Create an application from a template and return its ID
func (a *applicationService) CreateAppli(ctx context.Context, appName string, appTemplate string) (string, error) {

	var appID string
	topologyTemplateID, err := a.client.topologyService.GetTopologyTemplateIDByName(ctx, appTemplate)
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

	request, err := a.client.NewRequest(ctx,
		"POST",
		fmt.Sprintf("%s/applications", a4CRestAPIPrefix),
		bytes.NewReader(appliCreateJSON))
	if err != nil {
		return appID, errors.Wrap(err, "Cannot create a request to create an application")
	}

	var appStruct struct {
		Data string `json:"data"`
	}
	response, err := a.client.Do(request)
	if err != nil {
		return appID, errors.Wrap(err, "Cannot send a request to create an application")
	}
	err = ReadA4CResponse(response, &appStruct)
	return appStruct.Data, errors.Wrap(err, "Cannot create an application")
}

// GetEnvironmentIDbyName Return the Alien4Cloud environment ID from a given application ID and environment name
func (a *applicationService) GetEnvironmentIDbyName(ctx context.Context, appID string, envName string) (string, error) {

	envsSearchBody, err := json.Marshal(
		SearchRequest{
			From: 0,
			Size: 0,
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "Cannot marshal a SearchRequest structure")
	}

	request, err := a.client.NewRequest(ctx,
		"POST",
		fmt.Sprintf("%s/applications/%s/environments/search", a4CRestAPIPrefix, appID),
		bytes.NewReader(envsSearchBody))

	if err != nil {
		return "", errors.Wrap(err, "Unable to create request to get environment ID from its name of an application")
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
	response, err := a.client.Do(request)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to create request to get environment ID named '%s' for application '%s'", envName, appID)
	}
	err = ReadA4CResponse(response, &res)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to get environment ID for environment named '%s' in application '%s'", envName, appID)
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
func (a *applicationService) IsApplicationExist(ctx context.Context, applicationID string) (bool, error) {

	request, err := a.client.NewRequest(ctx,
		"GET",
		fmt.Sprintf("%s/applications/%s", a4CRestAPIPrefix, applicationID),
		nil)

	if err != nil {
		return false, errors.Wrap(err, "Cannot create a request to ensure an application exists")
	}

	response, err := a.client.Do(request)
	if err != nil {
		return false, errors.Wrap(err, "Can't check if an application exists")
	}

	switch response.StatusCode {
	case http.StatusNotFound:
		discardHTTPResponseBody(response)
		return false, nil

	default:
		err = ReadA4CResponse(response, nil)
		return err == nil, errors.Wrap(err, "Can't check if an application exists")
	}
}

// GetApplicationsID returns the application ID using the given filter
func (a *applicationService) GetApplicationsID(ctx context.Context, filter string) ([]string, error) {

	appsSearchBody, err := json.Marshal(
		SearchRequest{
			filter,
			0,
			0,
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Cannot marshal a SearchRequest structure")
	}

	request, err := a.client.NewRequest(ctx,
		"POST",
		fmt.Sprintf("%s/applications/search", a4CRestAPIPrefix),
		bytes.NewReader(appsSearchBody))

	if err != nil {
		return nil, errors.Wrap(err, "Unable to create request to search A4C application")
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

	response, err := a.client.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to send request to search A4C application")
	}
	switch response.StatusCode {

	case http.StatusNotFound:
		discardHTTPResponseBody(response)
		// No application with this filter have been found
		return nil, nil

	default:
		err = ReadA4CResponse(response, &res)
		if err != nil {
			return nil, errors.Wrap(err, "Can't get applications IDs")
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

// GetApplicationByID returns the application with the given ID
func (a *applicationService) GetApplicationByID(ctx context.Context, id string) (*Application, error) {

	request, err := a.client.NewRequest(ctx,
		"GET",
		fmt.Sprintf("%s/applications/%s", a4CRestAPIPrefix, id),
		nil)

	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get application with ID '%s'", id)
	}

	// RuntimeTopology represents runtime topology from a4c rest api
	var res struct {
		Data  Application `json:"data,omitempty"`
		Error Error       `json:"error,omitempty"`
	}

	resp, err := a.client.Do(request)
	if err != nil {
		return nil, err
	}
	err = ReadA4CResponse(resp, &res)
	return &res.Data, errors.Wrapf(err, "Can't get application by ID: %q", id)

}

// DeleteApplication delete an application
func (a *applicationService) DeleteApplication(ctx context.Context, appID string) error {

	request, err := a.client.NewRequest(ctx,
		"DELETE",
		fmt.Sprintf("%s/applications/%s", a4CRestAPIPrefix, appID),
		nil)

	if err != nil {
		return errors.Wrap(err, "Unable to create request to delete A4C application")
	}
	response, err := a.client.Do(request)
	if err != nil {
		return errors.Wrap(err, "Unable to create request to delete A4C application")
	}

	err = ReadA4CResponse(response, nil)

	return errors.Wrapf(err, "Unable to delete A4C application with ID: %q", appID)
}

// SetTagToApplication set tag tagKey/tagValue to application
func (a *applicationService) SetTagToApplication(ctx context.Context, applicationID string, tagKey string, tagValue string) error {

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

	request, err := a.client.NewRequest(ctx,
		"POST",
		fmt.Sprintf("%s/applications/%s/tags", a4CRestAPIPrefix, applicationID),
		bytes.NewReader(tag))
	if err != nil {
		return errors.Wrap(err, "Unable to create request to set a tag to an application")
	}

	response, err := a.client.Do(request)
	if err != nil {
		return errors.Wrap(err, "Unable to create request to set a tag to an application")
	}
	err = ReadA4CResponse(response, nil)
	return errors.Wrapf(err, "Unable to set tags to an application")
}

// GetApplicationTag returns the tag value for the given application ID and tag key
func (a *applicationService) GetApplicationTag(ctx context.Context, applicationID string, tagKey string) (string, error) {

	application, err := a.GetApplicationByID(ctx, applicationID)

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

func (a *applicationService) GetDeploymentTopology(ctx context.Context, appID string, envID string) (*Topology, error) {
	request, err := a.client.NewRequest(ctx,
		"GET",
		fmt.Sprintf("%s/applications/%s/environments/%s/deployment-topology", a4CRestAPIPrefix, appID, envID),
		nil)

	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get the deployment topology content for application '%s' on environment '%s'", appID, envID)
	}

	res := new(Topology)
	resp, err := a.client.Do(request)
	if err != nil {
		return res, errors.Wrapf(err, "Cannot get the deployment topology content for application '%s' on environment '%s'", appID, envID)
	}
	err = ReadA4CResponse(resp, res)
	return res, errors.Wrapf(err, "Cannot get the deployment topology content for application '%s' on environment '%s'", appID, envID)
}
