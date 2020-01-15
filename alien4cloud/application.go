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
	"net/http"

	"github.com/pkg/errors"
)

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
}

type applicationService struct {
	client          restClient
	topologyService *topologyService
}

// CreateAppli Create an application from a template and return its ID
func (a *applicationService) CreateAppli(ctx context.Context, appName string, appTemplate string) (string, error) {

	var appID string
	topologyTemplateID, err := a.topologyService.GetTopologyTemplateIDByName(ctx, appTemplate)
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

	response, err := a.client.doWithContext(ctx,
		"POST",
		fmt.Sprintf("%s/applications", a4CRestAPIPrefix),
		[]byte(string(appliCreateJSON)),
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return appID, errors.Wrap(err, "Cannot send a request to create an application")
	}

	var appStruct struct {
		Data string `json:"data"`
	}
	err = processA4CResponse(response, &appStruct, http.StatusCreated)
	return appStruct.Data, errors.Wrap(err, "Cannot unmarshal the reponse of the application creation")
}

// GetEnvironmentIDbyName Return the Alien4Cloud environment ID from a given application ID and environment name
func (a *applicationService) GetEnvironmentIDbyName(ctx context.Context, appID string, envName string) (string, error) {

	envsSearchBody, err := json.Marshal(
		searchRequest{
			From: "0",
			Size: "20",
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "Cannot marshal a searchRequest structure")
	}

	response, err := a.client.doWithContext(ctx,
		"POST",
		fmt.Sprintf("%s/applications/%s/environments/search", a4CRestAPIPrefix, appID),
		[]byte(string(envsSearchBody)),
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return "", errors.Wrap(err, "Unable to send request to get environment ID from its name of an application")
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
	err = processA4CResponse(response, &res, http.StatusOK)
	if err != nil {
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
func (a *applicationService) IsApplicationExist(ctx context.Context, applicationID string) (bool, error) {

	response, err := a.client.doWithContext(ctx,
		"GET",
		fmt.Sprintf("%s/applications/%s", a4CRestAPIPrefix, applicationID),
		nil,
		nil,
	)

	if err != nil {
		return false, errors.Wrap(err, "Cannot send a request to ensure an application exists")
	}
	switch response.StatusCode {

	case http.StatusNotFound:
		// to fully read response
		_ = processA4CResponse(response, nil, http.StatusNotFound)
		return false, nil

	default:
		err = processA4CResponse(response, nil, http.StatusOK)
		return err == nil, err
	}
}

// GetApplicationsID returns the application ID using the given filter
func (a *applicationService) GetApplicationsID(ctx context.Context, filter string) ([]string, error) {

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

	response, err := a.client.doWithContext(ctx,
		"POST",
		fmt.Sprintf("%s/applications/search", a4CRestAPIPrefix),
		[]byte(string(appsSearchBody)),
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to send request to search A4C application")
	}
	defer response.Body.Close()

	switch response.StatusCode {

	case http.StatusNotFound:
		// No application with this filter have been found
		// to fully read response
		_ = processA4CResponse(response, nil, http.StatusNotFound)
		return nil, nil

	default:
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
		err = processA4CResponse(response, &res, http.StatusOK)
		if err != nil {
			return nil, err
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

	response, err := a.client.doWithContext(ctx,
		"POST",
		fmt.Sprintf("%s/applications/search", a4CRestAPIPrefix),
		[]byte(string(appsSearchBody)),
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to send request to search A4C application")
	}
	switch response.StatusCode {
	case http.StatusNotFound:
		// No application with this filter have been found
		// to fully read response
		_ = processA4CResponse(response, nil, http.StatusNotFound)
		return nil, nil

	default:

		var res struct {
			Data struct {
				Types        []string      `json:"types"`
				Data         []Application `json:"data"`
				TotalResults int           `json:"totalResults"`
			} `json:"data"`
			Error Error `json:"error"`
		}
		err = processA4CResponse(response, &res, http.StatusOK)
		if err != nil {
			return nil, err
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
func (a *applicationService) DeleteApplication(ctx context.Context, appID string) error {

	response, err := a.client.doWithContext(ctx,
		"DELETE",
		fmt.Sprintf("%s/applications/%s", a4CRestAPIPrefix, appID),
		nil,
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send request to delete A4C application")
	}
	return processA4CResponse(response, nil, http.StatusOK)
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

	response, err := a.client.doWithContext(ctx,
		"POST",
		fmt.Sprintf("%s/applications/%s/tags", a4CRestAPIPrefix, applicationID),
		[]byte(string(tag)),
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send request to set a tag to an application")
	}
	return processA4CResponse(response, nil, http.StatusOK)
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
