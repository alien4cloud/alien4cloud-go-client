package alien4cloud

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"encoding/json"

	"github.com/pkg/errors"
)

func (d *deploymentService) GetExecutions(ctx context.Context, deploymentID, query string, from, size int) ([]WorkflowExecution, FacetedSearchResult, error) {
	u := fmt.Sprintf("%s/executions/search?from=%s&size=%s", a4CRestAPIPrefix, url.QueryEscape(strconv.Itoa(from)), url.QueryEscape(strconv.Itoa(size)))

	if deploymentID != "" {
		u = fmt.Sprintf("%s&deploymentId=%s", u, url.QueryEscape(deploymentID))
	}

	if query != "" {
		u = fmt.Sprintf("%s&query=%s", u, url.QueryEscape(query))
	}
	response, err := d.client.doWithContext(ctx,
		"GET",
		u,
		nil,
		[]Header{acceptAppJSONHeader},
	)

	if err != nil {
		return nil, FacetedSearchResult{}, errors.Wrapf(err, "Failed to get executions for deployment %q", deploymentID)
	}

	var res struct {
		Data struct {
			Types []string            `json:"types"`
			Data  []WorkflowExecution `json:"data"`
			FacetedSearchResult
		} `json:"data"`
	}

	err = processA4CResponse(response, &res, http.StatusOK)
	return res.Data.Data, res.Data.FacetedSearchResult, errors.Wrapf(err, "Cannot convert the body response to request on executions for deployment %q", deploymentID)
}

func (d *deploymentService) CancelExecution(ctx context.Context, environmentID string, executionID string) error {


	cancelExecBody, err := json.Marshal(
		CancelExecRequest{
			EnvironmentID: environmentID,
			ExecutionID: executionID,
		},
	)
	if err != nil {
		return errors.Wrap(err, "Cannot marshal a cancelExecRequest structure")
	}

	_, err = d.client.doWithContext(ctx,
		"POST",
		fmt.Sprintf("%s/executions/cancel", a4CRestAPIPrefix),
		[]byte(string(cancelExecBody)),
		[]Header{contentTypeAppJSONHeader, acceptAppJSONHeader},
	)
	
	if err != nil {
		return errors.Wrapf(err, "Failed to cancel execution for execution '%s' on environment '%s'", executionID, environmentID)
	}

	return nil
}
