// Copyright 2020 Bull S.A.S. Atos Technologies - Bull, Rue Jean Jaures, B.P.68, 78340, Les Clayes-sous-Bois, France.
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
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

func (c *a4cClient) NewRequest(ctx context.Context, method, urlStr string, body io.ReadSeeker) (*http.Request, error) {
	var contentLength int64
	switch v := body.(type) {
	case *bytes.Reader:
		contentLength = int64(v.Len())
	case *strings.Reader:
		contentLength = int64(v.Len())
	}

	if body != nil {
		body = &nopCloserReadSeeker{body}
	}
	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+urlStr, body)
	if err != nil {
		return nil, err
	}

	if contentLength > 0 {
		request.ContentLength = contentLength
	}

	// Add default headers
	request.Header.Add(contentTypeHeaderName, appJSONHeader)
	request.Header.Add(acceptHeaderName, appJSONHeader)
	return request, nil
}

func (c *a4cClient) Do(request *http.Request, retries ...Retry) (*http.Response, error) {
	// Close request body if underling reader allows it.
	var ncrsBody *nopCloserReadSeeker
	if request.Body != nil {
		var ok bool
		ncrsBody, ok = request.Body.(*nopCloserReadSeeker)
		if ok {
			c, okCloser := ncrsBody.ReadSeeker.(io.Closer)
			if okCloser {
				defer c.Close()
			}
		}
	}

	// always add retry forbidden errors
	retriesWithDefaults := append(retries, retryForbidden)

	response, err := c.client.Do(request)
	if err != nil {
		return response, err
	}

	for _, retry := range retriesWithDefaults {
		if ncrsBody != nil {
			// Restart reading request body from the beginning
			ncrsBody.Seek(0, io.SeekStart)
		}
		req, err := retry(c, request, response)
		if err != nil {
			return response, err
		}
		if req != nil {
			// Before retrying we need to fully read and close this response
			discardHTTPResponseBody(response)
			return c.Do(req, retries...)
		}
	}

	return response, nil
}

// ReadA4CResponse is an helper function that allow to fully read and close a response body and
// unmarshal its json content into a provided data structure.
// If response status code is greather or equal to 400 it automatically parse an error response and
// returns it as a non-nil error.
func ReadA4CResponse(response *http.Response, data interface{}) error {
	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.Wrap(err, "Cannot read the response from Alien4Cloud")
	}
	if response.StatusCode >= 400 {
		var res struct {
			Error Error `json:"error"`
		}
		err = json.Unmarshal(responseBody, &res)
		if err != nil {
			return errors.Wrap(err, "Unable to unmarshal content of the Alien4Cloud error response")
		}
		return errors.New(res.Error.Message)
	}
	if data != nil {
		err = json.Unmarshal(responseBody, &data)
	}
	return errors.Wrap(err, "Unable to unmarshal content of the Alien4Cloud response")
}

func retryForbidden(client Client, request *http.Request, response *http.Response) (*http.Request, error) {
	if response.StatusCode != http.StatusForbidden {
		// Nothing to retry
		return nil, nil
	}
	err := client.Login(request.Context())
	return request, err
}

type nopCloserReadSeeker struct {
	io.ReadSeeker
}

func (nc nopCloserReadSeeker) Close() error {
	return nil
}

func (nc nopCloserReadSeeker) Read(p []byte) (n int, err error) {
	return nc.ReadSeeker.Read(p)
}

func (nc nopCloserReadSeeker) Seek(offset int64, whence int) (int64, error) {
	return nc.ReadSeeker.Seek(offset, whence)
}

func discardHTTPResponseBody(response *http.Response) error {
	defer response.Body.Close()
	_, err := io.Copy(ioutil.Discard, response.Body)
	return errors.Wrap(err, "failed to fully read and discard response body")
}
