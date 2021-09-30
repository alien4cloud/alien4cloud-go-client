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

	"github.com/pkg/errors"
)

//go:generate mockgen -destination=../a4cmocks/${GOFILE} -package a4cmocks . OrchestratorService

// OrchestratorService is the interface to the service mamaging orchestrators
type OrchestratorService interface {
	// Returns the Alien4Cloud locations for orchestratorID
	GetOrchestratorLocations(ctx context.Context, orchestratorID string) ([]Location, error)
	// Returns the Alien4Cloud orchestrator ID from a given orchestator name
	GetOrchestratorIDbyName(ctx context.Context, orchestratorName string) (string, error)
}

type orchestratorService struct {
	client *a4cClient
}

// GetOrchestratorLocations returns the Alien4Cloud locations for orchestratorID
func (o *orchestratorService) GetOrchestratorLocations(ctx context.Context, orchestratorID string) ([]Location, error) {
	// Get orchestrator location
	request, err := o.client.NewRequest(ctx,
		"GET",
		fmt.Sprintf("%s/orchestrators/%s/locations", a4CRestAPIPrefix, orchestratorID),
		nil,
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create request to get orchestrator location for orchestrator '%s'", orchestratorID)
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
	response, err := o.client.Do(request)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to send request to get orchestrator location for orchestrator '%s'", orchestratorID)
	}
	err = ReadA4CResponse(response, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get orchestrator location for orchestrator '%s'", orchestratorID)
	}

	for _, orchestrator := range res.Data {
		loc.ID = orchestrator.Location.ID
		loc.Name = orchestrator.Location.Name

		locationstoreturn = append(locationstoreturn, loc)
	}

	return locationstoreturn, nil
}

// GetOrchestratorIDbyName Return the Alien4Cloud orchestrator ID from a given orchestator name
func (o *orchestratorService) GetOrchestratorIDbyName(ctx context.Context, orchestratorName string) (string, error) {

	orchestratorsSearchBody, err := json.Marshal(SearchRequest{orchestratorName, 0, 1, nil})

	if err != nil {
		return "", errors.Wrap(err, "Cannot marshal a SearchRequest structure")
	}

	request, err := o.client.NewRequest(ctx,
		"GET",
		fmt.Sprintf("%s/orchestrators", a4CRestAPIPrefix),
		bytes.NewReader(orchestratorsSearchBody),
	)

	if err != nil {
		return "", errors.Wrapf(err, "Unable to send request to get orchestrator ID from its name '%s'", orchestratorName)
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

	response, err := o.client.Do(request)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to send request to get orchestrator ID from its name '%s'", orchestratorName)
	}
	err = ReadA4CResponse(response, &res)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to get orchestrator ID from its name '%s'", orchestratorName)
	}
	if res.Data.TotalResults <= 0 {
		return "", errors.Errorf("'%s' orchestrator name does not exist", orchestratorName)
	}

	orchestratorID := res.Data.Data[0].ID
	if orchestratorID == "" {
		return orchestratorID, errors.Errorf("no ID for '%s' orchestrator", orchestratorName)
	}
	return orchestratorID, nil
}
