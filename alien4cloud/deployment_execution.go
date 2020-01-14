package alien4cloud

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

func (d *deploymentService) GetExecutions(ctx context.Context, deploymentID, query string, from, size int) ([]WorkflowExecution, FacetedSearchResult, error) {
	u := fmt.Sprintf("%s/executions/search?environmentId=%s&from=%s&size=%s", a4CRestAPIPrefix, deploymentID, url.QueryEscape(strconv.Itoa(from)), url.QueryEscape(strconv.Itoa(from)))
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
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, FacetedSearchResult{}, getError(response.Body)
	}

	var res struct {
		Data struct {
			Types []string            `json:"types"`
			Data  []WorkflowExecution `json:"data"`
			FacetedSearchResult
		} `json:"data"`
	}

	err = readBodyData(response, &res)
	if err != nil {
		return nil, FacetedSearchResult{}, errors.Wrapf(err, "Cannot convert the body response to request on executions for deployment %q", deploymentID)
	}
	return res.Data.Data, res.Data.FacetedSearchResult, nil
}
