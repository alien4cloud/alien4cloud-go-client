package alien4cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

func (d *deploymentService) GetExecutions(ctx context.Context, deploymentID, query string, from, size int) ([]Execution, FacetedSearchResult, error) {
	u := fmt.Sprintf("%s/executions/search?from=%s&size=%s", a4CRestAPIPrefix, url.QueryEscape(strconv.Itoa(from)), url.QueryEscape(strconv.Itoa(size)))

	if deploymentID != "" {
		u = fmt.Sprintf("%s&deploymentId=%s", u, url.QueryEscape(deploymentID))
	}

	if query != "" {
		u = fmt.Sprintf("%s&query=%s", u, url.QueryEscape(query))
	}
	request, err := d.client.NewRequest(ctx,
		"GET",
		u,
		nil)

	if err != nil {
		return nil, FacetedSearchResult{}, errors.Wrapf(err, "Failed to get executions for deployment %q", deploymentID)
	}

	var res struct {
		Data struct {
			Types []string    `json:"types"`
			Data  []Execution `json:"data"`
			FacetedSearchResult
		} `json:"data"`
	}

	response, err := d.client.Do(request)
	if err != nil {
		return nil, res.Data.FacetedSearchResult, errors.Wrapf(err, "Cannot send request to get executions for deployment %q", deploymentID)
	}
	err = ReadA4CResponse(response, &res)
	return res.Data.Data, res.Data.FacetedSearchResult, errors.Wrapf(err, "Cannot response on get executions for deployment %q", deploymentID)
}

// GetExecution returns details of a given execution
// Returns an error if no execution with such ID was found
func (d *deploymentService) GetExecution(ctx context.Context, deploymentID, workflowName, executionID string) (Execution, error) {
	startIndex := 0
	size := 50
	var exec Execution
	var err error
	for {
		execs, res, err := d.GetExecutions(ctx, deploymentID, workflowName, startIndex, size)
		if err != nil {
			break
		}
		for _, e := range execs {
			if e.ID == executionID {
				exec = e
				return exec, err
			}
		}
		// Execution not found in this range
		if res.TotalResults < (size + startIndex) {
			return exec, errors.Errorf("Found no execution with ID %s for deployment %s workflow %s",
				executionID, deploymentID, workflowName)
		}
		startIndex = startIndex + size
		size = res.TotalResults
	}
	return exec, err
}

func (d *deploymentService) CancelExecution(ctx context.Context, environmentID string, executionID string) error {

	cancelExecBody, err := json.Marshal(
		CancelExecRequest{
			EnvironmentID: environmentID,
			ExecutionID:   executionID,
		},
	)
	if err != nil {
		return errors.Wrap(err, "Cannot marshal a cancelExecRequest structure")
	}

	request, err := d.client.NewRequest(ctx,
		"POST",
		fmt.Sprintf("%s/executions/cancel", a4CRestAPIPrefix),
		bytes.NewReader(cancelExecBody))

	if err != nil {
		return errors.Wrapf(err, "Failed to cancel execution for execution '%s' on environment '%s'", executionID, environmentID)
	}

	response, err := d.client.Do(request)
	if err != nil {
		return errors.Wrapf(err, "Failed to cancel execution for execution '%s' on environment '%s'", executionID, environmentID)
	}
	err = ReadA4CResponse(response, nil)
	return errors.Wrapf(err, "Failed to cancel execution for execution '%s' on environment '%s'", executionID, environmentID)
}
