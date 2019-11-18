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
	"fmt"
	"sort"
	"time"
)

// TopologyEditorContext A4C topology editor context to store PreviousOperationID
type TopologyEditorContext struct {
	AppID               string
	EnvID               string
	TopologyID          string
	PreviousOperationID string
}

// Header is the representation of an http header
type Header struct {
	Key   string
	Value string
}

// Error is the representation of an A4C error
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// searchRequest is the representation of a request to search objects as tpologies, orchestrators in the A4C catalog
type searchRequest struct {
	Query string `json:"query"`
	From  string `json:"from"`
	Size  string `json:"size"`
}

// environmentsSearchRequest is the representation of a request to search environments of an application in the A4C catalog
// TODO (HJo): Misnamed type
type environmentsSearchRequest struct {
	From string `json:"from"`
	Size string `json:"size"`
}

// logsSearchRequest is the representation of a request to search logs of an application in the A4C catalog
type logsSearchRequest struct {
	From    int    `json:"from"`
	Size    int    `json:"size,omitempty"`
	Query   string `json:"query,omitempty"`
	Filters struct {
		LogFilter
		DeploymentID []string `json:"deploymentId,omitempty"`
	} `json:"filters"`
	SortConfiguration struct {
		Ascending bool   `json:"ascending"`
		SortBy    string `json:"sortBy"`
	} `json:"sortConfiguration"`
}

// nodeTemplate is the representation a node template
type nodeTemplate struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// nodeType is the representation a node type
type nodeType struct {
	ArchiveName    string                 `json:"archiveName"`
	ArchiveVersion string                 `json:"archiveVersion"`
	ElementID      string                 `json:"elementId"`
	Requirements   []componentRequirement `json:"requirements"`
	Capabilities   []componentCapability  `json:"capabilities"`
	Properties     []componentProperty    `json:"properties"`
}

// relationshipType is the representation a relationship type
type relationshipType struct {
	ArchiveName    string   `json:"archiveName"`
	ArchiveVersion string   `json:"archiveVersion"`
	ElementID      string   `json:"elementId"`
	DerivedFrom    []string `json:"deviredFrom"`
	ValidTargets   []string `json:"validTargets"`
	ID             string   `json:"id"`
}

// componentRequirement is the representation a component relationship requirement
type componentRequirement struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	RelationshipType string `json:"relationshipType"`
}

// capabilityType is the representation a component capability type
type capabilityType struct {
	ArchiveName    string   `json:"archiveName"`
	ArchiveVersion string   `json:"archiveVersion"`
	ElementID      string   `json:"elementId"`
	DerivedFrom    []string `json:"deviredFrom"`
	ID             string   `json:"id"`
}

// componentCapability is the representation a component capability
type componentCapability struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// componentProperty is the representation a component property
type componentProperty struct {
	Key   string `json:"key"`
	Value struct {
		Type     string `json:"type"`
		Required bool   `json:"required"`
	} `json:"value"`
}

// Location is the representation a location
type Location struct {
	ID   string
	Name string
}

// Deployment is the representation a deployment
type Deployment struct {
	DeploymentUsername       string      `json:"deploymentUsername"`
	EndDate                  Time        `json:"endDate"`
	EnvironmentID            string      `json:"environmentId"`
	ID                       string      `json:"id"`
	LocationIds              []string    `json:"locationIds"`
	OrchestratorDeploymentID string      `json:"orchestratorDeploymentId"`
	OrchestratorID           string      `json:"orchestratorId"`
	SourceID                 string      `json:"sourceId"`
	SourceName               string      `json:"sourceName"`
	SourceType               string      `json:"sourceType"`
	StartDate                Time        `json:"startDate"`
	VersionID                string      `json:"versionId"`
	WorkflowExecutions       interface{} `json:"workflowExecutions"`
}

// Topology is the representation a topology template
type Topology struct {
	Data struct {
		NodeTypes         map[string]nodeType         `json:"nodeTypes"`
		RelationshipTypes map[string]relationshipType `json:"relationshipTypes"`
		CapabilityTypes   map[string]capabilityType   `json:"capabilityTypes"`
		Topology          struct {
			ArchiveName    string                  `json:"archiveName"`
			ArchiveVersion string                  `json:"archiveVersion"`
			NodeTemplates  map[string]nodeTemplate `json:"nodeTemplates"`
		} `json:"topology"`
	} `json:"data"`
}

// ApplicationCreateRequest is the representation of a request to create an application from a topology template
type ApplicationCreateRequest struct {
	Name                      string `json:"name"`
	ArchiveName               string `json:"archiveName"`
	TopologyTemplateVersionID string `json:"topologyTemplateVersionId"`
}

// TopologyEditor is the representation a topology template editor
type TopologyEditor interface {
	getNodeName() string
	getPreviousOperationID() string
	getOperationType() string
}

// TopologyEditorExecuteRequest is the representation of a request to edit an application from a topology template
type TopologyEditorExecuteRequest struct {
	NodeName            string `json:"nodeName"`
	PreviousOperationID string `json:"previousOperationId,omitempty"`
	OperationType       string `json:"type"`
}

// getNodeName return the TopologyEditorExecuteRequest node name
func (r TopologyEditorExecuteRequest) getNodeName() string {
	return r.NodeName
}

// getPreviousOperationID return the TopologyEditorExecuteRequest previous operation ID
func (r TopologyEditorExecuteRequest) getPreviousOperationID() string {
	return r.PreviousOperationID
}

// getOperationType return the TopologyEditorExecuteRequest operation type
func (r TopologyEditorExecuteRequest) getOperationType() string {
	return r.OperationType
}

// TopologyEditorUpdateNodeProperty is the representation of a request to execute the topology editor
type TopologyEditorUpdateNodeProperty struct {
	TopologyEditorExecuteRequest
	PropertyName  string `json:"propertyName"`
	PropertyValue string `json:"propertyValue"`
	NodeTypeID    string `json:"indexedNodeTypeId"`
}

// TopologyEditorUpdateNodePropertyComplexType is the representation of a request to update complex property of a topology
type TopologyEditorUpdateNodePropertyComplexType struct {
	TopologyEditorExecuteRequest
	PropertyName  string                 `json:"propertyName"`
	PropertyValue map[string]interface{} `json:"propertyValue"`
	NodeTypeID    string                 `json:"indexedNodeTypeId"`
}

// TopologyEditorUpdateCapabilityProperty is the representation of a request to update property of a topology
type TopologyEditorUpdateCapabilityProperty struct {
	TopologyEditorExecuteRequest
	PropertyName   string `json:"propertyName"`
	PropertyValue  string `json:"propertyValue"`
	CapabilityName string `json:"capabilityName"`
}

// TopologyEditorAddNode is the representation of a request to set node of a topology
type TopologyEditorAddNode struct {
	TopologyEditorExecuteRequest
	NodeTypeID string `json:"indexedNodeTypeId"`
}

// TopologyEditorAddRelationships is the representation of a request to set relationships of a topology
type TopologyEditorAddRelationships struct {
	TopologyEditorExecuteRequest
	RelationshipName       string `json:"relationshipName"`
	RelationshipType       string `json:"relationshipType"`
	RelationshipVersion    string `json:"relationshipVersion"`
	RequirementName        string `json:"requirementName"`
	RequirementType        string `json:"requirementType"`
	Target                 string `json:"target"`
	TargetedCapabilityName string `json:"targetedCapabilityName"`
}

// LocationPoliciesPostRequestIn is the representation of a request to set location policies of a topology
type LocationPoliciesPostRequestIn struct {
	GroupsToLocations struct {
		A4CAll string `json:"_A4C_ALL"`
	} `json:"groupsToLocations"`
	OrchestratorID string `json:"orchestratorId"`
}

// ApplicationDeployRequest is the representation of a request to deploy an application in the A4C
type ApplicationDeployRequest struct {
	ApplicationEnvironmentID string `json:"applicationEnvironmentId"`
	ApplicationID            string `json:"applicationId"`
}

// Informations represents information returned from a4c rest api
type Informations struct {
	Data map[string]map[string]struct {
		State      string            `json:"state"`
		Attributes map[string]string `json:"attributes"`
	} `json:"data"`
	Error Error `json:"error"`
}

// RuntimeTopology represents runtime topology from a4c rest api
type RuntimeTopology struct {
	Data struct {
		Topology struct {
			OutputAttributes map[string][]string
		} `json:"topology"`
	} `json:"data"`
	Error Error `json:"error"`
}

// Log represents the log entry return by the a4c rest api
type Log struct {
	ID               string `json:"id"`
	DeploymentID     string `json:"deploymentId"`
	DeploymentPaaSID string `json:"deploymentPaaSId"`
	Level            string `json:"level"`
	Timestamp        Time   `json:"timestamp"`
	WorkflowID       string `json:"workflowId"`
	ExecutionID      string `json:"executionId"`
	NodeID           string `json:"nodeId"`
	InstanceID       string `json:"instanceId"`
	InterfaceName    string `json:"interfaceName"`
	OperationName    string `json:"operationName"`
	Content          string `json:"content"`
}

// Logs a list of a4c logs
type Logs []Log

// UnmarshalJSON unmarshals the a4c logs
func (l *Logs) UnmarshalJSON(b []byte) (err error) {

	logs := []Log{}

	if err := json.Unmarshal(b, &logs); err != nil {
		return err
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].ID < logs[j].ID
	})

	fmt.Printf("Logs here : %v\n", logs)

	a4cLogs := Logs(logs)

	*l = a4cLogs
	return
}

// LogFilter represents rest api A4C logs
type LogFilter struct {
	Level      []string `json:"level,omitempty"`
	WorkflowID []string `json:"workflowId,omitempty"`
}

// WorkflowExecution represents rest api workflow execution
type WorkflowExecution struct {
	DisplayWorkflowName string `json:"displayWorkflowName"`
	Status              string `json:"status"`
}

// Time represents the timestamp field from A4C
type Time struct {
	time.Time
}

// MarshalJSON marshals a4c json time data and return the result
func (t Time) MarshalJSON() ([]byte, error) {
	// 1 ms = 1 000 000 ns
	return json.Marshal(t.UnixNano() / int64(1000000))
}

// UnmarshalJSON unmarshal a4c json time data and sets the Time
func (t *Time) UnmarshalJSON(b []byte) (err error) {
	var parsedTime int64

	if err := json.Unmarshal(b, &parsedTime); err != nil {
		return err
	}

	// We try to Unmarshal data with nanoseconds precision.
	// Because timestamp from Alien4Cloud is Millisecond, we need to initialize the time
	// object with the number of seconds and the number of nano seconds
	t.Time = time.Unix(parsedTime/int64(1000), (parsedTime%int64(1000))*int64(1000000))
	return nil
}
