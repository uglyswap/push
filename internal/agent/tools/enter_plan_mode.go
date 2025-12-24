package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/uglyswap/push/internal/planmode"
)

// EnterPlanModeTool is the tool for entering plan mode.
type EnterPlanModeTool struct {
	planManager *planmode.PlanModeManager
}

// NewEnterPlanModeTool creates a new EnterPlanMode tool.
func NewEnterPlanModeTool(pm *planmode.PlanModeManager) *EnterPlanModeTool {
	return &EnterPlanModeTool{
		planManager: pm,
	}
}

// Name returns the tool name.
func (t *EnterPlanModeTool) Name() string {
	return "EnterPlanMode"
}

// Description returns the tool description.
func (t *EnterPlanModeTool) Description() string {
	return `Use this tool proactively when you're about to start a non-trivial implementation task. Getting user sign-off on your approach before writing code prevents wasted effort and ensures alignment. This tool transitions you into plan mode where you can explore the codebase and design an implementation approach for user approval.

## When to Use This Tool

**Prefer using EnterPlanMode** for implementation tasks unless they're simple. Use it when ANY of these conditions apply:

1. **New Feature Implementation**: Adding meaningful new functionality
2. **Multiple Valid Approaches**: The task can be solved in several different ways
3. **Code Modifications**: Changes that affect existing behavior or structure
4. **Architectural Decisions**: The task requires choosing between patterns or technologies
5. **Multi-File Changes**: The task will likely touch more than 2-3 files
6. **Unclear Requirements**: You need to explore before understanding the full scope
7. **User Preferences Matter**: The implementation could reasonably go multiple ways

## When NOT to Use This Tool

Only skip EnterPlanMode for simple tasks:
- Single-line or few-line fixes (typos, obvious bugs, small tweaks)
- Adding a single function with clear requirements
- Tasks where the user has given very specific, detailed instructions
- Pure research/exploration tasks

## What Happens in Plan Mode

In plan mode, you'll:
1. Thoroughly explore the codebase using Glob, Grep, and Read tools
2. Understand existing patterns and architecture
3. Design an implementation approach
4. Present your plan to the user for approval
5. Exit plan mode with ExitPlanMode when ready to implement`
}

// EnterPlanModeParams represents the parameters for EnterPlanMode.
type EnterPlanModeParams struct {
	Title     string `json:"title,omitempty"`
	Objective string `json:"objective,omitempty"`
}

// Parameters returns the JSON schema for the tool parameters.
func (t *EnterPlanModeTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Optional title for the plan",
			},
			"objective": map[string]interface{}{
				"type":        "string",
				"description": "Optional objective description",
			},
		},
		"additionalProperties": false,
	}
}

// Execute runs the EnterPlanMode tool.
func (t *EnterPlanModeTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
	var p EnterPlanModeParams
	if len(params) > 0 {
		if err := json.Unmarshal(params, &p); err != nil {
			return "", fmt.Errorf("failed to parse parameters: %w", err)
		}
	}

	// Use defaults if not provided
	if p.Title == "" {
		p.Title = "Implementation Plan"
	}
	if p.Objective == "" {
		p.Objective = "Plan the implementation approach"
	}

	// Check if already in plan mode
	if t.planManager.IsActive() {
		plan := t.planManager.GetCurrentPlan()
		return fmt.Sprintf("Already in plan mode with plan: %s\n\nCurrent plan status: %s\n\n%s",
			plan.Title, plan.Status, t.planManager.GetPlanSummary()), nil
	}

	// Enter plan mode
	plan, err := t.planManager.EnterPlanMode(ctx, p.Title, p.Objective)
	if err != nil {
		return "", fmt.Errorf("failed to enter plan mode: %w", err)
	}

	response := fmt.Sprintf(`# Entered Plan Mode

**Plan ID:** %s
**Title:** %s
**Objective:** %s

## Instructions

You are now in plan mode. In this mode:

1. **Explore the codebase** - Use Glob, Grep, and Read tools to understand the current implementation
2. **Identify patterns** - Note existing conventions and architectural decisions
3. **Design your approach** - Consider multiple solutions and their tradeoffs
4. **Document your plan** - Add steps, decisions, and risks to the plan
5. **Get approval** - Use ExitPlanMode when ready for user review

## Available Plan Actions

While in plan mode, you can:
- Add implementation steps
- Record architectural decisions
- Identify risks and mitigations
- Set the scope and approach
- List files that will be affected

## When Ready

Use the ExitPlanMode tool to finalize your plan and present it for user approval.

---

*Now begin exploring the codebase to understand the context for this implementation.*`,
		plan.ID, plan.Title, plan.Objective)

	return response, nil
}

// RequiresApproval returns whether this tool requires user approval.
func (t *EnterPlanModeTool) RequiresApproval() bool {
	return true // User must consent to entering plan mode
}
