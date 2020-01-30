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
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// TopologyService is the interface to the service mamaging topologies
type TopologyService interface {
	// Returns the topology ID on a given application and environment
	GetTopologyID(ctx context.Context, appID string, envID string) (string, error)
	// Returns the topology template ID for the given topologyName
	GetTopologyTemplateIDByName(ctx context.Context, topologyName string) (string, error)
	// Returns Topology details for a given application and environment
	GetTopology(ctx context.Context, appID string, envID string) (*Topology, error)
	// Updates the property value (type string) of a component of an application
	UpdateComponentProperty(ctx context.Context, a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue string) error
	// Updates the property value (type tosca complex) of a component of an application
	UpdateComponentPropertyComplexType(ctx context.Context, a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue map[string]interface{}) error
	// Updates the property value of a capability related to a component of an application
	UpdateCapabilityProperty(ctx context.Context, a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue string, capabilityName string) error
	// Adds a new node in the A4C topology
	AddNodeInA4CTopology(ctx context.Context, a4cCtx *TopologyEditorContext, nodeTypeID string, nodeName string) error
	// Adds a new relationship in the A4C topology
	AddRelationship(ctx context.Context, a4cCtx *TopologyEditorContext, sourceNodeName string, targetNodeName string, relType string) error
	// Saves the topology context
	SaveA4CTopology(ctx context.Context, a4cCtx *TopologyEditorContext) error
	// Creates an empty workflow in the given topology
	CreateWorkflow(ctx context.Context, a4cCtx *TopologyEditorContext, workflowName string) error
	// Deletes a workflow in the given topology
	DeleteWorkflow(ctx context.Context, a4cCtx *TopologyEditorContext, workflowName string) error
	// Adds an activity to a workflow
	AddWorkflowActivity(ctx context.Context, a4cCtx *TopologyEditorContext, workflowName string, activity *WorkflowActivity) error
	// Adds a policy to the topology
	AddPolicy(ctx context.Context, a4cCtx *TopologyEditorContext, policyName, policyTypeID string) error
	// Adds targets to a previously created policy
	AddTargetsToPolicy(ctx context.Context, a4cCtx *TopologyEditorContext, policyName string, targets []string) error
	// Deletes a policy from the topology
	DeletePolicy(ctx context.Context, a4cCtx *TopologyEditorContext, policyName string) error
	// Returns a list of available topologies
	GetTopologies(ctx context.Context, query string) ([]BasicTopologyInfo, error)
}

type topologyService struct {
	client restClient
}

const (
	// a4cUpdateNodePropertyValueOperationJavaClassName a4c class name to update node property value operation
	a4cUpdateNodePropertyValueOperationJavaClassName = "org.alien4cloud.tosca.editor.operations.nodetemplate.UpdateNodePropertyValueOperation"
)

// GetTopologyID returns the A4C topology ID on a given application and environment
func (t *topologyService) GetTopologyID(ctx context.Context, appID string, envID string) (string, error) {

	response, err := t.client.doWithContext(ctx,
		"GET",
		fmt.Sprintf("%s/applications/%s/environments/%s/topology", a4CRestAPIPrefix, appID, envID),
		nil,
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return "", errors.Wrapf(err, "Cannot send a request in order to find the topology for application '%s' in '%s' environment", appID, envID)
	}

	var res struct {
		Data string `json:"data"`
	}
	err = processA4CResponse(response, &res, http.StatusOK)
	return res.Data, errors.Wrapf(err, "Cannot read a4c response of the topology get data for application '%s' in '%s' environment", appID, envID)
}

// GetTopologyTemplateIDByName return the topology template ID for the given topologyName
func (t *topologyService) GetTopologyTemplateIDByName(ctx context.Context, topologyName string) (string, error) {

	toposSearchBody, err := json.Marshal(searchRequest{topologyName, "0", "1"})
	if err != nil {
		return "", errors.Wrap(err, "Cannot marshal an searchRequest structure")
	}

	response, err := t.client.doWithContext(ctx,
		"POST",
		fmt.Sprintf("%s/catalog/topologies/search", a4CRestAPIPrefix),
		[]byte(string(toposSearchBody)),
		[]Header{contentTypeAppJSONHeader},
	)
	if err != nil {
		return "", errors.Wrap(err, "Cannot send a request to get the topology id")
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
	err = processA4CResponse(response, &res, http.StatusOK)
	if err != nil {
		return "", errors.Wrapf(err, "Cannot read response when getting topology id for topology named %q", topologyName)
	}

	if res.Data.TotalResults <= 0 {
		return "", errors.Errorf("%q topology template does not exist", topologyName)
	}

	return res.Data.Data[0].ID, nil
}

// editTopology Edit the topology of an application
func (t *topologyService) editTopology(ctx context.Context, a4cCtx *TopologyEditorContext, a4cTopoEditorExecute TopologyEditor) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	if a4cCtx.TopologyID == "" {
		var err error
		a4cCtx.TopologyID, err = t.GetTopologyID(ctx, a4cCtx.AppID, a4cCtx.EnvID)
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
		[]Header{contentTypeAppJSONHeader, acceptAppJSONHeader},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send the request edit an A4C topology")
	}

	var resExec struct {
		Data struct {
			LastOperationIndex int `json:"lastOperationIndex"`
			Operations         []struct {
				PreviousOperationID string `json:"id"`
			} `json:"operations"`
		} `json:"data"`
	}

	err = processA4CResponse(response, &resExec, http.StatusOK)
	if err != nil {
		return errors.Wrap(err, "Unable to unmarshal a topology edition response")
	}

	lastOperationIndex := resExec.Data.LastOperationIndex
	if len(resExec.Data.Operations) > lastOperationIndex {
		a4cCtx.PreviousOperationID = resExec.Data.Operations[lastOperationIndex].PreviousOperationID
	}

	return nil
}

// GetTopology method returns topology details for a given application and environment
func (t *topologyService) GetTopology(ctx context.Context, appID string, envID string) (*Topology, error) {

	a4cTopologyID, err := t.GetTopologyID(ctx, appID, envID)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", appID, envID)
	}

	response, err := t.client.doWithContext(ctx,
		"GET",
		fmt.Sprintf("%s/topologies/%s", a4CRestAPIPrefix, a4cTopologyID),
		nil,
		[]Header{contentTypeAppJSONHeader},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get the topology content for application '%s' in '%s' environment", appID, envID)
	}

	res := new(Topology)
	err = processA4CResponse(response, res, http.StatusOK)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot convert the body of topology get data for application '%s' in '%s' environment", appID, envID)
	}

	return res, nil
}

// UpdateComponentPropertyComplexType Update the property value of a component of an application when propertyValue is not a simple type (map, array..)
func (t *topologyService) UpdateComponentPropertyComplexType(ctx context.Context, a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue map[string]interface{}) error {

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
		a4cCtx.TopologyID, err = t.GetTopologyID(ctx, a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s\n", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}
	err := t.editTopology(ctx, a4cCtx, topoEditorExecute)
	if err != nil {
		return errors.Wrapf(err, "UpdateComponentProperty : Unable to edit the topology of application '%s' and environment '%s'\n", a4cCtx.AppID, a4cCtx.EnvID)
	}

	return nil
}

// UpdateComponentProperty Update the property value of a component of an application
func (t *topologyService) UpdateComponentProperty(ctx context.Context, a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue string) error {

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
		a4cCtx.TopologyID, err = t.GetTopologyID(ctx, a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s\n", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}
	err := t.editTopology(ctx, a4cCtx, topoEditorExecute)

	if err != nil {
		return errors.Wrapf(err, "UpdateComponentProperty : Unable to edit the topology of application '%s' and environment '%s'\n", a4cCtx.AppID, a4cCtx.EnvID)
	}

	return nil
}

// UpdateCapabilityProperty Update the property value of a capability related to a component of an application
func (t *topologyService) UpdateCapabilityProperty(ctx context.Context, a4cCtx *TopologyEditorContext, componentName string, propertyName string, propertyValue string, capabilityName string) error {

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
		a4cCtx.TopologyID, err = t.GetTopologyID(ctx, a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}

	err := t.editTopology(ctx, a4cCtx, topoEditorExecute)

	if err != nil {
		return errors.Wrapf(err, "Unable to edit the topology of application '%s' and environment '%s'", a4cCtx.AppID, a4cCtx.EnvID)
	}

	return nil
}

// AddNodeInA4CTopology Add a new node in the A4C topology
func (t *topologyService) AddNodeInA4CTopology(ctx context.Context, a4cCtx *TopologyEditorContext, NodeTypeID string, nodeName string) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	a4cTopology, err := t.GetTopology(ctx, a4cCtx.AppID, a4cCtx.EnvID)

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
		a4cCtx.TopologyID, err = t.GetTopologyID(ctx, a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}

	err = t.editTopology(ctx, a4cCtx, topoEditorExecute)

	if err != nil {
		return errors.Wrapf(err, "Unable to edit the topology of application '%s' and environment '%s'", a4cCtx.AppID, a4cCtx.EnvID)
	}

	return nil
}

// AddRelationship Add a new relationship in the A4C topology
func (t *topologyService) AddRelationship(ctx context.Context, a4cCtx *TopologyEditorContext, sourceNodeName string, targetNodeName string, relType string) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	var sourceNodeDef nodeType
	var targetNodeDef nodeType
	var requirementDef componentRequirement
	var relationshipDef relationshipType
	var capabilityDef componentCapability

	a4cTopology, err := t.GetTopology(ctx, a4cCtx.AppID, a4cCtx.EnvID)

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
		a4cCtx.TopologyID, err = t.GetTopologyID(ctx, a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}

	err = t.editTopology(ctx, a4cCtx, topoEditorExecute)

	if err != nil {
		return errors.Wrapf(err, "Unable to edit the topology of application '%s' and environment '%s'", a4cCtx.AppID, a4cCtx.EnvID)
	}

	return nil
}

// SaveA4CTopology saves the topology context
func (t *topologyService) SaveA4CTopology(ctx context.Context, a4cCtx *TopologyEditorContext) error {

	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	if a4cCtx.TopologyID == "" {
		var err error
		a4cCtx.TopologyID, err = t.GetTopologyID(ctx, a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}

	response, err := t.client.doWithContext(ctx,
		"POST",
		fmt.Sprintf("%s/editor/%s?lastOperationId=%s", a4CRestAPIPrefix, a4cCtx.TopologyID, a4cCtx.PreviousOperationID),
		nil,
		[]Header{contentTypeAppJSONHeader, acceptAppJSONHeader},
	)

	if err != nil {
		return errors.Wrap(err, "Unable to send the request to save an A4C topology")
	}

	// After saving topology, get come back to a clear state.
	a4cCtx.PreviousOperationID = ""

	return processA4CResponse(response, nil, http.StatusOK)
}

func (t *topologyService) GetTopologies(ctx context.Context, query string) ([]BasicTopologyInfo, error) {

	getTopoJSON, err := json.Marshal(
		searchRequest{
			From:  "",
			Query: query,
			Size:  "",
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Cannot marshal an a4cgetTopologiesCreateRequest structure")
	}

	response, err := t.client.doWithContext(ctx,
		"POST",
		fmt.Sprintf("%s/catalog/topologies/search", a4CRestAPIPrefix),
		[]byte(string(getTopoJSON)),
		[]Header{contentTypeAppJSONHeader, acceptAppJSONHeader},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Cannot send request to list topologies")
	}

	var res struct {
		Data struct {
			Types []string `json:"types"`
			Data  []struct {
				ArchiveName string `json:"archiveName"`
				Workspace   string `json:"workspace"`
				ID          string `json:"id"`
			} `json:"data"`
		} `json:"data"`
	}

	err = processA4CResponse(response, &res, http.StatusOK)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot read response when getting topologies with query %q", query)
	}
	var topologyInfo []BasicTopologyInfo

	for i := range res.Data.Data {
		temp := BasicTopologyInfo{ArchiveName: res.Data.Data[i].ArchiveName, Workspace: res.Data.Data[i].Workspace, ID: res.Data.Data[i].ID}
		topologyInfo = append(topologyInfo, temp)
	}

	return topologyInfo, nil
}
