package alien4cloud

import (
	"context"

	"github.com/pkg/errors"
)

// Adds a policy to the topology
func (t *topologyService) AddPolicy(ctx context.Context, a4cCtx *TopologyEditorContext, policyName, policyTypeID string) error {
	req := topologyEditorPolicies{
		topologyEditorExecuteRequest: topologyEditorExecuteRequest{
			OperationType: "org.alien4cloud.tosca.editor.operations.policies.AddPolicyOperation",
		},
		PolicyName:   policyName,
		PolicyTypeID: policyTypeID,
	}
	if a4cCtx.PreviousOperationID != "" {
		req.topologyEditorExecuteRequest.PreviousOperationID = &a4cCtx.PreviousOperationID
	}
	err := t.editTopology(ctx, a4cCtx, req)
	return errors.Wrapf(err, "Unable to add policy %q in topology of application %q and environment %q", policyName, a4cCtx.AppID, a4cCtx.EnvID)
}

// Deletes a policy from the topology
func (t *topologyService) DeletePolicy(ctx context.Context, a4cCtx *TopologyEditorContext, policyName string) error {
	req := topologyEditorPolicies{
		topologyEditorExecuteRequest: topologyEditorExecuteRequest{
			OperationType: "org.alien4cloud.tosca.editor.operations.policies.DeletePolicyOperation",
		},
		PolicyName: policyName,
	}
	if a4cCtx.PreviousOperationID != "" {
		req.topologyEditorExecuteRequest.PreviousOperationID = &a4cCtx.PreviousOperationID
	}
	err := t.editTopology(ctx, a4cCtx, req)
	return errors.Wrapf(err, "Unable to delete policy %q in topology of application %q and environment %q", policyName, a4cCtx.AppID, a4cCtx.EnvID)
}
