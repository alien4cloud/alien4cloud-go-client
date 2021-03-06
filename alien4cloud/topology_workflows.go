package alien4cloud

import (
	"context"

	"github.com/pkg/errors"
)

const (
	// CallOperationWorkflowActivityType is the type of a call operation activity
	CallOperationWorkflowActivityType = "org.alien4cloud.tosca.model.workflow.activities.CallOperationWorkflowActivity"
	// InlineWorkflowActivityType is the type of an inline workflow activity
	InlineWorkflowActivityType = "org.alien4cloud.tosca.model.workflow.activities.InlineWorkflowActivity"
	// SetStateWorkflowActivityType is the type of an activity setting the state of a component
	SetStateWorkflowActivityType = "org.alien4cloud.tosca.model.workflow.activities.SetStateWorkflowActivity"
	// DelegateWorkflowActivity is the type of an activity delegated to an orchestrator
	DelegateWorkflowActivity = "org.alien4cloud.tosca.model.workflow.activities.DelegateWorkflowActivity"

	// StepStarted is the status of a workflow step that is started (currently running, not yet completed)
	StepStarted = "STARTED"
	// StepCompletedSuccessfull is the status of a workflow step that has completed successfully
	StepCompletedSuccessfull = "COMPLETED_SUCCESSFULL"
	// StepCompletedSuccessfull is the status of a workflow step that has failed
	StepCompletedWithError = "COMPLETED_WITH_ERROR"
)

// WorkflowActivity is a workflow activity payload.
//
// It allows to create:
// - Inline workflows activities (Inline should be set)
// - Operation Call activities (InterfaceName and OperationName should be set)
// - Set state activities (StateName should be set)
type WorkflowActivity struct {
	activitytype       string
	target             string
	targetRelationship string
	relatedStepID      string
	before             bool

	// For Inline Workflow activity
	inline string
	// For Operation Call activity
	interfaceName string
	operationName string
	// For Set State activity
	stateName string
}

// InsertBefore allows to insert the activity before the given step name in the workflow
func (wa *WorkflowActivity) InsertBefore(stepName string) *WorkflowActivity {
	wa.relatedStepID = stepName
	wa.before = true
	return wa
}

// AppendAfter allows to insert the activity after the given step name in the workflow
func (wa *WorkflowActivity) AppendAfter(stepName string) *WorkflowActivity {
	wa.relatedStepID = stepName
	return wa
}

// OperationCall allows to configure the workflow activity to be an operation call activity
// targetRelationship is optional and applies only on relationships-related operations
func (wa *WorkflowActivity) OperationCall(target, targetRelationship, interfaceName, operationName string) *WorkflowActivity {
	wa.activitytype = CallOperationWorkflowActivityType
	wa.target = target
	wa.targetRelationship = targetRelationship
	wa.interfaceName = interfaceName
	wa.operationName = operationName
	return wa
}

// InlineWorkflow allows to configure the workflow activity to be an inline workflow activity
func (wa *WorkflowActivity) InlineWorkflow(inlineWorkflow string) *WorkflowActivity {
	wa.activitytype = InlineWorkflowActivityType
	wa.inline = inlineWorkflow
	return wa
}

// SetState allows to configure the workflow activity to be an inline workflow call
func (wa *WorkflowActivity) SetState(target, stateName string) *WorkflowActivity {
	wa.activitytype = SetStateWorkflowActivityType
	wa.target = target
	wa.stateName = stateName
	return wa
}

// workflowActivityReq is a workflow activity payload.
//
// It allows to create:
// - Inline workflows activities (Inline should be set)
// - Operation Call activities (InterfaceName and OperationName should be set)
// - Set state activities (StateName should be set)
type workflowActivityReq struct {
	Type string `json:"type"`
	// For Inline Workflow activity
	Inline string `json:"inline,omitempty"`
	// For Operation Call activity
	InterfaceName string `json:"interfaceName,omitempty"`
	OperationName string `json:"operationName,omitempty"`
	// For Set State activity
	StateName string `json:"stateName,omitempty"`
}

type addWorkflowActivityReq struct {
	Type                string              `json:"type"`
	WorkflowName        string              `json:"workflowName"`
	Target              string              `json:"target,omitempty"`
	TargetRelationship  string              `json:"targetRelationship,omitempty"`
	RelatedStepID       string              `json:"relatedStepID,omitempty"`
	Before              *bool               `json:"before,omitempty"`
	Activity            workflowActivityReq `json:"activity"`
	PreviousOperationID *string             `json:"previousOperationId"`
}

func (r addWorkflowActivityReq) getPreviousOperationID() string {
	if r.PreviousOperationID == nil {
		return ""
	}
	return *r.PreviousOperationID
}
func (r addWorkflowActivityReq) getOperationType() string {
	return r.Type
}

// AddWorkflowActivity adds an activity to a workflow
func (t *topologyService) AddWorkflowActivity(ctx context.Context, a4cCtx *TopologyEditorContext, workflowName string, activity *WorkflowActivity) error {
	req := addWorkflowActivityReq{
		Type:               "org.alien4cloud.tosca.editor.operations.workflow.AddActivityOperation",
		WorkflowName:       workflowName,
		Target:             activity.target,
		TargetRelationship: activity.targetRelationship,
		Activity: workflowActivityReq{
			Type: activity.activitytype,
		},
	}

	if a4cCtx.PreviousOperationID != "" {
		req.PreviousOperationID = &a4cCtx.PreviousOperationID
	}

	if activity.relatedStepID != "" {
		req.RelatedStepID = activity.relatedStepID
		req.Before = &activity.before
	}

	switch activity.activitytype {
	case SetStateWorkflowActivityType:
		req.Activity.StateName = activity.stateName
	case InlineWorkflowActivityType:
		req.Activity.Inline = activity.inline
	case CallOperationWorkflowActivityType:
		req.Activity.InterfaceName = activity.interfaceName
		req.Activity.OperationName = activity.operationName
	default:
		return errors.Errorf("Unenexpected activity type %s", activity.activitytype)
	}
	err := t.editTopology(ctx, a4cCtx, req)
	return errors.Wrapf(err, "Unable to add activity to workflow %q in topology of application %q and environment %q", workflowName, a4cCtx.AppID, a4cCtx.EnvID)
}

// CreateWorkflow creates an empty workflow in the given topology
func (t *topologyService) CreateWorkflow(ctx context.Context, a4cCtx *TopologyEditorContext, workflowName string) error {
	return t.createOrDeleteWorkflow(ctx, a4cCtx, "org.alien4cloud.tosca.editor.operations.workflow.CreateWorkflowOperation", workflowName)
}

// AddRelationship Add a new relationship in the A4C topology
func (t *topologyService) DeleteWorkflow(ctx context.Context, a4cCtx *TopologyEditorContext, workflowName string) error {
	return t.createOrDeleteWorkflow(ctx, a4cCtx, "org.alien4cloud.tosca.editor.operations.workflow.RemoveWorkflowOperation", workflowName)
}

func (t *topologyService) createOrDeleteWorkflow(ctx context.Context, a4cCtx *TopologyEditorContext, operationName, workflowName string) error {
	var err error
	if a4cCtx == nil {
		return errors.New("Context object must be defined")
	}

	if a4cCtx.TopologyID == "" {
		a4cCtx.TopologyID, err = t.GetTopologyID(ctx, a4cCtx.AppID, a4cCtx.EnvID)
		if err != nil {
			return errors.Wrapf(err, "Unable to get A4C application topology for app %s and env %s", a4cCtx.AppID, a4cCtx.EnvID)
		}
	}

	topoEditorExecute := TopologyEditorWorkflow{
		TopologyEditorExecuteRequest: TopologyEditorExecuteRequest{
			PreviousOperationID: a4cCtx.PreviousOperationID,
			OperationType:       operationName,
		},
		WorkflowName: workflowName,
	}

	err = t.editTopology(ctx, a4cCtx, topoEditorExecute)

	if err != nil {
		return errors.Wrapf(err, "Unable to edit the topology of application '%s' and environment '%s'", a4cCtx.AppID, a4cCtx.EnvID)
	}

	return nil

}
