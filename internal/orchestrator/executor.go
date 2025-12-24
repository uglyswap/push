package orchestrator

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/uglyswap/push/internal/config"
	"github.com/uglyswap/push/pkg/fantasy"
)

// ThinkingLevel represents the extended thinking budget for agent execution.
type ThinkingLevel string

const (
	ThinkingLevelNone        ThinkingLevel = ""
	ThinkingLevelThink       ThinkingLevel = "think"        // 1024 tokens
	ThinkingLevelThinkHard   ThinkingLevel = "think_hard"   // 4096 tokens
	ThinkingLevelThinkHarder ThinkingLevel = "think_harder" // 16384 tokens
	ThinkingLevelUltrathink  ThinkingLevel = "ultrathink"   // 32768 tokens
)

// ThinkingBudget returns the token budget for the thinking level.
func (t ThinkingLevel) ThinkingBudget() int64 {
	switch t {
	case ThinkingLevelThink:
		return 1024
	case ThinkingLevelThinkHard:
		return 4096
	case ThinkingLevelThinkHarder:
		return 16384
	case ThinkingLevelUltrathink:
		return 32768
	default:
		return 0
	}
}

// IsEnabled returns true if extended thinking is enabled.
func (t ThinkingLevel) IsEnabled() bool {
	return t != ThinkingLevelNone && t != ""
}

// ExecutorConfig holds configuration for the agent executor.
type ExecutorConfig struct {
	// ThinkingLevel controls extended thinking budget
	ThinkingLevel ThinkingLevel

	// MaxOutputTokens limits the response length
	MaxOutputTokens int64

	// Temperature controls response randomness (0.0-1.0)
	Temperature float64

	// Timeout for agent execution
	Timeout time.Duration
}

// DefaultExecutorConfig returns sensible defaults.
func DefaultExecutorConfig() ExecutorConfig {
	return ExecutorConfig{
		ThinkingLevel:   ThinkingLevelThinkHard, // Default to think_hard for quality
		MaxOutputTokens: 8192,
		Temperature:     0.7,
		Timeout:         5 * time.Minute,
	}
}

// Executor handles the actual LLM calls for subagents.
type Executor struct {
	config ExecutorConfig
}

// NewExecutor creates a new agent executor.
func NewExecutor(cfg ExecutorConfig) *Executor {
	return &Executor{
		config: cfg,
	}
}

// ExecuteAgent runs an agent with the given context and returns results.
func (e *Executor) ExecuteAgent(ctx context.Context, agent *Agent, agentCtx *AgentContext) (*AgentResult, error) {
	// Get the appropriate model based on agent configuration
	model, err := e.getModelForAgent(agent)
	if err != nil {
		return nil, fmt.Errorf("failed to get model for agent %s: %w", agent.ID, err)
	}

	// Build the system prompt
	systemPrompt := e.buildSystemPrompt(agent, agentCtx)

	// Build the user prompt
	userPrompt := e.buildUserPrompt(agent, agentCtx)

	// Create the fantasy agent
	fantasyAgent := fantasy.NewAgent(
		model,
		fantasy.WithSystemPrompt(systemPrompt),
	)

	// Prepare timeout context
	execCtx, cancel := context.WithTimeout(ctx, e.config.Timeout)
	defer cancel()

	// Build call options
	callOpts := fantasy.AgentStreamCall{
		Prompt:          userPrompt,
		MaxOutputTokens: &e.config.MaxOutputTokens,
	}

	if e.config.Temperature > 0 {
		callOpts.Temperature = &e.config.Temperature
	}

	// Execute the agent
	startTime := time.Now()
	result, err := fantasyAgent.Stream(execCtx, callOpts)
	executionTime := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("agent %s execution failed: %w", agent.ID, err)
	}

	// Parse the response into AgentResult
	agentResult := e.parseAgentResponse(agent, result, executionTime)

	return agentResult, nil
}

// getModelForAgent returns the appropriate fantasy model for the agent.
func (e *Executor) getModelForAgent(agent *Agent) (fantasy.LanguageModel, error) {
	cfg := config.Get()

	// Determine which model tier to use
	var modelType config.SelectedModelType
	switch agent.Model {
	case ModelOpus:
		modelType = config.SelectedModelTypeLarge
	case ModelHaiku:
		modelType = config.SelectedModelTypeSmall
	default: // ModelSonnet
		modelType = config.SelectedModelTypeLarge
	}

	selectedModel := cfg.Models[modelType]
	if selectedModel.Provider == "" || selectedModel.Model == "" {
		return nil, fmt.Errorf("no model configured for type %s", modelType)
	}

	// Get the provider configuration
	providers, err := config.Providers(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}

	// Find the matching provider and model
	for _, provider := range providers {
		if string(provider.ID) != selectedModel.Provider {
			continue
		}

		for _, model := range provider.Models {
			if model.ID == selectedModel.Model {
				// Create the fantasy model
				return e.createFantasyModel(provider, model)
			}
		}
	}

	return nil, fmt.Errorf("model %s not found for provider %s", selectedModel.Model, selectedModel.Provider)
}

// createFantasyModel creates a fantasy.LanguageModel from provider config.
func (e *Executor) createFantasyModel(provider interface{}, model interface{}) (fantasy.LanguageModel, error) {
	// This is a simplified implementation - in production, this would use
	// the actual provider SDK to create the model.
	// For now, we return a placeholder that indicates the model info.
	return nil, fmt.Errorf("model creation requires provider SDK integration - use main agent's model instead")
}

// buildSystemPrompt constructs the system prompt for the agent.
func (e *Executor) buildSystemPrompt(agent *Agent, agentCtx *AgentContext) string {
	var sb strings.Builder

	// Get squad-specific base prompt
	basePrompt := GetSquadPrompt(agent.Squad)
	sb.WriteString(basePrompt)
	sb.WriteString("\n\n")

	// Add agent-specific role
	sb.WriteString(fmt.Sprintf("# Agent: %s\n\n", agent.Name))
	sb.WriteString(fmt.Sprintf("## Role\n%s\n\n", agent.Description))

	// Add expertise
	if len(agent.Expertise) > 0 {
		sb.WriteString("## Expertise\n")
		for _, exp := range agent.Expertise {
			sb.WriteString(fmt.Sprintf("- %s\n", exp))
		}
		sb.WriteString("\n")
	}

	// Add trust level context
	if agentCtx.TrustLevel.Level > 0 {
		sb.WriteString(fmt.Sprintf("## Trust Level\nCurrent trust level: %s\n", agentCtx.TrustLevel.Name))
		if agentCtx.TrustLevel.Level <= TrustLevelSupervised {
			sb.WriteString("⚠️ Operating under supervision - all changes require verification.\n")
		}
		sb.WriteString("\n")
	}

	// Add thinking level hint if enabled
	if e.config.ThinkingLevel.IsEnabled() {
		sb.WriteString(fmt.Sprintf("## Thinking Mode\nExtended thinking enabled: %s (budget: %d tokens)\n",
			e.config.ThinkingLevel, e.config.ThinkingLevel.ThinkingBudget()))
		sb.WriteString("Take time to reason through complex problems thoroughly.\n\n")
	}

	// Add output format requirements
	sb.WriteString(OutputFormatPrompt)

	return sb.String()
}

// buildUserPrompt constructs the user prompt from task context.
func (e *Executor) buildUserPrompt(agent *Agent, agentCtx *AgentContext) string {
	var sb strings.Builder

	sb.WriteString("# Task\n\n")

	if agentCtx.Task != nil {
		sb.WriteString(fmt.Sprintf("## Description\n%s\n\n", agentCtx.Task.Description))

		if agentCtx.Task.Status != "" {
			sb.WriteString(fmt.Sprintf("## Status\n%s\n\n", agentCtx.Task.Status))
		}
	}

	// Add previous handoff context if available
	if agentCtx.PreviousHandoff != nil {
		sb.WriteString("## Context from Previous Agent\n")
		sb.WriteString(fmt.Sprintf("From: %s\n", agentCtx.PreviousHandoff.FromAgent))
		sb.WriteString(fmt.Sprintf("Context: %s\n\n", agentCtx.PreviousHandoff.Context))

		if len(agentCtx.PreviousHandoff.PriorityItems) > 0 {
			sb.WriteString("### Priority Items\n")
			for _, item := range agentCtx.PreviousHandoff.PriorityItems {
				sb.WriteString(fmt.Sprintf("- %s\n", item))
			}
			sb.WriteString("\n")
		}
	}

	// Add project context if available
	if len(agentCtx.ProjectContext) > 0 {
		sb.WriteString("## Project Context\n")
		for key, value := range agentCtx.ProjectContext {
			sb.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\nPlease complete this task and respond with the agent_output YAML format.")

	return sb.String()
}

// parseAgentResponse parses the LLM response into an AgentResult.
func (e *Executor) parseAgentResponse(agent *Agent, result *fantasy.AgentResult, executionTime time.Duration) *AgentResult {
	if result == nil || len(result.Steps) == 0 {
		return &AgentResult{
			AgentID:       agent.ID,
			TaskCompleted: false,
			Summary:       "No response from agent",
		}
	}

	// Get the response text
	responseText := result.Response.Content.Text()

	// Try to parse YAML agent_output format
	agentResult := e.parseYAMLOutput(agent.ID, responseText)

	// Add execution metrics
	agentResult.Metrics.ExecutionTime = executionTime.Milliseconds()
	agentResult.Metrics.TokensUsed = result.TotalUsage.InputTokens + result.TotalUsage.OutputTokens

	return agentResult
}

// parseYAMLOutput attempts to parse the agent_output YAML from the response.
func (e *Executor) parseYAMLOutput(agentID, text string) *AgentResult {
	result := &AgentResult{
		AgentID:       agentID,
		TaskCompleted: false,
		Summary:       "",
	}

	// Look for agent_output YAML block
	yamlPattern := regexp.MustCompile(`(?s)agent_output:\s*(.+?)(?:\n\n|\z|\x60\x60\x60)`)
	matches := yamlPattern.FindStringSubmatch(text)

	if len(matches) < 2 {
		// No YAML found, use full text as summary
		result.Summary = text
		if len(result.Summary) > 500 {
			result.Summary = result.Summary[:500] + "..."
		}
		return result
	}

	yamlContent := matches[1]

	// Parse key fields from YAML
	result.TaskCompleted = e.parseYAMLBool(yamlContent, "task_completed")
	result.Summary = e.parseYAMLString(yamlContent, "summary")
	result.NextAgent = e.parseYAMLString(yamlContent, "next_agent")
	result.HandoffContext = e.parseYAMLString(yamlContent, "context_for_next")

	// Parse artifacts
	result.Artifacts = e.parseYAMLArtifacts(yamlContent)

	// Parse issues
	result.Issues = e.parseYAMLIssues(yamlContent)

	// Parse decisions
	result.Decisions = e.parseYAMLDecisions(yamlContent)

	return result
}

// parseYAMLBool extracts a boolean value from YAML.
func (e *Executor) parseYAMLBool(yaml, key string) bool {
	pattern := regexp.MustCompile(fmt.Sprintf(`%s:\s*(true|false)`, key))
	matches := pattern.FindStringSubmatch(yaml)
	if len(matches) < 2 {
		return false
	}
	return matches[1] == "true"
}

// parseYAMLString extracts a string value from YAML.
func (e *Executor) parseYAMLString(yaml, key string) string {
	// Try quoted string first
	pattern := regexp.MustCompile(fmt.Sprintf(`%s:\s*["']([^"']*)["']`, key))
	matches := pattern.FindStringSubmatch(yaml)
	if len(matches) >= 2 {
		return matches[1]
	}

	// Try unquoted string
	pattern = regexp.MustCompile(fmt.Sprintf(`%s:\s*(.+?)(?:\n|$)`, key))
	matches = pattern.FindStringSubmatch(yaml)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

// parseYAMLArtifacts extracts artifacts from YAML.
func (e *Executor) parseYAMLArtifacts(yaml string) []Artifact {
	var artifacts []Artifact

	// Find artifacts section
	artifactsPattern := regexp.MustCompile(`(?s)artifacts:\s*(.+?)(?:decisions:|issues:|handoff:|$)`)
	matches := artifactsPattern.FindStringSubmatch(yaml)
	if len(matches) < 2 {
		return artifacts
	}

	// Parse individual artifacts
	itemPattern := regexp.MustCompile(`-\s*path:\s*["']?([^"'\n]+)["']?\s*\n\s*action:\s*([^\n]+)`)
	items := itemPattern.FindAllStringSubmatch(matches[1], -1)

	for _, item := range items {
		if len(item) >= 3 {
			action := strings.TrimSpace(item[2])
			artifacts = append(artifacts, Artifact{
				Path:      strings.TrimSpace(item[1]),
				Action:    ArtifactAction(action),
				CreatedAt: time.Now(),
			})
		}
	}

	return artifacts
}

// parseYAMLIssues extracts issues from YAML.
func (e *Executor) parseYAMLIssues(yaml string) []Issue {
	var issues []Issue

	// Find issues section
	issuesPattern := regexp.MustCompile(`(?s)issues:\s*(.+?)(?:handoff:|$)`)
	matches := issuesPattern.FindStringSubmatch(yaml)
	if len(matches) < 2 {
		return issues
	}

	// Parse individual issues
	itemPattern := regexp.MustCompile(`-\s*severity:\s*([^\n]+)\s*\n\s*message:\s*["']?([^"'\n]+)`)
	items := itemPattern.FindAllStringSubmatch(matches[1], -1)

	for _, item := range items {
		if len(item) >= 3 {
			severity := strings.TrimSpace(item[1])
			issues = append(issues, Issue{
				Severity: IssueSeverity(severity),
				Message:  strings.TrimSpace(item[2]),
			})
		}
	}

	return issues
}

// parseYAMLDecisions extracts decisions from YAML.
func (e *Executor) parseYAMLDecisions(yaml string) []Decision {
	var decisions []Decision

	// Find decisions section
	decisionsPattern := regexp.MustCompile(`(?s)decisions:\s*(.+?)(?:issues:|handoff:|$)`)
	matches := decisionsPattern.FindStringSubmatch(yaml)
	if len(matches) < 2 {
		return decisions
	}

	// Parse individual decisions
	itemPattern := regexp.MustCompile(`-\s*decision:\s*["']?([^"'\n]+)["']?\s*\n\s*rationale:\s*["']?([^"'\n]+)`)
	items := itemPattern.FindAllStringSubmatch(matches[1], -1)

	for _, item := range items {
		if len(item) >= 3 {
			decisions = append(decisions, Decision{
				Decision:  strings.TrimSpace(item[1]),
				Rationale: strings.TrimSpace(item[2]),
			})
		}
	}

	return decisions
}

// ParseThinkingLevel converts a string to ThinkingLevel.
func ParseThinkingLevel(s string) ThinkingLevel {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "think":
		return ThinkingLevelThink
	case "think_hard", "think-hard", "thinkhard":
		return ThinkingLevelThinkHard
	case "think_harder", "think-harder", "thinkharder":
		return ThinkingLevelThinkHarder
	case "ultrathink", "ultra_think", "ultra-think":
		return ThinkingLevelUltrathink
	default:
		return ThinkingLevelNone
	}
}

// GetThinkingLevelForTask returns appropriate thinking level based on task complexity.
func GetThinkingLevelForTask(task *Task, keywords []string) ThinkingLevel {
	if task == nil {
		return ThinkingLevelThinkHard
	}

	desc := strings.ToLower(task.Description)

	// Ultrathink triggers
	ultrathinkKeywords := []string{"architecture", "design system", "migration", "refactor major", "security audit"}
	for _, kw := range ultrathinkKeywords {
		if strings.Contains(desc, kw) {
			return ThinkingLevelUltrathink
		}
	}

	// Think harder triggers
	harderKeywords := []string{"complex", "optimize", "performance", "debug", "bug"}
	for _, kw := range harderKeywords {
		if strings.Contains(desc, kw) {
			return ThinkingLevelThinkHarder
		}
	}

	// Simple task triggers (use standard thinking)
	simpleKeywords := []string{"simple", "quick", "fix typo", "formatting", "update comment"}
	for _, kw := range simpleKeywords {
		if strings.Contains(desc, kw) {
			return ThinkingLevelThink
		}
	}

	// Default to think_hard
	return ThinkingLevelThinkHard
}


// parseYAMLInt extracts an integer value from YAML.
func (e *Executor) parseYAMLInt(yaml, key string) int {
	pattern := regexp.MustCompile(fmt.Sprintf(`%s:\s*(\d+)`, key))
	matches := pattern.FindStringSubmatch(yaml)
	if len(matches) < 2 {
		return 0
	}
	val, _ := strconv.Atoi(matches[1])
	return val
}
