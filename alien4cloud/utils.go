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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/pkg/errors"
)

// ------------------------------------------
// Implementation of http.CookieJar interface
// ------------------------------------------

// jar structure used tO implement http.CookieJar interface
type jar struct {
	lk      sync.Mutex
	cookies map[string][]*http.Cookie
}

// newJar allows to create a Jar structure and initialize cookies field
func newJar() *jar {
	jar := new(jar)
	jar.cookies = make(map[string][]*http.Cookie)
	return jar
}

// SetCookies handles the receipt of the cookies in a reply for the
// given URL.  It may or may not choose to save the cookies, depending
// on the jar's policy and implementation.
func (jar *jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.lk.Lock()
	jar.cookies[u.Host] = cookies
	jar.lk.Unlock()
}

// Cookies returns the cookies to send in a request for the given URL.
// It is up to the implementation to honor the standard cookie use
// restrictions such as in RFC 6265.
func (jar *jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies[u.Host]
}

func processA4CResponse(response *http.Response, expectedData interface{}, expectedStatus int) error {
	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.Wrap(err, "Cannot read the response from Alien4Cloud")
	}
	if response.StatusCode != expectedStatus {
		var res struct {
			Error Error `json:"error"`
		}
		err = json.Unmarshal(responseBody, &res)
		if err != nil {
			return errors.Wrap(err, "Unable to unmarshal content of the Alien4Cloud response")
		}
		return errors.New(res.Error.Message)
	}
	if expectedData != nil {
		err = json.Unmarshal(responseBody, &expectedData)
	}
	return errors.Wrap(err, "Unable to unmarshal content of the Alien4Cloud response")
}
