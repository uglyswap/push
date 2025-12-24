package orchestrator

import (
	"fmt"
	"strings"
)

// ScoringWeights defines the weights for each scoring dimension.
type ScoringWeights struct {
	Completeness     float64 // 30% - Everything requested is done
	Precision        float64 // 30% - Code is correct and works
	Coherence        float64 // 25% - Code integrates well with existing
	ContextRetention float64 // 15% - Project context is respected
}

// DefaultWeights returns the default scoring weights.
func DefaultWeights() ScoringWeights {
	return ScoringWeights{
		Completeness:     0.30,
		Precision:        0.30,
		Coherence:        0.25,
		ContextRetention: 0.15,
	}
}

// AgentScore represents the quality score for an agent's work.
type AgentScore struct {
	Completeness     float64
	Precision        float64
	Coherence        float64
	ContextRetention float64
	Weights          ScoringWeights
}

// Total calculates the weighted total score.
func (s AgentScore) Total() float64 {
	return s.Completeness*s.Weights.Completeness +
		s.Precision*s.Weights.Precision +
		s.Coherence*s.Weights.Coherence +
		s.ContextRetention*s.Weights.ContextRetention
}

// Grade returns a letter grade based on the total score.
func (s AgentScore) Grade() string {
	total := s.Total()
	switch {
	case total >= 0.90:
		return "A"
	case total >= 0.80:
		return "B"
	case total >= 0.70:
		return "C"
	case total >= 0.60:
		return "D"
	default:
		return "F"
	}
}

// Status returns the approval status based on score.
func (s AgentScore) Status() string {
	total := s.Total()
	switch {
	case total >= 0.75:
		return "auto-approved"
	case total >= 0.60:
		return "warning"
	default:
		return "blocked"
	}
}

// String returns a formatted score string.
func (s AgentScore) String() string {
	return fmt.Sprintf("%.0f%% (%s) [C:%.0f%% P:%.0f%% Co:%.0f%% R:%.0f%%]",
		s.Total()*100,
		s.Grade(),
		s.Completeness*100,
		s.Precision*100,
		s.Coherence*100,
		s.ContextRetention*100)
}

// ScoringEngine evaluates agent performance.
type ScoringEngine struct {
	weights ScoringWeights
}

// NewScoringEngine creates a new scoring engine with default weights.
func NewScoringEngine() *ScoringEngine {
	return &ScoringEngine{
		weights: DefaultWeights(),
	}
}

// NewScoringEngineWithWeights creates a scoring engine with custom weights.
func NewScoringEngineWithWeights(weights ScoringWeights) *ScoringEngine {
	return &ScoringEngine{
		weights: weights,
	}
}

// ScoreAgentForTask scores how well an agent matches a task.
func (e *ScoringEngine) ScoreAgentForTask(agent *Agent, taskDescription string) float64 {
	return agent.MatchesTask(taskDescription)
}

// ScoreResult evaluates an agent's result.
func (e *ScoringEngine) ScoreResult(result *AgentResult) AgentScore {
	score := AgentScore{
		Weights: e.weights,
	}

	// Score completeness based on task completion and artifacts
	if result.TaskCompleted {
		score.Completeness = 1.0
	} else {
		// Partial completion based on artifacts produced
		if len(result.Artifacts) > 0 {
			score.Completeness = 0.5
		} else {
			score.Completeness = 0.0
		}
	}

	// Score precision based on absence of blockers/critical issues
	score.Precision = 1.0
	for _, issue := range result.Issues {
		switch issue.Severity {
		case IssueSeverityBlocker:
			score.Precision -= 0.4
		case IssueSeverityCritical:
			score.Precision -= 0.25
		case IssueSeverityMajor:
			score.Precision -= 0.1
		case IssueSeverityMinor:
			score.Precision -= 0.05
		}
	}
	if score.Precision < 0 {
		score.Precision = 0
	}

	// Score coherence based on decisions and integration
	score.Coherence = 0.8 // Base score
	if len(result.Decisions) > 0 {
		// Having documented decisions improves coherence
		score.Coherence += 0.1
	}
	if result.HandoffContext != "" {
		// Good handoff context improves coherence
		score.Coherence += 0.1
	}
	if score.Coherence > 1.0 {
		score.Coherence = 1.0
	}

	// Score context retention based on summary quality
	if len(result.Summary) > 50 {
		score.ContextRetention = 0.9
	} else if len(result.Summary) > 20 {
		score.ContextRetention = 0.7
	} else {
		score.ContextRetention = 0.5
	}

	return score
}

// QualityCheck represents a single quality check.
type QualityCheck struct {
	Name        string
	Description string
	Passed      bool
	Message     string
}

// QualityChecklist runs a 10-point quality check.
func (e *ScoringEngine) QualityChecklist(result *AgentResult) []QualityCheck {
	checks := []QualityCheck{
		{
			Name:        "task_completed",
			Description: "Task was marked as completed",
			Passed:      result.TaskCompleted,
			Message:     boolToMessage(result.TaskCompleted, "Task completed", "Task incomplete"),
		},
		{
			Name:        "has_summary",
			Description: "Summary is provided",
			Passed:      len(result.Summary) > 0,
			Message:     boolToMessage(len(result.Summary) > 0, "Summary provided", "Missing summary"),
		},
		{
			Name:        "no_blockers",
			Description: "No blocker issues",
			Passed:      !hasBlockerIssues(result.Issues),
			Message:     boolToMessage(!hasBlockerIssues(result.Issues), "No blockers", "Has blocker issues"),
		},
		{
			Name:        "artifacts_documented",
			Description: "All artifacts are documented",
			Passed:      allArtifactsDocumented(result.Artifacts),
			Message:     boolToMessage(allArtifactsDocumented(result.Artifacts), "Artifacts documented", "Missing artifact descriptions"),
		},
		{
			Name:        "decisions_rationalized",
			Description: "Decisions have rationales",
			Passed:      allDecisionsRationalized(result.Decisions),
			Message:     boolToMessage(allDecisionsRationalized(result.Decisions), "Decisions explained", "Missing rationales"),
		},
		{
			Name:        "handoff_specified",
			Description: "Next agent or 'none' specified",
			Passed:      result.NextAgent != "",
			Message:     boolToMessage(result.NextAgent != "", "Handoff specified", "Missing handoff"),
		},
		{
			Name:        "issues_have_fixes",
			Description: "Issues have fix suggestions",
			Passed:      allIssuesHaveFixes(result.Issues),
			Message:     boolToMessage(allIssuesHaveFixes(result.Issues), "Issues have fixes", "Missing fix suggestions"),
		},
		{
			Name:        "code_blocks_valid",
			Description: "Code blocks have file paths",
			Passed:      allCodeBlocksHaveFiles(result.CodeBlocks),
			Message:     boolToMessage(allCodeBlocksHaveFiles(result.CodeBlocks), "Code blocks valid", "Missing file paths"),
		},
		{
			Name:        "metrics_provided",
			Description: "Metrics are provided",
			Passed:      result.Metrics.TokensUsed > 0,
			Message:     boolToMessage(result.Metrics.TokensUsed > 0, "Metrics provided", "Missing metrics"),
		},
		{
			Name:        "priority_items_set",
			Description: "Priority items for next agent",
			Passed:      len(result.PriorityItems) > 0 || result.NextAgent == "none",
			Message:     boolToMessage(len(result.PriorityItems) > 0 || result.NextAgent == "none", "Priority items set", "Missing priority items"),
		},
	}

	return checks
}

// Helper functions

func boolToMessage(condition bool, trueMsg, falseMsg string) string {
	if condition {
		return trueMsg
	}
	return falseMsg
}

func hasBlockerIssues(issues []Issue) bool {
	for _, issue := range issues {
		if issue.Severity == IssueSeverityBlocker {
			return true
		}
	}
	return false
}

func allArtifactsDocumented(artifacts []Artifact) bool {
	for _, a := range artifacts {
		if strings.TrimSpace(a.Description) == "" {
			return false
		}
	}
	return true
}

func allDecisionsRationalized(decisions []Decision) bool {
	for _, d := range decisions {
		if strings.TrimSpace(d.Rationale) == "" {
			return false
		}
	}
	return true
}

func allIssuesHaveFixes(issues []Issue) bool {
	for _, i := range issues {
		if strings.TrimSpace(i.FixSuggestion) == "" {
			return false
		}
	}
	return true
}

func allCodeBlocksHaveFiles(blocks []CodeBlock) bool {
	for _, b := range blocks {
		if strings.TrimSpace(b.File) == "" {
			return false
		}
	}
	return true
}
