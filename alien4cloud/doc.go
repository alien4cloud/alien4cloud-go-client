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

/*
Package alien4cloud provides a client for using the https://alien4cloud.github.io API.

Usage:
	import "github.com/alien4cloud/alien4cloud-go-client/v2/alien4cloud"	// with go modules enabled (GO111MODULE=on or outside GOPATH)
	import "github.com/alien4cloud/alien4cloud-go-client/alien4cloud"       // with go modules disabled

Then you could create a client and use the different services exposed by the Alien4Cloud API:

	client, err := alien4cloud.NewClient(url, user, password, caFile, skipSecure)
	if err != nil {
		log.Panic(err)
	}

	// Timeout after one minute (this is optional you can use a context without timeout or cancelation)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err = client.Login(ctx)
	if err != nil {
		log.Panic(err)
	}

	userDetails, err := client.UserService().GetUser(ctx, username)

NOTE: Using the https://pkg.go.dev/context package, allows to easily pass cancelation signals and deadlines
to API calls for handling a request.

For more sample code snippets, see the https://github.com/alien4cloud/alien4cloud-go-client/tree/master/examples directory.

*/
package alien4cloud
