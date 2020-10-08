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

// LogService is the interface to the service mamaging logs
type LogService interface {
	// Returns the logs of the application and environment filtered
	GetLogsOfApplication(ctx context.Context, applicationID string, environmentID string, filters LogFilter, fromIndex int) ([]Log, int, error)
}

type logService struct {
	client *a4cClient
}

// GetLogsOfApplication returns the logs of the application and environment filtered
func (l *logService) GetLogsOfApplication(ctx context.Context, applicationID string, environmentID string,
	filters LogFilter, fromIndex int) ([]Log, int, error) {

	deployments, err := l.client.deploymentService.GetDeploymentList(ctx, applicationID, environmentID)

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

	request, err := l.client.NewRequest(ctx,
		"POST",
		fmt.Sprintf("%s/deployment/logs/search", a4CRestAPIPrefix),
		bytes.NewReader(body),
	)

	if err != nil {
		return nil, 0, errors.Wrapf(err, "Cannot create a request to get number of logs from application '%s' and environment '%s'", applicationID, environmentID)
	}
	var res struct {
		Data struct {
			Data         []Log `json:"data"`
			From         int   `json:"from"`
			To           int   `json:"to"`
			TotalResults int   `json:"totalResults"`
		} `json:"data"`
	}

	response, err := l.client.Do(request)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "Cannot send a request to get number of logs from application '%s' and environment '%s'", applicationID, environmentID)
	}
	err = ReadA4CResponse(response, &res)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "Cannot get number of logs from application '%s' and environment '%s'", applicationID, environmentID)
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

	request, err = l.client.NewRequest(ctx,
		"POST",
		fmt.Sprintf("%s/deployment/logs/search", a4CRestAPIPrefix),
		bytes.NewReader(body),
	)

	if err != nil {
		return nil, 0, errors.Wrapf(err, "Cannot create a request to get logs from application '%s' and environment '%s'", applicationID, environmentID)
	}
	response, err = l.client.Do(request)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "Cannot send a request to get logs from application '%s' and environment '%s'", applicationID, environmentID)
	}
	err = ReadA4CResponse(response, &res)

	return res.Data.Data, len(res.Data.Data), errors.Wrapf(err, "Cannot get logs from application '%s' and environment '%s'", applicationID, environmentID)
}
