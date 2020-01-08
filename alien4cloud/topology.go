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
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// TopologyService is the interface to the service mamaging topologies
type TopologyService interface {
	// Returns the topology ID on a given application and environment
	GetTopologyID(appID string, envID string) (string, error)
	// Returns the topology template ID for the given topologyName
	GetTopologyTemplateIDByName(topologyName string) (string, error)
	// Updates the property value (type string) of a component of an application
	UpdateComponentProperty(a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue string) error
	// Updates the property value (type tosca complex) of a component of an application
	UpdateComponentPropertyComplexType(a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue map[string]interface{}) error
	// Updates the property value of a capability related to a component of an application
	UpdateCapabilityProperty(a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue string, capabilityName string) error
	// Adds a new node in the A4C topology
	AddNodeInA4CTopology(a4cCtx *TopologyEditorContext, nodeTypeID string, nodeName string) error
	// Adds a new relationship in the A4C topology
	AddRelationship(a4cCtx *TopologyEditorContext, sourceNodeName string, targetNodeName string, relType string) error
	// Saves the topology context
	SaveA4CTopology(a4cCtx *TopologyEditorContext) error
	// Creates an empty workflow in the given topology
	CreateWorkflow(ctx context.Context, a4cCtx *TopologyEditorContext, workflowName string) error
	// Deletes a workflow in the given topology
	DeleteWorkflow(ctx context.Context, a4cCtx *TopologyEditorContext, workflowName string) error
	// Adds an activity to a workflow
	AddWorkflowActivity(ctx context.Context, a4cCtx *TopologyEditorContext, workflowName string, activity *WorkflowActivity) error
}

type topologyService struct {
	client restClient
}

const (
	// a4cUpdateNodePropertyValueOperationJavaClassName a4c class name to update node property value operation
	a4cUpdateNodePropertyValueOperationJavaClassName = "org.alien4cloud.tosca.editor.operations.nodetemplate.UpdateNodePropertyValueOperation"

	// a4cUpdateNodePropertyValueSlurmJobOptions yorc struct name for slurm JobOptions
	a4cUpdateNodePropertyValueSlurmJobOptions = "yorc.datatypes.slurm.JobOptions"

	// a4cUpdateCapabilityPropertyValueOperationJavaClassName a4c class name to update capability value operation
	a4cUpdateCapabilityPropertyValueOperationJavaClassName = "org.alien4cloud.tosca.editor.operations.nodetemplate.UpdateCapabilityPropertyValueOperation"

	// a4cAddNodeOperationJavaClassName a4c class name to add node operation
	a4cAddNodeOperationJavaClassName = "org.alien4cloud.tosca.editor.operations.nodetemplate.AddNodeOperation"

	// a4cAddRelationshipOperationJavaClassName a4c class name to add relationship operation
	a4cAddRelationshipOperationJavaClassName = "org.alien4cloud.tosca.editor.operations.relationshiptemplate.AddRelationshipOperation"
)

// GetTopologyID returns the A4C topology ID on a given application and environment
func (t *topologyService) GetTopologyID(appID string, envID string) (string, error) {

	response, err := t.client.do(
		"GET",
		fmt.Sprintf("%s/applications/%s/environments/%s/topology", a4CRestAPIPrefix, appID, envID),
		nil,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return "", errors.Wrapf(err, "Cannot send a request in order to find the topology for application '%s' in '%s' environment", appID, envID)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", errors.Wrapf(err, "Cannot read the body of the topology get data for application '%s' in '%s' environment", appID, envID)
	}
	var res struct {
		Data string `json:"data"`
	}

	if err = json.Unmarshal([]byte(responseBody), &res); err != nil {
		return "", errors.Wrapf(err, "Cannot convert the body of topology get data for application '%s' in '%s' environment", appID, envID)
	}

	return res.Data, nil
}

// GetTopologyTemplateIDByName return the topology template ID for the given topologyName
func (t *topologyService) GetTopologyTemplateIDByName(topologyName string) (string, error) {

	toposSearchBody, err := json.Marshal(searchRequest{topologyName, "0", "1"})

	if err != nil {
		return "", errors.Wrap(err, "Cannot marshal an searchRequest structure")
	}

	response, err := t.client.do(
		"POST",
		fmt.Sprintf("%s/catalog/topologies/search", a4CRestAPIPrefix),
		[]byte(string(toposSearchBody)),
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return "", errors.Wrap(err, "Cannot send a request to get the topology id")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", errors.Wrap(err, "Cannot read the body of the request when getting topology id")
	}

	var res struct {
		Data struct {
			Types []string `json:"types"`
			Data  []struct {
				ID          string `json:"id"`
				ArchiveName string `json:"name"`
			} `json:"data"`
			TotalResults int `json:"totalResults"`
		} `json:"data"`
	}

	if err = json.Unmarshal([]byte(responseBody), &res); err != nil {
		return "", errors.Wrap(err, "Cannot unmarshal the request to get topology id")
	}

	if res.Data.TotalResults <= 0 {
		return "", fmt.Errorf("'%s' topology template does not exist", topologyName)
	}

	templateID := res.Data.Data[0].ID

	return templateID, nil
}

// editTopology Edit the topology of an application
func (t *topologyService) editTopology(ctx context.Context, a4cCtx *TopologyEditorContext, a4cTopoEditorExecute TopologyEditor) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	if a4cCtx.TopologyID == "" {
		var err error
		a4cCtx.TopologyID, err = t.GetTopologyID(a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}

	topoEditorExecuteBody, err := json.Marshal(a4cTopoEditorExecute)

	if err != nil {
		return errors.Wrap(err, "Cannot marshal an a4cTopoEditorExecuteRequestIn structure")
	}

	response, err := t.client.doWithContext(ctx,
		"POST",
		fmt.Sprintf("%s/editor/%s/execute", a4CRestAPIPrefix, a4cCtx.TopologyID),
		[]byte(string(topoEditorExecuteBody)),
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
			{
				"Accept",
				"application/json",
			},
		},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send the request edit an A4C topology")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return getError(response.Body)
	}
	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return errors.Wrap(err, "Unable to read the content of a topology edition request")
	}

	var resExec struct {
		Data struct {
			LastOperationIndex int `json:"lastOperationIndex"`
			Operations         []struct {
				PreviousOperationID string `json:"id"`
			} `json:"operations"`
		} `json:"data"`
	}

	if err = json.Unmarshal([]byte(responseBody), &resExec); err != nil {
		return errors.Wrap(err, "Unable to unmarshal a topology edition response")
	}

	lastOperationIndex := resExec.Data.LastOperationIndex
	if len(resExec.Data.Operations) > lastOperationIndex {
		a4cCtx.PreviousOperationID = resExec.Data.Operations[lastOperationIndex].PreviousOperationID
	}

	return nil
}

// getTopology method returns the A4C topology on a given application and environment
func (t *topologyService) getTopology(appID string, envID string) (*Topology, error) {

	a4cTopologyID, err := t.GetTopologyID(appID, envID)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", appID, envID)
	}

	response, err := t.client.do(
		"GET",
		fmt.Sprintf("%s/topologies/%s", a4CRestAPIPrefix, a4cTopologyID),
		nil,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get the topology content for application '%s' in '%s' environment", appID, envID)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, getError(response.Body)
	}

	responseBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, errors.Wrapf(err, "Cannot read the body of the topology get data for application '%s' in '%s' environment", appID, envID)
	}

	res := new(Topology)

	if err = json.Unmarshal([]byte(responseBody), res); err != nil {
		return nil, errors.Wrapf(err, "Cannot convert the body of topology get data for application '%s' in '%s' environment", appID, envID)
	}

	return res, nil
}

// UpdateComponentPropertyComplexType Update the property value of a component of an application when propertyValue is not a simple type (map, array..)
func (t *topologyService) UpdateComponentPropertyComplexType(a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue map[string]interface{}) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	mapProp := propertyValue

	topoEditorExecute := TopologyEditorUpdateNodePropertyComplexType{
		TopologyEditorExecuteNodeRequest: TopologyEditorExecuteNodeRequest{
			NodeName: componentName,
			TopologyEditorExecuteRequest: TopologyEditorExecuteRequest{
				PreviousOperationID: a4cCtx.PreviousOperationID,
				OperationType:       a4cUpdateNodePropertyValueOperationJavaClassName,
			},
		},
		PropertyName:  propertyName,
		PropertyValue: mapProp,
	}

	if a4cCtx.TopologyID == "" {
		var err error
		a4cCtx.TopologyID, err = t.GetTopologyID(a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s\n", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}
	err := t.editTopology(nil, a4cCtx, topoEditorExecute)
	if err != nil {
		return errors.Wrapf(err, "UpdateComponentProperty : Unable to edit the topology of application '%s' and environment '%s'\n", a4cCtx.AppID, a4cCtx.EnvID)
	}

	return nil
}

// UpdateComponentProperty Update the property value of a component of an application
func (t *topologyService) UpdateComponentProperty(a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue string) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	topoEditorExecute := TopologyEditorUpdateNodeProperty{
		TopologyEditorExecuteNodeRequest: TopologyEditorExecuteNodeRequest{
			NodeName: componentName,
			TopologyEditorExecuteRequest: TopologyEditorExecuteRequest{
				PreviousOperationID: a4cCtx.PreviousOperationID,
				OperationType:       a4cUpdateNodePropertyValueOperationJavaClassName,
			},
		},
		PropertyName:  propertyName,
		PropertyValue: propertyValue,
	}

	if a4cCtx.TopologyID == "" {
		var err error
		a4cCtx.TopologyID, err = t.GetTopologyID(a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s\n", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}
	err := t.editTopology(nil, a4cCtx, topoEditorExecute)

	if err != nil {
		return errors.Wrapf(err, "UpdateComponentProperty : Unable to edit the topology of application '%s' and environment '%s'\n", a4cCtx.AppID, a4cCtx.EnvID)
	}

	return nil
}

// UpdateCapabilityProperty Update the property value of a capability related to a component of an application
func (t *topologyService) UpdateCapabilityProperty(a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue string, capabilityName string) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	topoEditorExecute := TopologyEditorUpdateCapabilityProperty{
		TopologyEditorExecuteNodeRequest: TopologyEditorExecuteNodeRequest{
			NodeName: componentName,
			TopologyEditorExecuteRequest: TopologyEditorExecuteRequest{
				PreviousOperationID: a4cCtx.PreviousOperationID,
				OperationType:       a4cUpdateNodePropertyValueOperationJavaClassName,
			},
		},
		PropertyName:   propertyName,
		PropertyValue:  propertyValue,
		CapabilityName: capabilityName,
	}

	if a4cCtx.TopologyID == "" {
		var err error
		a4cCtx.TopologyID, err = t.GetTopologyID(a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}

	err := t.editTopology(nil, a4cCtx, topoEditorExecute)

	if err != nil {
		return errors.Wrapf(err, "Unable to edit the topology of application '%s' and environment '%s'", a4cCtx.AppID, a4cCtx.EnvID)
	}

	return nil
}

// AddNodeInA4CTopology Add a new node in the A4C topology
func (t *topologyService) AddNodeInA4CTopology(a4cCtx *TopologyEditorContext, NodeTypeID string, nodeName string) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	a4cTopology, err := t.getTopology(a4cCtx.AppID, a4cCtx.EnvID)

	if err != nil {
		return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
	}

	var nodeTypeVersion string

	for _, node := range a4cTopology.Data.NodeTypes {
		if NodeTypeID == node.ElementID {
			nodeTypeVersion = node.ArchiveVersion
		}
	}

	if reflect.DeepEqual(nodeTypeVersion, reflect.Zero(reflect.TypeOf(nodeTypeVersion)).Interface()) {
		return errors.Wrapf(err, "Unable to get archive version for node '%s' from A4C application topology for app %s and env %s", NodeTypeID, a4cCtx.AppID, a4cCtx.EnvID)
	}

	topoEditorExecute := TopologyEditorAddNode{
		TopologyEditorExecuteNodeRequest: TopologyEditorExecuteNodeRequest{
			NodeName: nodeName,
			TopologyEditorExecuteRequest: TopologyEditorExecuteRequest{
				PreviousOperationID: a4cCtx.PreviousOperationID,
				OperationType:       a4cUpdateNodePropertyValueOperationJavaClassName,
			},
		},
		NodeTypeID: NodeTypeID + ":" + nodeTypeVersion,
	}

	if a4cCtx.TopologyID == "" {
		a4cCtx.TopologyID, err = t.GetTopologyID(a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}

	err = t.editTopology(nil, a4cCtx, topoEditorExecute)

	if err != nil {
		return errors.Wrapf(err, "Unable to edit the topology of application '%s' and environment '%s'", a4cCtx.AppID, a4cCtx.EnvID)
	}

	return nil
}

// AddRelationship Add a new relationship in the A4C topology
func (t *topologyService) AddRelationship(a4cCtx *TopologyEditorContext, sourceNodeName string, targetNodeName string, relType string) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	var sourceNodeDef nodeType
	var targetNodeDef nodeType
	var requirementDef componentRequirement
	var relationshipDef relationshipType
	var capabilityDef componentCapability

	a4cTopology, err := t.getTopology(a4cCtx.AppID, a4cCtx.EnvID)

	if err != nil {
		return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
	}

	for _, node := range a4cTopology.Data.Topology.NodeTemplates {

		if sourceNodeName == node.Name {
			for _, nodeDef := range a4cTopology.Data.NodeTypes {
				if node.Type == nodeDef.ElementID {
					sourceNodeDef = nodeDef
					break
				}
			}
		}

		if targetNodeName == node.Name {
			for _, nodeDef := range a4cTopology.Data.NodeTypes {
				if node.Type == nodeDef.ElementID {
					targetNodeDef = nodeDef
					break
				}
			}
		}

	}

	if reflect.DeepEqual(sourceNodeDef, reflect.Zero(reflect.TypeOf(sourceNodeDef)).Interface()) {
		return errors.New("Missing relationship source node attribute")
	}

	if reflect.DeepEqual(targetNodeDef, reflect.Zero(reflect.TypeOf(targetNodeDef)).Interface()) {
		return errors.New("Missing relationship target node attribute")
	}

	for _, req := range sourceNodeDef.Requirements {
		if relType == req.RelationshipType {
			requirementDef = req
		}
	}

	if reflect.DeepEqual(requirementDef, reflect.Zero(reflect.TypeOf(requirementDef)).Interface()) {
		return errors.New("Missing relationship requirement attribute")
	}

	for _, rel := range a4cTopology.Data.RelationshipTypes {
		if relType == rel.ElementID {
			relationshipDef = rel
		}
	}

	if reflect.DeepEqual(relationshipDef, reflect.Zero(reflect.TypeOf(relationshipDef)).Interface()) {
		return errors.New("Missing relationship type")
	}

	for _, c := range targetNodeDef.Capabilities {
		if requirementDef.Type == c.Type {
			capabilityDef = c
		}
	}

	if reflect.DeepEqual(capabilityDef, reflect.Zero(reflect.TypeOf(capabilityDef)).Interface()) {
		return errors.New("Missing relationship capability type")
	}

	relTmp := strings.Split(relType, ".")
	relationshipName := sourceNodeName + strings.Title(relTmp[len(relTmp)-1]) + strings.Title(targetNodeName)

	topoEditorExecute := TopologyEditorAddRelationships{
		TopologyEditorExecuteNodeRequest: TopologyEditorExecuteNodeRequest{
			NodeName: sourceNodeName,
			TopologyEditorExecuteRequest: TopologyEditorExecuteRequest{
				PreviousOperationID: a4cCtx.PreviousOperationID,
				OperationType:       a4cUpdateNodePropertyValueOperationJavaClassName,
			},
		},
		RelationshipName:       relationshipName,
		RelationshipType:       relType,
		RelationshipVersion:    relationshipDef.ArchiveVersion,
		RequirementName:        requirementDef.ID,
		RequirementType:        requirementDef.Type,
		Target:                 targetNodeName,
		TargetedCapabilityName: capabilityDef.ID,
	}

	if a4cCtx.TopologyID == "" {
		a4cCtx.TopologyID, err = t.GetTopologyID(a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}

	err = t.editTopology(nil, a4cCtx, topoEditorExecute)

	if err != nil {
		return errors.Wrapf(err, "Unable to edit the topology of application '%s' and environment '%s'", a4cCtx.AppID, a4cCtx.EnvID)
	}

	return nil
}

// SaveA4CTopology saves the topology context
func (t *topologyService) SaveA4CTopology(a4cCtx *TopologyEditorContext) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	if a4cCtx.TopologyID == "" {
		var err error
		a4cCtx.TopologyID, err = t.GetTopologyID(a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}

	response, err := t.client.do(
		"POST",
		fmt.Sprintf("%s/editor/%s?lastOperationId=%s", a4CRestAPIPrefix, a4cCtx.TopologyID, a4cCtx.PreviousOperationID),
		nil,
		[]Header{
			{
				"Content-Type",
				"application/json",
			},
			{
				"Accept",
				"application/json",
			},
		},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send the request to save an A4C topology")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return getError(response.Body)
	}

	// After saving topology, get come back to a clear state.
	a4cCtx.PreviousOperationID = ""

	return nil
}
