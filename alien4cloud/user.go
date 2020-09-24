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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// UserService is the interface to the service mamaging users and groups
type UserService interface {
	// CreateUser creates a user
	CreateUser(ctx context.Context, createRequest CreateUserRequest) error
	// UpdateUser updates a user parameters
	UpdateUser(ctx context.Context, userName string, updateRequest UpdateUserRequest) error
	// GetUser returns the parameters of a user whose name is provided in argument
	GetUser(ctx context.Context, userName string) (*User, error)
	// GetUsers returns the parameters of users whose names are provided in argument
	GetUsers(ctx context.Context, userNames []string) ([]User, error)
	// DeleteUser deletes a user
	DeleteUser(ctx context.Context, userName string) error
	// AddRole adds a role to a user
	AddRole(ctx context.Context, userName, role string) Error
	// RemoveRole removes a role that was granted user
	RemoveRole(ctx context.Context, userName, role string) Error

	// CreateGroup creates a group and returns its identifier
	CreateGroup(ctx context.Context, createRequest CreateGroupRequest) (string, error)
	// UpdateGroup updates a group parameters
	UpdateGroup(ctx context.Context, groupID string, updateRequest UpdateGroupRequest) error
	// GetGroup returns the parameters of a group whose identifier is provided in argument
	GetGroup(ctx context.Context, groupID string) (Group, error)
	// GetGroups returns the parameters of groups whose identifiers are provided in argument
	GetGroups(ctx context.Context, groupIDs []string) ([]Group, error)
	// DeleteGroup deletes a group
	DeleteGroup(ctx context.Context, groupID string) error
}

type userService struct {
	client restClient
}

// CreateUser creates a user
func (u *userService) CreateUser(ctx context.Context, createRequest CreateUserRequest) error {

	req, err := json.Marshal(createRequest)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal create request")
	}

	response, err := u.client.doWithContext(ctx,
		"POST",
		fmt.Sprintf("%s/users", a4CRestAPIPrefix),
		req,
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send request to create a user")
	}
	return processA4CResponse(response, nil, http.StatusOK)
}

// UpdateUser updates a user parameters
func (u *userService) UpdateUser(ctx context.Context, userName string, updateRequest UpdateUserRequest) error {

	req, err := json.Marshal(updateRequest)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal update request")
	}

	response, err := u.client.doWithContext(ctx,
		"PUT",
		fmt.Sprintf("%s/users/%s", a4CRestAPIPrefix, userName),
		req,
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return errors.Wrapf(err, "Unable to send request to update user %s", userName)
	}
	return processA4CResponse(response, nil, http.StatusOK)
}

// GetUser returns the parameters of a user whose name is provided in argument
func (u *userService) GetUser(ctx context.Context, userName string) (*User, error) {
	response, err := u.client.doWithContext(ctx,
		"GET",
		fmt.Sprintf("%s/users/%s", a4CRestAPIPrefix, userName),
		nil,
		nil)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to send request to get user %s", userName)
	}

	var res struct {
		Data  User  `json:"data,omitempty"`
		Error Error `json:"error,omitempty"`
	}

	err = processA4CResponse(response, &res, http.StatusOK)
	if err != nil {
		return nil, err
	}

	return &res.Data, err
}

// GetUsers returns the parameters of a user whose name is provided in argument
func (u *userService) GetUsers(ctx context.Context, userNames []string) ([]User, error) {
	req, err := json.Marshal(userNames)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal uner names")
	}

	response, err := u.client.doWithContext(ctx,
		"POST",
		fmt.Sprintf("%s/users/getUsers", a4CRestAPIPrefix),
		req,
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to send request to get users %v", userNames)
	}

	var res struct {
		Data  []User `json:"data,omitempty"`
		Error Error  `json:"error,omitempty"`
	}

	err = processA4CResponse(response, &res, http.StatusOK)
	if err != nil {
		return nil, err
	}

	return res.Data, err
}

// DeleteUser deletes a user
func (u *userService) DeleteUser(ctx context.Context, userName string) error {

	response, err := u.client.doWithContext(ctx,
		"DELETE",
		fmt.Sprintf("%s/users/%s", a4CRestAPIPrefix, userName),
		nil,
		nil)

	if err != nil {
		return errors.Wrapf(err, "Unable to send request to delete user %s", userName)
	}
	return processA4CResponse(response, nil, http.StatusOK)
}

// AddRole adds a role to a user
func (u *userService) AddRole(ctx context.Context, userName, roleName string) error {

	response, err := u.client.doWithContext(ctx,
		"PUT",
		fmt.Sprintf("%s/users/%s/roles/%s", a4CRestAPIPrefix, userName, roleName),
		nil,
		nil)

	if err != nil {
		return errors.Wrapf(err, "Unable to send request to add role %s to user %s", roleName, userName)
	}
	return processA4CResponse(response, nil, http.StatusOK)
}

// RemoveRole removes a role to a user
func (u *userService) RemoveRole(ctx context.Context, userName, roleName string) error {

	response, err := u.client.doWithContext(ctx,
		"DELETE",
		fmt.Sprintf("%s/users/%s/roles/%s", a4CRestAPIPrefix, userName, roleName),
		nil,
		nil)

	if err != nil {
		return errors.Wrapf(err, "Unable to send request to delete role %s to user %s", roleName, userName)
	}
	return processA4CResponse(response, nil, http.StatusOK)
}

// CreateGroup creates a group
func (u *userService) CreateGroup(ctx context.Context, createRequest CreateGroupRequest) error {

	req, err := json.Marshal(createRequest)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal create request")
	}

	response, err := u.client.doWithContext(ctx,
		"POST",
		fmt.Sprintf("%s/groups", a4CRestAPIPrefix),
		req,
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send request to create a group")
	}
	return processA4CResponse(response, nil, http.StatusOK)
}

// UpdateGroup updates a group parameters
func (u *userService) UpdateGroup(ctx context.Context, groupID string, updateRequest UpdateGroupRequest) error {

	req, err := json.Marshal(updateRequest)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal update request")
	}

	response, err := u.client.doWithContext(ctx,
		"PUT",
		fmt.Sprintf("%s/groups/%s", a4CRestAPIPrefix, groupID),
		req,
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return errors.Wrapf(err, "Unable to send request to update group %s", groupID)
	}
	return processA4CResponse(response, nil, http.StatusOK)
}

// GetGroup returns the parameters of a group whose name is provided in argument
func (u *userService) GetGroup(ctx context.Context, groupID string) (*Group, error) {
	response, err := u.client.doWithContext(ctx,
		"GET",
		fmt.Sprintf("%s/groups/%s", a4CRestAPIPrefix, groupID),
		nil,
		nil)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to send request to get group %s", groupID)
	}

	var res struct {
		Data  Group `json:"data,omitempty"`
		Error Error `json:"error,omitempty"`
	}

	err = processA4CResponse(response, &res, http.StatusOK)
	if err != nil {
		return nil, err
	}

	return &res.Data, err
}

// GetGroups returns the parameters of a group whose name is provided in argument
func (u *userService) GetGroups(ctx context.Context, groupIDs []string) ([]Group, error) {
	req, err := json.Marshal(groupIDs)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to marshal uner names")
	}

	response, err := u.client.doWithContext(ctx,
		"POST",
		fmt.Sprintf("%s/groups/getGroups", a4CRestAPIPrefix),
		req,
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to send request to get groups %v", groupIDs)
	}

	var res struct {
		Data  []Group `json:"data,omitempty"`
		Error Error   `json:"error,omitempty"`
	}

	err = processA4CResponse(response, &res, http.StatusOK)
	if err != nil {
		return nil, err
	}

	return res.Data, err
}

// DeleteGroup deletes a group
func (u *userService) DeleteGroup(ctx context.Context, groupID string) error {

	response, err := u.client.doWithContext(ctx,
		"DELETE",
		fmt.Sprintf("%s/groups/%s", a4CRestAPIPrefix, groupID),
		nil,
		nil)

	if err != nil {
		return errors.Wrapf(err, "Unable to send request to delete group %s", groupID)
	}
	return processA4CResponse(response, nil, http.StatusOK)
}
