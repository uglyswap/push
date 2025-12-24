package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/uglyswap/push/internal/planmode"
)

// ExitPlanModeTool is the tool for exiting plan mode.
type ExitPlanModeTool struct {
	planManager *planmode.PlanModeManager
}

// NewExitPlanModeTool creates a new ExitPlanMode tool.
func NewExitPlanModeTool(pm *planmode.PlanModeManager) *ExitPlanModeTool {
	return &ExitPlanModeTool{
		planManager: pm,
	}
}

// Name returns the tool name.
func (t *ExitPlanModeTool) Name() string {
	return "ExitPlanMode"
}

// Description returns the tool description.
func (t *ExitPlanModeTool) Description() string {
	return `Use this tool when you are in plan mode and have finished writing your plan and are ready for user approval.

## How This Tool Works
- You should have already designed your implementation plan during plan mode
- This tool signals that you're done planning and ready for the user to review and approve
- The user will see the contents of your plan when they review it

## When to Use This Tool
IMPORTANT: Only use this tool when the task requires planning the implementation steps of a task that requires writing code. For research tasks where you're gathering information, searching files, reading files or in general trying to understand the codebase - do NOT use this tool.

## Handling Ambiguity in Plans
Before using this tool, ensure your plan is clear and unambiguous. If there are multiple valid approaches or unclear requirements:
1. Use the AskUserQuestion tool to clarify with the user
2. Ask about specific implementation choices
3. Clarify any assumptions that could affect the implementation
4. Only proceed with ExitPlanMode after resolving ambiguities`
}

// ExitPlanModeParams represents the parameters for ExitPlanMode.
type ExitPlanModeParams struct {
	Scope         string   `json:"scope,omitempty"`
	Approach      string   `json:"approach,omitempty"`
	FilesAffected []string `json:"files_affected,omitempty"`
}

// Parameters returns the JSON schema for the tool parameters.
func (t *ExitPlanModeTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"scope": map[string]interface{}{
				"type":        "string",
				"description": "Summary of what will be implemented",
			},
			"approach": map[string]interface{}{
				"type":        "string",
				"description": "High-level implementation approach",
			},
			"files_affected": map[string]interface{}{
				"type":        "array",
				"description": "List of files that will be created or modified",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"additionalProperties": false,
	}
}

// Execute runs the ExitPlanMode tool.
func (t *ExitPlanModeTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
	var p ExitPlanModeParams
	if len(params) > 0 {
		if err := json.Unmarshal(params, &p); err != nil {
			return "", fmt.Errorf("failed to parse parameters: %w", err)
		}
	}

	// Check if in plan mode
	if !t.planManager.IsActive() {
		return "Not currently in plan mode. Use EnterPlanMode first to start planning.", nil
	}

	// Update plan with final details if provided
	if p.Scope != "" {
		if err := t.planManager.SetScope(p.Scope); err != nil {
			return "", fmt.Errorf("failed to set scope: %w", err)
		}
	}

	if p.Approach != "" {
		if err := t.planManager.SetApproach(p.Approach); err != nil {
			return "", fmt.Errorf("failed to set approach: %w", err)
		}
	}

	if len(p.FilesAffected) > 0 {
		if err := t.planManager.SetFilesAffected(p.FilesAffected); err != nil {
			return "", fmt.Errorf("failed to set files affected: %w", err)
		}
	}

	// Submit for approval
	if err := t.planManager.SubmitForApproval(); err != nil {
		return "", fmt.Errorf("failed to submit plan for approval: %w", err)
	}

	// Get plan summary for display
	summary := t.planManager.GetPlanSummary()

	// Get YAML representation
	yamlContent, err := t.planManager.ToYAML()
	if err != nil {
		yamlContent = "(unable to generate YAML)"
	}

	response := fmt.Sprintf(`# Plan Ready for Review

%s

---

## Plan Details (YAML)

%s

---

**Status:** Pending Approval

The plan has been submitted for your review. Please review the implementation approach and:
- Approve the plan to proceed with implementation
- Request modifications if changes are needed
- Reject the plan to start over with a different approach`,
		summary, "```yaml\n"+yamlContent+"\n```")

	return response, nil
}

// RequiresApproval returns whether this tool requires user approval.
func (t *ExitPlanModeTool) RequiresApproval() bool {
	return false // The plan itself will be shown for approval
}
