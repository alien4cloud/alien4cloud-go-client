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
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

// EventService is the interface to the service mamaging events
type EventService interface {
	// Returns a given number of events for a given deployed application from a given index
	// Events are sorted by date in descending order. This call returns as well
	// the total number of events on this application
	GetEventsForApplicationEnvironment(ctx context.Context, environmentID string, fromIndex, size int) ([]Event, int, error)
}

type eventService struct {
	client restClient
}

// GetEventsForApplicationEnvironment returns the events for the application environment
// Results are sorted by descending date. This call returns as well
// the total number of events on this application
func (e *eventService) GetEventsForApplicationEnvironment(ctx context.Context, environmentID string,
	fromIndex, size int) ([]Event, int, error) {

	var res struct {
		Data struct {
			Data         []Event `json:"data"`
			From         int     `json:"from"`
			To           int     `json:"to"`
			TotalResults int     `json:"totalResults"`
		} `json:"data"`
	}

	// Then we send the resquest to get the events returned for this deployment.
	evURL := fmt.Sprintf("%s/deployments/%s/events?from=%s&size=%s", a4CRestAPIPrefix, environmentID,
		url.QueryEscape(strconv.Itoa(fromIndex)), url.QueryEscape(strconv.Itoa(size)))
	response, err := e.client.doWithContext(ctx,
		"GET",
		evURL,
		nil,
		[]Header{acceptAppJSONHeader},
	)

	if err != nil {
		return nil, 0, errors.Wrapf(err, "Cannot send a request to get events from application environment '%s'", environmentID)
	}
	err = processA4CResponse(response, &res, http.StatusOK)
	if err != nil {
		return nil, 0, errors.Wrap(err, "Unable to unmarshal events from response")
	}

	return res.Data.Data, res.Data.TotalResults, nil
}
