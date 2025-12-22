package orchestrator

import (
	"fmt"
	"strings"
	"time"
)

// HandoffLevel represents the token budget for handoff context.
type HandoffLevel string

const (
	// HandoffMinimal is for simple tasks with predictable results (400 tokens)
	HandoffMinimal HandoffLevel = "minimal"
	// HandoffStandard is for typical development tasks (1000 tokens)
	HandoffStandard HandoffLevel = "standard"
	// HandoffExtended is for complex tasks with architectural changes (2500 tokens)
	HandoffExtended HandoffLevel = "extended"
)

// TokenLimit returns the token limit for the handoff level.
func (h HandoffLevel) TokenLimit() int {
	switch h {
	case HandoffMinimal:
		return 400
	case HandoffStandard:
		return 1000
	case HandoffExtended:
		return 2500
	default:
		return 1000
	}
}

// Handoff represents context passed between agents.
type Handoff struct {
	FromAgent     string
	ToAgent       string
	Context       string
	Level         HandoffLevel
	PriorityItems []string
	Timestamp     time.Time
}

// HandoffBuilder helps construct handoff context.
type HandoffBuilder struct {
	agentID       string
	taskCompleted bool
	summary       string
	artifacts     []ArtifactSummary
	decisions     []DecisionSummary
	issues        []IssueSummary
	nextAgent     string
	contextForNext string
	priorityItems []string
}

// ArtifactSummary is a compact artifact representation for handoff.
type ArtifactSummary struct {
	Path        string
	Action      string
	Description string
}

// DecisionSummary is a compact decision representation for handoff.
type DecisionSummary struct {
	Decision  string
	Rationale string
}

// IssueSummary is a compact issue representation for handoff.
type IssueSummary struct {
	Severity string
	Location string
	Message  string
}

// NewHandoffBuilder creates a new handoff builder.
func NewHandoffBuilder(agentID string) *HandoffBuilder {
	return &HandoffBuilder{
		agentID: agentID,
	}
}

// SetTaskCompleted sets whether the task was completed.
func (b *HandoffBuilder) SetTaskCompleted(completed bool) *HandoffBuilder {
	b.taskCompleted = completed
	return b
}

// SetSummary sets the task summary (max 500 tokens).
func (b *HandoffBuilder) SetSummary(summary string) *HandoffBuilder {
	b.summary = summary
	return b
}

// AddArtifact adds an artifact to the handoff.
func (b *HandoffBuilder) AddArtifact(path, action, description string) *HandoffBuilder {
	b.artifacts = append(b.artifacts, ArtifactSummary{
		Path:        path,
		Action:      action,
		Description: description,
	})
	return b
}

// AddDecision adds a decision to the handoff.
func (b *HandoffBuilder) AddDecision(decision, rationale string) *HandoffBuilder {
	b.decisions = append(b.decisions, DecisionSummary{
		Decision:  decision,
		Rationale: rationale,
	})
	return b
}

// AddIssue adds an issue to the handoff.
func (b *HandoffBuilder) AddIssue(severity, location, message string) *HandoffBuilder {
	b.issues = append(b.issues, IssueSummary{
		Severity: severity,
		Location: location,
		Message:  message,
	})
	return b
}

// SetNextAgent sets the next agent to handle the task.
func (b *HandoffBuilder) SetNextAgent(agentID string) *HandoffBuilder {
	b.nextAgent = agentID
	return b
}

// SetContextForNext sets context for the next agent.
func (b *HandoffBuilder) SetContextForNext(context string) *HandoffBuilder {
	b.contextForNext = context
	return b
}

// AddPriorityItem adds a priority item for the next agent.
func (b *HandoffBuilder) AddPriorityItem(item string) *HandoffBuilder {
	b.priorityItems = append(b.priorityItems, item)
	return b
}

// Build creates the handoff structure.
func (b *HandoffBuilder) Build(level HandoffLevel) *Handoff {
	return &Handoff{
		FromAgent:     b.agentID,
		ToAgent:       b.nextAgent,
		Context:       b.buildContext(level),
		Level:         level,
		PriorityItems: b.priorityItems,
		Timestamp:     time.Now(),
	}
}

// buildContext creates the context string within token limits.
func (b *HandoffBuilder) buildContext(level HandoffLevel) string {
	var sb strings.Builder

	// Always include summary
	sb.WriteString(fmt.Sprintf("## Summary\n%s\n\n", b.summary))

	// Include artifacts (paths only)
	if len(b.artifacts) > 0 {
		sb.WriteString("## Artifacts\n")
		for _, a := range b.artifacts {
			sb.WriteString(fmt.Sprintf("- %s [%s]: %s\n", a.Path, a.Action, a.Description))
		}
		sb.WriteString("\n")
	}

	// Include decisions for standard and extended levels
	if level != HandoffMinimal && len(b.decisions) > 0 {
		sb.WriteString("## Decisions\n")
		for _, d := range b.decisions {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", d.Decision, d.Rationale))
		}
		sb.WriteString("\n")
	}

	// Include issues
	if len(b.issues) > 0 {
		sb.WriteString("## Issues\n")
		for _, i := range b.issues {
			sb.WriteString(fmt.Sprintf("- [%s] %s: %s\n", i.Severity, i.Location, i.Message))
		}
		sb.WriteString("\n")
	}

	// Include context for next agent
	if b.contextForNext != "" {
		sb.WriteString(fmt.Sprintf("## Context for Next Agent\n%s\n\n", b.contextForNext))
	}

	// Include priority items
	if len(b.priorityItems) > 0 {
		sb.WriteString("## Priority Items\n")
		for _, p := range b.priorityItems {
			sb.WriteString(fmt.Sprintf("- %s\n", p))
		}
	}

	return sb.String()
}

// ToYAML converts the handoff to YAML format.
func (b *HandoffBuilder) ToYAML() string {
	var sb strings.Builder

	sb.WriteString("agent_output:\n")
	sb.WriteString(fmt.Sprintf("  agent_id: \"%s\"\n", b.agentID))
	sb.WriteString(fmt.Sprintf("  task_completed: %t\n", b.taskCompleted))
	sb.WriteString(fmt.Sprintf("\n  summary: |\n    %s\n", strings.ReplaceAll(b.summary, "\n", "\n    ")))

	if len(b.artifacts) > 0 {
		sb.WriteString("\n  artifacts:\n")
		for _, a := range b.artifacts {
			sb.WriteString(fmt.Sprintf("    - path: \"%s\"\n", a.Path))
			sb.WriteString(fmt.Sprintf("      action: \"%s\"\n", a.Action))
			sb.WriteString(fmt.Sprintf("      description: \"%s\"\n", a.Description))
		}
	}

	if len(b.decisions) > 0 {
		sb.WriteString("\n  decisions:\n")
		for _, d := range b.decisions {
			sb.WriteString(fmt.Sprintf("    - decision: \"%s\"\n", d.Decision))
			sb.WriteString(fmt.Sprintf("      rationale: \"%s\"\n", d.Rationale))
		}
	}

	if len(b.issues) > 0 {
		sb.WriteString("\n  issues:\n")
		for _, i := range b.issues {
			sb.WriteString(fmt.Sprintf("    - severity: \"%s\"\n", i.Severity))
			sb.WriteString(fmt.Sprintf("      location: \"%s\"\n", i.Location))
			sb.WriteString(fmt.Sprintf("      message: \"%s\"\n", i.Message))
		}
	}

	sb.WriteString("\n  handoff:\n")
	sb.WriteString(fmt.Sprintf("    next_agent: \"%s\"\n", b.nextAgent))
	if b.contextForNext != "" {
		sb.WriteString(fmt.Sprintf("    context_for_next: |\n      %s\n", strings.ReplaceAll(b.contextForNext, "\n", "\n      ")))
	}
	if len(b.priorityItems) > 0 {
		sb.WriteString("    priority_items:\n")
		for _, p := range b.priorityItems {
			sb.WriteString(fmt.Sprintf("      - \"%s\"\n", p))
		}
	}

	return sb.String()
}
