package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/uglyswap/crush/internal/skills"
)

// SkillTool invokes skills from the registry.
type SkillTool struct {
	invoker  *skills.SkillInvoker
	registry *skills.SkillRegistry
}

// NewSkillTool creates a new Skill tool.
func NewSkillTool(invoker *skills.SkillInvoker, registry *skills.SkillRegistry) *SkillTool {
	return &SkillTool{
		invoker:  invoker,
		registry: registry,
	}
}

// Name returns the tool name.
func (t *SkillTool) Name() string {
	return "Skill"
}

// Description returns the tool description.
func (t *SkillTool) Description() string {
	return `Execute a skill within the main conversation.

Skills provide specialized capabilities and domain knowledge for specific tasks.

When users ask you to run a "slash command" or reference "/<something>" (e.g., "/commit", "/review-pr"), they are referring to a skill.

How to invoke:
- Use this tool with the skill name and optional arguments
- Examples:
  - skill: "pdf" - invoke the pdf skill
  - skill: "commit", args: "-m 'Fix bug'" - invoke with arguments
  - skill: "review-pr", args: "123" - invoke with arguments
  - skill: "namespace:skill" - invoke using fully qualified name

Important:
- When a skill is relevant, invoke this tool IMMEDIATELY as your first action
- NEVER just announce a skill without actually calling this tool
- Only use skills listed in available skills
- Do not invoke a skill that is already running`
}

// SkillParams represents the parameters for the Skill tool.
type SkillParams struct {
	Skill string `json:"skill"`
	Args  string `json:"args,omitempty"`
}

// Parameters returns the JSON schema for the tool parameters.
func (t *SkillTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"skill": map[string]interface{}{
				"type":        "string",
				"description": "The skill name. E.g., 'commit', 'review-pr', or 'pdf'",
			},
			"args": map[string]interface{}{
				"type":        "string",
				"description": "Optional arguments for the skill",
			},
		},
		"required": []string{"skill"},
	}
}

// Execute runs the Skill tool.
func (t *SkillTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
	var p SkillParams
	if err := json.Unmarshal(params, &p); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if p.Skill == "" {
		return "", fmt.Errorf("skill name is required")
	}

	// Invoke the skill
	result, err := t.invoker.Invoke(ctx, p.Skill, p.Args)
	if err != nil {
		return "", err
	}

	return result.Prompt, nil
}

// RequiresApproval returns whether this tool requires user approval.
func (t *SkillTool) RequiresApproval() bool {
	return false
}

// ListAvailableSkills returns a formatted list of available skills.
func (t *SkillTool) ListAvailableSkills() string {
	skillList := t.registry.List()

	if len(skillList) == 0 {
		return "No skills available."
	}

	var sb strings.Builder
	sb.WriteString("## Available Skills\n\n")

	// Group by source
	bySource := make(map[string][]*skills.Skill)
	for _, s := range skillList {
		bySource[s.Source] = append(bySource[s.Source], s)
	}

	for source, sourceSkills := range bySource {
		sb.WriteString(fmt.Sprintf("### %s\n\n", strings.Title(source)))
		for _, s := range sourceSkills {
			desc := s.Description
			if len(desc) > 80 {
				desc = desc[:80] + "..."
			}
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", s.Name, desc))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
