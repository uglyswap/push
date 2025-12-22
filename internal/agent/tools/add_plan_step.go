package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/uglyswap/crush/internal/planmode"
)

// AddPlanStepTool is the tool for adding steps to the current plan.
type AddPlanStepTool struct {
	planManager *planmode.PlanModeManager
}

// NewAddPlanStepTool creates a new AddPlanStep tool.
func NewAddPlanStepTool(pm *planmode.PlanModeManager) *AddPlanStepTool {
	return &AddPlanStepTool{
		planManager: pm,
	}
}

// Name returns the tool name.
func (t *AddPlanStepTool) Name() string {
	return "AddPlanStep"
}

// Description returns the tool description.
func (t *AddPlanStepTool) Description() string {
	return `Add a step to the current implementation plan. Use this while in plan mode to build out your implementation strategy.

Each step should be:
- Actionable and specific
- Small enough to complete in a reasonable time
- Clear about what files will be affected
- Properly sequenced with dependencies`
}

// AddPlanStepParams represents the parameters for AddPlanStep.
type AddPlanStepParams struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Files       []string `json:"files,omitempty"`
	DependsOn   []string `json:"depends_on,omitempty"`
}

// Parameters returns the JSON schema for the tool parameters.
func (t *AddPlanStepTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Short title for the step",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Detailed description of what this step involves",
			},
			"files": map[string]interface{}{
				"type":        "array",
				"description": "Files that will be created or modified in this step",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"depends_on": map[string]interface{}{
				"type":        "array",
				"description": "IDs of steps that must be completed before this one",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required":             []string{"title"},
		"additionalProperties": false,
	}
}

// Execute runs the AddPlanStep tool.
func (t *AddPlanStepTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
	var p AddPlanStepParams
	if err := json.Unmarshal(params, &p); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if p.Title == "" {
		return "", fmt.Errorf("title is required")
	}

	// Check if in plan mode
	if !t.planManager.IsActive() {
		return "Not in plan mode. Use EnterPlanMode first to start planning.", nil
	}

	// Add the step
	step := planmode.PlanStep{
		Title:       p.Title,
		Description: p.Description,
		Files:       p.Files,
		DependsOn:   p.DependsOn,
		Completed:   false,
	}

	if err := t.planManager.AddStep(step); err != nil {
		return "", fmt.Errorf("failed to add step: %w", err)
	}

	plan := t.planManager.GetCurrentPlan()
	stepNum := len(plan.Steps)

	return fmt.Sprintf("Added step %d: %s\n\nCurrent plan has %d steps.", stepNum, p.Title, stepNum), nil
}

// RequiresApproval returns whether this tool requires user approval.
func (t *AddPlanStepTool) RequiresApproval() bool {
	return false
}
