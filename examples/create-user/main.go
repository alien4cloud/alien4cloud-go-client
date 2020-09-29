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
var url, user, password, username, firstname, lastname, email, userpassword string

type stringArrayFlags []string

var roles stringArrayFlags

func (i *stringArrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *stringArrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func init() {
	// Initialize command arguments
	flag.StringVar(&url, "url", "http://localhost:8088", "Alien4Cloud URL")
	flag.StringVar(&user, "user", "admin", "Admin user used to connect to Alien4Cloud")
	flag.StringVar(&password, "password", "changeme", "Admin password")
	flag.StringVar(&username, "username", "", "name of user to create")
	flag.StringVar(&userpassword, "userpassword", "", "name of user to create")
	flag.StringVar(&firstname, "firstname", "", "name of user to create")
	flag.StringVar(&lastname, "lastname", "", "name of user to create")
	flag.StringVar(&email, "email", "", "name of user to create")

	flag.Var(&roles, "role", "Role to add (repeat this flag several times to add several roles")
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

	createRequest := alien4cloud.CreateUpdateUserRequest{
		UserName: username, FirstName: firstname, LastName: lastname, Email: email, Roles: roles,
		Password: userpassword,
	}
	err = client.UserService().CreateUser(ctx, createRequest)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("User %s created.\n", username)
}
