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

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/alien4cloud/alien4cloud-go-client/v2/alien4cloud"
)

// Command arguments
var url, user, password, query string
var from, size int

func init() {
	// Initialize command arguments
	flag.StringVar(&url, "url", "http://localhost:8088", "Alien4Cloud URL")
	flag.StringVar(&user, "user", "admin", "User")
	flag.StringVar(&password, "password", "changeme", "Password")
	flag.StringVar(&query, "query", "", "string to query")
	flag.IntVar(&from, "from", 0, "Index from which to return users")
	flag.IntVar(&size, "size", 0, "Maximum number of users to return")

}

func main() {

	// Parsing command arguments
	flag.Parse()

	client, err := alien4cloud.NewClient(url, user, password, "", true)
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

	searchRequest := alien4cloud.SearchRequest{
		From:  from,
		Size:  size,
		Query: query,
	}
	users, totalNumber, err := client.UserService().SearchUsers(ctx, searchRequest)
	if err != nil {
		log.Panic(err)
	}

	for _, user := range users {
		fmt.Printf("User %s, roles: %v\n", user.Username, user.Roles)
	}
	fmt.Printf("Total number of users: %d\n", totalNumber)
}
