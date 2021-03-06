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

//go:generate mockgen -destination=../a4cmocks/${GOFILE} -package a4cmocks . UserService

// UserService is the interface to the service mamaging users and groups
type UserService interface {
	// CreateUser creates a user
	CreateUser(ctx context.Context, createRequest CreateUpdateUserRequest) error
	// UpdateUser updates a user parameters
	UpdateUser(ctx context.Context, userName string, updateRequest CreateUpdateUserRequest) error
	// GetUser returns the parameters of a user whose name is provided in argument
	GetUser(ctx context.Context, userName string) (User, error)
	// GetUsers returns the parameters of users whose names are provided in argument
	GetUsers(ctx context.Context, userNames []string) ([]User, error)
	// SearchUsers searches for users and returns an array of users as well as the
	// total number of users matching the search request
	SearchUsers(ctx context.Context, searchRequest SearchRequest) ([]User, int, error)
	// DeleteUser deletes a user
	DeleteUser(ctx context.Context, userName string) error
	// AddRole adds a role to a user
	AddRole(ctx context.Context, userName, role string) error
	// RemoveRole removes a role that was granted user
	RemoveRole(ctx context.Context, userName, role string) error

	// CreateGroup creates a group and returns its identifier
	CreateGroup(ctx context.Context, group Group) (string, error)
	// UpdateGroup updates a group parameters
	UpdateGroup(ctx context.Context, groupID string, group Group) error
	// GetGroup returns the parameters of a group whose identifier is provided in argument
	// returns nil if no such group was found
	GetGroup(ctx context.Context, groupID string) (Group, error)
	// GetGroups returns the parameters of groups whose identifiers are provided in argument
	GetGroups(ctx context.Context, groupIDs []string) ([]Group, error)
	// SearchGroups searches for groups and returns an array of groups as well as the
	// total number of groups matching the search request
	SearchGroups(ctx context.Context, searchRequest SearchRequest) ([]Group, int, error)
	// DeleteGroup deletes a group
	DeleteGroup(ctx context.Context, groupID string) error
}

type userService struct {
	client *a4cClient
}

const (
	userEndpointFormat  = "%s/users/%s"
	groupEndpointFormat = "%s/groups/%s"
)

// CreateUser creates a user
func (u *userService) CreateUser(ctx context.Context, createRequest CreateUpdateUserRequest) error {

	req, err := json.Marshal(createRequest)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal create request")
	}

	request, err := u.client.NewRequest(ctx,
		"POST",
		fmt.Sprintf("%s/users", a4CRestAPIPrefix),
		bytes.NewReader(req),
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send request to create a user")
	}

	response, err := u.client.Do(request)
	if err != nil {
		return errors.Wrap(err, "Unable to send request to create a user")
	}
	err = ReadA4CResponse(response, nil)
	return errors.Wrap(err, "Unable to create a user")
}

// UpdateUser updates a user parameters
func (u *userService) UpdateUser(ctx context.Context, userName string, updateRequest CreateUpdateUserRequest) error {

	req, err := json.Marshal(updateRequest)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal update request")
	}

	request, err := u.client.NewRequest(ctx,
		"PUT",
		fmt.Sprintf(userEndpointFormat, a4CRestAPIPrefix, userName),
		bytes.NewReader(req),
	)

	if err != nil {
		return errors.Wrapf(err, "Unable to send request to update user %s", userName)
	}
	response, err := u.client.Do(request)
	if err != nil {
		return errors.Wrapf(err, "Unable to send request to update user %s", userName)
	}
	err = ReadA4CResponse(response, nil)
	return errors.Wrapf(err, "Unable to update user %s", userName)
}

// GetUser returns the parameters of a user whose name is provided in argument
func (u *userService) GetUser(ctx context.Context, userName string) (User, error) {
	var res struct {
		Data  User  `json:"data,omitempty"`
		Error Error `json:"error,omitempty"`
	}

	request, err := u.client.NewRequest(ctx,
		"GET",
		fmt.Sprintf(userEndpointFormat, a4CRestAPIPrefix, userName),
		nil)

	if err != nil {
		return res.Data, errors.Wrapf(err, "Unable to send request to get user %s", userName)
	}

	response, err := u.client.Do(request)
	if err != nil {
		return res.Data, errors.Wrapf(err, "Unable to send request to get user %s", userName)
	}
	err = ReadA4CResponse(response, &res)
	return res.Data, errors.Wrapf(err, "Unable to send request to get user %s", userName)
}

// GetUsers returns the parameters of a user whose name is provided in argument
func (u *userService) GetUsers(ctx context.Context, userNames []string) ([]User, error) {
	req, err := json.Marshal(userNames)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal user names")
	}

	request, err := u.client.NewRequest(ctx,
		"POST",
		fmt.Sprintf("%s/users/getUsers", a4CRestAPIPrefix),
		bytes.NewReader(req),
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to send request to get users %v", userNames)
	}

	var res struct {
		Data  []User `json:"data,omitempty"`
		Error Error  `json:"error,omitempty"`
	}

	response, err := u.client.Do(request)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to send request to get users %v", userNames)
	}
	err = ReadA4CResponse(response, &res)
	return res.Data, errors.Wrapf(err, "Unable to send request to get users %v", userNames)
}

// SearchUsers searches for users and returns an array of users as well as the
// total number of users matching the search request
func (u *userService) SearchUsers(ctx context.Context, searchRequest SearchRequest) ([]User, int, error) {
	req, err := json.Marshal(searchRequest)
	if err != nil {
		return nil, 0, errors.Wrap(err, "Unable to marshal search request")
	}

	request, err := u.client.NewRequest(ctx,
		"POST",
		fmt.Sprintf("%s/users/search", a4CRestAPIPrefix),
		bytes.NewReader(req),
	)

	if err != nil {
		return nil, 0, errors.Wrapf(err, "Unable to send request to search users %v", searchRequest)
	}

	var res struct {
		Data struct {
			Data         []User `json:"data,omitempty"`
			TotalResults int    `json:"totalResults"`
		} `json:"data,omitempty"`
		Error Error `json:"error,omitempty"`
	}

	response, err := u.client.Do(request)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "Unable to send request to search users %v", searchRequest)
	}
	err = ReadA4CResponse(response, &res)
	return res.Data.Data, res.Data.TotalResults, errors.Wrapf(err, "Unable to send request to search users %v", searchRequest)
}

// DeleteUser deletes a user
func (u *userService) DeleteUser(ctx context.Context, userName string) error {

	request, err := u.client.NewRequest(ctx,
		"DELETE",
		fmt.Sprintf(userEndpointFormat, a4CRestAPIPrefix, userName),
		nil)

	if err != nil {
		return errors.Wrapf(err, "Unable to send request to delete user %s", userName)
	}
	response, err := u.client.Do(request)
	if err != nil {
		return errors.Wrapf(err, "Unable to send request to delete user %s", userName)
	}
	err = ReadA4CResponse(response, nil)
	return errors.Wrapf(err, "Unable to send request to delete user %s", userName)
}

// AddRole adds a role to a user
func (u *userService) AddRole(ctx context.Context, userName, roleName string) error {

	request, err := u.client.NewRequest(ctx,
		"PUT",
		fmt.Sprintf("%s/users/%s/roles/%s", a4CRestAPIPrefix, userName, roleName),
		nil)

	if err != nil {
		return errors.Wrapf(err, "Unable to send request to add role %s to user %s", roleName, userName)
	}
	response, err := u.client.Do(request)
	if err != nil {
		return errors.Wrapf(err, "Unable to send request to add role %s to user %s", roleName, userName)
	}
	err = ReadA4CResponse(response, nil)
	return errors.Wrapf(err, "Unable to add role %s to user %s", roleName, userName)
}

// RemoveRole removes a role to a user
func (u *userService) RemoveRole(ctx context.Context, userName, roleName string) error {

	request, err := u.client.NewRequest(ctx,
		"DELETE",
		fmt.Sprintf("%s/users/%s/roles/%s", a4CRestAPIPrefix, userName, roleName),
		nil)

	if err != nil {
		return errors.Wrapf(err, "Unable to send request to delete role %s to user %s", roleName, userName)
	}
	response, err := u.client.Do(request)
	if err != nil {
		return errors.Wrapf(err, "Unable to send request to delete role %s to user %s", roleName, userName)
	}
	err = ReadA4CResponse(response, nil)
	return errors.Wrapf(err, "Unable to delete role %s to user %s", roleName, userName)
}

// CreateGroup creates a group and returns the identifier of the created group
func (u *userService) CreateGroup(ctx context.Context, group Group) (string, error) {

	var groupID string
	req, err := json.Marshal(group)
	if err != nil {
		return groupID, errors.Wrap(err, "Unable to marshal create request")
	}

	request, err := u.client.NewRequest(ctx,
		"POST",
		fmt.Sprintf("%s/groups", a4CRestAPIPrefix),
		bytes.NewReader(req),
	)

	if err != nil {
		return groupID, errors.Wrap(err, "Unable to send request to create a group")
	}

	var res struct {
		Data  string `json:"data,omitempty"`
		Error Error  `json:"error,omitempty"`
	}

	response, err := u.client.Do(request)
	if err != nil {
		return groupID, errors.Wrap(err, "Unable to send request to create a group")
	}
	err = ReadA4CResponse(response, &res)
	return res.Data, errors.Wrap(err, "Unable to send request to create a group")

}

// UpdateGroup updates a group parameters
func (u *userService) UpdateGroup(ctx context.Context, groupID string, group Group) error {

	req, err := json.Marshal(group)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal update request")
	}

	request, err := u.client.NewRequest(ctx,
		"PUT",
		fmt.Sprintf(groupEndpointFormat, a4CRestAPIPrefix, groupID),
		bytes.NewReader(req),
	)

	if err != nil {
		return errors.Wrapf(err, "Unable to send request to update group %s", groupID)
	}
	response, err := u.client.Do(request)
	if err != nil {
		return errors.Wrapf(err, "Unable to send request to update group %s", groupID)
	}
	err = ReadA4CResponse(response, nil)
	return errors.Wrapf(err, "Unable to update group %s", groupID)
}

// GetGroup returns the parameters of a group whose name is provided in argument
func (u *userService) GetGroup(ctx context.Context, groupID string) (Group, error) {

	var res struct {
		Data  Group `json:"data,omitempty"`
		Error Error `json:"error,omitempty"`
	}

	request, err := u.client.NewRequest(ctx,
		"GET",
		fmt.Sprintf(groupEndpointFormat, a4CRestAPIPrefix, groupID),
		nil)

	if err != nil {
		return res.Data, errors.Wrapf(err, "Unable to send request to get group %s", groupID)
	}

	response, err := u.client.Do(request)
	if err != nil {
		return res.Data, errors.Wrapf(err, "Unable to send request to get group %s", groupID)
	}
	err = ReadA4CResponse(response, &res)
	return res.Data, errors.Wrapf(err, "Unable to get group %s", groupID)
}

// GetGroups returns the parameters of a group whose name is provided in argument
func (u *userService) GetGroups(ctx context.Context, groupIDs []string) ([]Group, error) {
	req, err := json.Marshal(groupIDs)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal group IDs")
	}

	request, err := u.client.NewRequest(ctx,
		"POST",
		fmt.Sprintf("%s/groups/getGroups", a4CRestAPIPrefix),
		bytes.NewReader(req),
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to send request to get groups %v", groupIDs)
	}

	var res struct {
		Data  []Group `json:"data,omitempty"`
		Error Error   `json:"error,omitempty"`
	}

	response, err := u.client.Do(request)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to send request to get groups %v", groupIDs)
	}
	err = ReadA4CResponse(response, &res)
	return res.Data, errors.Wrapf(err, "Unable to get groups %v", groupIDs)
}

// SearchGroups searches for groups and returns an array of groups as well as the
// total number of groups matching the search request
func (u *userService) SearchGroups(ctx context.Context, searchRequest SearchRequest) ([]Group, int, error) {
	req, err := json.Marshal(searchRequest)
	if err != nil {
		return nil, 0, errors.Wrap(err, "Unable to marshal search request")
	}

	request, err := u.client.NewRequest(ctx,
		"POST",
		fmt.Sprintf("%s/groups/search", a4CRestAPIPrefix),
		bytes.NewReader(req),
	)

	if err != nil {
		return nil, 0, errors.Wrapf(err, "Unable to send request to search groups %v", searchRequest)
	}

	var res struct {
		Data struct {
			Data         []Group `json:"data,omitempty"`
			TotalResults int     `json:"totalResults"`
		} `json:"data,omitempty"`
		Error Error `json:"error,omitempty"`
	}

	response, err := u.client.Do(request)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "Unable to send request to search groups %v", searchRequest)
	}
	err = ReadA4CResponse(response, &res)
	return res.Data.Data, res.Data.TotalResults, errors.Wrapf(err, "Unable to search groups %v", searchRequest)
}

// DeleteGroup deletes a group
func (u *userService) DeleteGroup(ctx context.Context, groupID string) error {

	request, err := u.client.NewRequest(ctx,
		"DELETE",
		fmt.Sprintf(groupEndpointFormat, a4CRestAPIPrefix, groupID),
		nil)

	if err != nil {
		return errors.Wrapf(err, "Unable to send request to delete group %s", groupID)
	}
	response, err := u.client.Do(request)
	if err != nil {
		return errors.Wrapf(err, "Unable to send request to delete group %s", groupID)
	}
	err = ReadA4CResponse(response, nil)
	return errors.Wrapf(err, "Unable to delete group %s", groupID)
}
