package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/uglyswap/push/internal/planmode"
)

// AddPlanDecisionTool is the tool for adding architectural decisions to the plan.
type AddPlanDecisionTool struct {
	planManager *planmode.PlanModeManager
}

// NewAddPlanDecisionTool creates a new AddPlanDecision tool.
func NewAddPlanDecisionTool(pm *planmode.PlanModeManager) *AddPlanDecisionTool {
	return &AddPlanDecisionTool{
		planManager: pm,
	}
}

// Name returns the tool name.
func (t *AddPlanDecisionTool) Name() string {
	return "AddPlanDecision"
}

// Description returns the tool description.
func (t *AddPlanDecisionTool) Description() string {
	return `Record an architectural decision made during planning. This creates a decision record (ADR) that documents:
- What decision was made
- Why this approach was chosen
- What alternatives were considered and rejected
- The expected impact

Use this for significant decisions that affect the codebase architecture.`
}

// AddPlanDecisionParams represents the parameters for AddPlanDecision.
type AddPlanDecisionParams struct {
	Decision             string   `json:"decision"`
	Rationale            string   `json:"rationale"`
	AlternativesRejected []string `json:"alternatives_rejected,omitempty"`
	Impact               string   `json:"impact,omitempty"`
}

// Parameters returns the JSON schema for the tool parameters.
func (t *AddPlanDecisionTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"decision": map[string]interface{}{
				"type":        "string",
				"description": "The architectural decision made",
			},
			"rationale": map[string]interface{}{
				"type":        "string",
				"description": "Why this decision was made",
			},
			"alternatives_rejected": map[string]interface{}{
				"type":        "array",
				"description": "Alternative approaches that were considered but rejected",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"impact": map[string]interface{}{
				"type":        "string",
				"description": "Expected impact of this decision",
			},
		},
		"required":             []string{"decision", "rationale"},
		"additionalProperties": false,
	}
}

// Execute runs the AddPlanDecision tool.
func (t *AddPlanDecisionTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
	var p AddPlanDecisionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if p.Decision == "" {
		return "", fmt.Errorf("decision is required")
	}
	if p.Rationale == "" {
		return "", fmt.Errorf("rationale is required")
	}

	// Check if in plan mode
	if !t.planManager.IsActive() {
		return "Not in plan mode. Use EnterPlanMode first to start planning.", nil
	}

	// Add the decision
	decision := planmode.ArchitecturalDecision{
		Decision:             p.Decision,
		Rationale:            p.Rationale,
		AlternativesRejected: p.AlternativesRejected,
		Impact:               p.Impact,
	}

	if err := t.planManager.AddDecision(decision); err != nil {
		return "", fmt.Errorf("failed to add decision: %w", err)
	}

	plan := t.planManager.GetCurrentPlan()
	decisionNum := len(plan.Decisions)

	return fmt.Sprintf("Recorded decision %d: %s\n\nRationale: %s", decisionNum, p.Decision, p.Rationale), nil
}

// RequiresApproval returns whether this tool requires user approval.
func (t *AddPlanDecisionTool) RequiresApproval() bool {
	return false
}
