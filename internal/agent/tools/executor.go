// Package tools provides agent tools including task execution.
package tools

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/uglyswap/push/pkg/fantasy"
)

// LLMTaskExecutor implements TaskExecutor using LLM providers.
type LLMTaskExecutor struct {
	// LLM providers for different model tiers
	sonnetLLM fantasy.LanguageModel
	opusLLM   fantasy.LanguageModel
	haikuLLM  fantasy.LanguageModel

	// Logger
	logger *slog.Logger
}

// NewLLMTaskExecutor creates a new LLM-based task executor.
func NewLLMTaskExecutor() *LLMTaskExecutor {
	return &LLMTaskExecutor{
		logger: slog.Default(),
	}
}

// SetLLM sets the LLM provider for a specific model tier.
func (e *LLMTaskExecutor) SetLLM(model AgentModel, llm fantasy.LanguageModel) {
	switch model {
	case ModelSonnet:
		e.sonnetLLM = llm
	case ModelOpus:
		e.opusLLM = llm
	case ModelHaiku:
		e.haikuLLM = llm
	}
}

// SetDefaultLLM sets the same LLM for all model tiers.
func (e *LLMTaskExecutor) SetDefaultLLM(llm fantasy.LanguageModel) {
	e.sonnetLLM = llm
	e.opusLLM = llm
	e.haikuLLM = llm
}

// getLLM returns the appropriate LLM for the given model tier.
func (e *LLMTaskExecutor) getLLM(model AgentModel) fantasy.LanguageModel {
	switch model {
	case ModelOpus:
		if e.opusLLM != nil {
			return e.opusLLM
		}
	case ModelHaiku:
		if e.haikuLLM != nil {
			return e.haikuLLM
		}
	}
	// Default to sonnet
	return e.sonnetLLM
}

// ExecuteTask implements TaskExecutor.
func (e *LLMTaskExecutor) ExecuteTask(ctx context.Context, task *TaskInfo, prompt string) (string, error) {
	if task == nil {
		return "", fmt.Errorf("task is nil")
	}

	// Get the appropriate LLM
	llm := e.getLLM(task.Model)
	if llm == nil {
		return "", fmt.Errorf("no LLM configured for model tier: %s", task.Model)
	}

	// Build system prompt based on subagent type
	systemPrompt := e.buildSystemPrompt(task.SubagentType)

	// Build messages
	messages := []fantasy.Message{
		{
			Role:    fantasy.MessageRoleSystem,
			Content: []fantasy.MessagePart{fantasy.TextPart{Text: systemPrompt}},
		},
		{
			Role:    fantasy.MessageRoleUser,
			Content: []fantasy.MessagePart{fantasy.TextPart{Text: prompt}},
		},
	}

	// Configure generation options
	maxTokens := int64(8192)
	opts := fantasy.GenerateOptions{
		MaxOutputTokens: &maxTokens,
	}

	// Execute the LLM call
	e.logger.Info("Executing task",
		"task_id", task.ID,
		"subagent_type", task.SubagentType,
		"model", llm.Model(),
	)

	response, err := llm.Generate(ctx, messages, opts)
	if err != nil {
		e.logger.Error("Task execution failed",
			"task_id", task.ID,
			"error", err,
		)
		return "", fmt.Errorf("LLM execution failed: %w", err)
	}

	// Extract text from response
	result := response.Content.Text()

	e.logger.Info("Task completed",
		"task_id", task.ID,
		"tokens_used", response.Usage.TotalTokens,
	)

	return result, nil
}

// buildSystemPrompt creates a system prompt for the given subagent type.
func (e *LLMTaskExecutor) buildSystemPrompt(subagentType SubagentType) string {
	switch subagentType {
	case SubagentExplore:
		return `You are an expert codebase explorer. Your task is to explore and analyze codebases efficiently.

Your capabilities:
- Find files by patterns (e.g., "src/components/**/*.tsx")
- Search code for keywords and patterns
- Answer questions about codebase structure and architecture

When exploring:
1. Start with a high-level overview
2. Identify key directories and files
3. Look for patterns in naming and organization
4. Summarize your findings clearly

Be thorough but concise. Focus on what's most relevant to the user's question.`

	case SubagentPlan:
		return `You are a software architect specialized in designing implementation plans.

Your task is to create step-by-step implementation plans that are:
1. Clear and actionable
2. Technically sound
3. Consider edge cases and trade-offs
4. Identify potential risks and mitigation strategies

When planning:
- Break down complex tasks into manageable steps
- Identify critical files and components
- Consider architectural implications
- Suggest testing strategies
- Note any dependencies or prerequisites

Output your plan in a structured format with clear sections.`

	case SubagentClaudeGuide:
		return `You are an expert on Claude Code CLI, the Claude Agent SDK, and the Claude API.

Your task is to answer questions about:
- Claude Code features, hooks, slash commands, MCP servers, settings
- IDE integrations and keyboard shortcuts
- The Claude Agent SDK for building custom agents
- The Claude API and Anthropic SDK usage

Provide accurate, helpful information based on the official documentation.
Include code examples where relevant.
Be precise about feature availability and limitations.`

	case SubagentGeneral:
		return `You are a general-purpose AI agent capable of handling complex, multi-step tasks.

Your capabilities include:
- Researching complex questions
- Searching and analyzing code
- Executing multi-step workflows
- Synthesizing information from multiple sources

When working on a task:
1. Break it down into clear steps
2. Execute each step thoroughly
3. Verify your work as you go
4. Provide a comprehensive summary

Be thorough and accurate. If you're uncertain about something, say so.`

	default:
		return fmt.Sprintf(`You are a specialized AI agent of type: %s.

Execute the given task to the best of your ability.
Be thorough, accurate, and provide clear outputs.`, subagentType)
	}
}

// Verify LLMTaskExecutor implements TaskExecutor
var _ TaskExecutor = (*LLMTaskExecutor)(nil)
