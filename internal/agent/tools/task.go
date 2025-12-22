package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// SubagentType defines the type of specialized agent.
type SubagentType string

const (
	SubagentGeneral      SubagentType = "general-purpose"
	SubagentExplore      SubagentType = "Explore"
	SubagentPlan         SubagentType = "Plan"
	SubagentClaudeGuide  SubagentType = "claude-code-guide"
	SubagentStatusLine   SubagentType = "statusline-setup"
)

// AgentModel specifies the model to use for the agent.
type AgentModel string

const (
	ModelSonnet AgentModel = "sonnet"
	ModelOpus   AgentModel = "opus"
	ModelHaiku  AgentModel = "haiku"
)

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

// TaskInfo contains information about a running or completed task.
type TaskInfo struct {
	ID            string       `json:"id"`
	Description   string       `json:"description"`
	SubagentType  SubagentType `json:"subagent_type"`
	Model         AgentModel   `json:"model,omitempty"`
	Status        TaskStatus   `json:"status"`
	Result        string       `json:"result,omitempty"`
	Error         string       `json:"error,omitempty"`
	StartedAt     time.Time    `json:"started_at"`
	CompletedAt   *time.Time   `json:"completed_at,omitempty"`
	Background    bool         `json:"background"`
}

// TaskExecutor is the interface for executing subagent tasks.
type TaskExecutor interface {
	ExecuteTask(ctx context.Context, task *TaskInfo, prompt string) (string, error)
}

// TaskManager manages subagent tasks.
type TaskManager struct {
	mu       sync.RWMutex
	tasks    map[string]*TaskInfo
	executor TaskExecutor
	nextID   int
}

// NewTaskManager creates a new task manager.
func NewTaskManager(executor TaskExecutor) *TaskManager {
	return &TaskManager{
		tasks:    make(map[string]*TaskInfo),
		executor: executor,
	}
}

// GetTask returns a task by ID.
func (tm *TaskManager) GetTask(id string) *TaskInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.tasks[id]
}

// ListTasks returns all tasks.
func (tm *TaskManager) ListTasks() []*TaskInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := make([]*TaskInfo, 0, len(tm.tasks))
	for _, t := range tm.tasks {
		result = append(result, t)
	}
	return result
}

// TaskTool launches specialized subagents for complex tasks.
type TaskTool struct {
	manager *TaskManager
}

// NewTaskTool creates a new Task tool.
func NewTaskTool(manager *TaskManager) *TaskTool {
	return &TaskTool{
		manager: manager,
	}
}

// Name returns the tool name.
func (t *TaskTool) Name() string {
	return "Task"
}

// Description returns the tool description.
func (t *TaskTool) Description() string {
	return `Launch a new agent to handle complex, multi-step tasks autonomously.

The Task tool launches specialized agents (subprocesses) that autonomously handle complex tasks. Each agent type has specific capabilities and tools available to it.

Available agent types:
- **general-purpose**: General-purpose agent for researching complex questions, searching for code, and executing multi-step tasks
- **Explore**: Fast agent specialized for exploring codebases. Use for finding files by patterns, searching code for keywords, or answering questions about the codebase
- **Plan**: Software architect agent for designing implementation plans
- **claude-code-guide**: Use for questions about Claude Code features, hooks, slash commands, MCP servers, settings

When using the Task tool:
- Always include a short description (3-5 words) summarizing what the agent will do
- Launch multiple agents concurrently whenever possible to maximize performance
- When the agent is done, it will return a single message back to you
- You can run agents in the background using run_in_background parameter
- Agents can be resumed using the resume parameter by passing the agent ID
- Provide clear, detailed prompts so the agent can work autonomously`
}

// TaskParams represents the parameters for the Task tool.
type TaskParams struct {
	Description     string       `json:"description"`
	Prompt          string       `json:"prompt"`
	SubagentType    SubagentType `json:"subagent_type"`
	Model           AgentModel   `json:"model,omitempty"`
	Resume          string       `json:"resume,omitempty"`
	RunInBackground bool         `json:"run_in_background,omitempty"`
}

// Parameters returns the JSON schema for the tool parameters.
func (t *TaskTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"description": map[string]interface{}{
				"type":        "string",
				"description": "A short (3-5 word) description of the task",
			},
			"prompt": map[string]interface{}{
				"type":        "string",
				"description": "The task for the agent to perform",
			},
			"subagent_type": map[string]interface{}{
				"type":        "string",
				"description": "The type of specialized agent to use",
			},
			"model": map[string]interface{}{
				"type":        "string",
				"description": "Optional model to use (sonnet, opus, haiku)",
				"enum":        []string{"sonnet", "opus", "haiku"},
			},
			"resume": map[string]interface{}{
				"type":        "string",
				"description": "Optional agent ID to resume from",
			},
			"run_in_background": map[string]interface{}{
				"type":        "boolean",
				"description": "Run this agent in the background",
			},
		},
		"required": []string{"description", "prompt", "subagent_type"},
	}
}

// Execute runs the Task tool.
func (t *TaskTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
	var p TaskParams
	if err := json.Unmarshal(params, &p); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Validate parameters
	if p.Description == "" {
		return "", fmt.Errorf("description is required")
	}
	if p.Prompt == "" {
		return "", fmt.Errorf("prompt is required")
	}
	if p.SubagentType == "" {
		return "", fmt.Errorf("subagent_type is required")
	}

	// Check if resuming an existing task
	if p.Resume != "" {
		existing := t.manager.GetTask(p.Resume)
		if existing == nil {
			return "", fmt.Errorf("task not found: %s", p.Resume)
		}
		if existing.Status == TaskStatusRunning {
			return fmt.Sprintf("Task %s is still running. Use TaskOutput to check its status.", p.Resume), nil
		}
		if existing.Status == TaskStatusCompleted {
			return fmt.Sprintf("Task %s already completed.\n\nResult:\n%s", p.Resume, existing.Result), nil
		}
	}

	// Create new task
	t.manager.mu.Lock()
	t.manager.nextID++
	taskID := fmt.Sprintf("task_%d_%d", time.Now().Unix(), t.manager.nextID)
	
	task := &TaskInfo{
		ID:           taskID,
		Description:  p.Description,
		SubagentType: p.SubagentType,
		Model:        p.Model,
		Status:       TaskStatusPending,
		StartedAt:    time.Now(),
		Background:   p.RunInBackground,
	}
	t.manager.tasks[taskID] = task
	t.manager.mu.Unlock()

	// Execute the task
	if p.RunInBackground {
		// Run in background
		go func() {
			t.manager.mu.Lock()
			task.Status = TaskStatusRunning
			t.manager.mu.Unlock()

			result, err := t.manager.executor.ExecuteTask(context.Background(), task, p.Prompt)

			t.manager.mu.Lock()
			now := time.Now()
			task.CompletedAt = &now
			if err != nil {
				task.Status = TaskStatusFailed
				task.Error = err.Error()
			} else {
				task.Status = TaskStatusCompleted
				task.Result = result
			}
			t.manager.mu.Unlock()
		}()

		return fmt.Sprintf("Task started in background.\n\n**Task ID**: %s\n**Description**: %s\n**Agent Type**: %s\n\nUse TaskOutput with task_id=%q to check the result.",
			taskID, p.Description, p.SubagentType, taskID), nil
	}

	// Run synchronously
	t.manager.mu.Lock()
	task.Status = TaskStatusRunning
	t.manager.mu.Unlock()

	result, err := t.manager.executor.ExecuteTask(ctx, task, p.Prompt)

	t.manager.mu.Lock()
	now := time.Now()
	task.CompletedAt = &now
	if err != nil {
		task.Status = TaskStatusFailed
		task.Error = err.Error()
		t.manager.mu.Unlock()
		return "", fmt.Errorf("task failed: %w", err)
	}
	task.Status = TaskStatusCompleted
	task.Result = result
	t.manager.mu.Unlock()

	return fmt.Sprintf("**Task Completed**: %s\n**Agent ID**: %s\n\n---\n\n%s", p.Description, taskID, result), nil
}

// RequiresApproval returns whether this tool requires user approval.
func (t *TaskTool) RequiresApproval() bool {
	return false
}

// TaskOutputTool retrieves output from background tasks.
type TaskOutputTool struct {
	manager *TaskManager
}

// NewTaskOutputTool creates a new TaskOutput tool.
func NewTaskOutputTool(manager *TaskManager) *TaskOutputTool {
	return &TaskOutputTool{
		manager: manager,
	}
}

// Name returns the tool name.
func (t *TaskOutputTool) Name() string {
	return "TaskOutput"
}

// Description returns the tool description.
func (t *TaskOutputTool) Description() string {
	return `Retrieves output from a running or completed task (background shell, agent, or remote session).

Usage:
- Takes a task_id parameter identifying the task
- Returns the task output along with status information
- Use block=true (default) to wait for task completion
- Use block=false for non-blocking check of current status
- Task IDs can be found using the /tasks command`
}

// TaskOutputParams represents the parameters for TaskOutput.
type TaskOutputParams struct {
	TaskID  string `json:"task_id"`
	Block   bool   `json:"block"`
	Timeout int    `json:"timeout,omitempty"`
}

// Parameters returns the JSON schema for the tool parameters.
func (t *TaskOutputTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task_id": map[string]interface{}{
				"type":        "string",
				"description": "The task ID to get output from",
			},
			"block": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to wait for completion",
				"default":     true,
			},
			"timeout": map[string]interface{}{
				"type":        "number",
				"description": "Max wait time in ms",
				"default":     30000,
				"minimum":     0,
				"maximum":     600000,
			},
		},
		"required": []string{"task_id"},
	}
}

// Execute runs the TaskOutput tool.
func (t *TaskOutputTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
	var p TaskOutputParams
	if err := json.Unmarshal(params, &p); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if p.TaskID == "" {
		return "", fmt.Errorf("task_id is required")
	}

	// Default timeout
	if p.Timeout == 0 {
		p.Timeout = 30000
	}

	task := t.manager.GetTask(p.TaskID)
	if task == nil {
		return "", fmt.Errorf("task not found: %s", p.TaskID)
	}

	// If blocking and task is running, wait
	if p.Block && task.Status == TaskStatusRunning {
		timeout := time.Duration(p.Timeout) * time.Millisecond
		deadline := time.Now().Add(timeout)

		for time.Now().Before(deadline) {
			task = t.manager.GetTask(p.TaskID)
			if task.Status != TaskStatusRunning {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Format result
	var result string
	switch task.Status {
	case TaskStatusRunning:
		result = fmt.Sprintf("**Task Status**: Running\n**Task ID**: %s\n**Description**: %s\n**Started**: %s\n\nTask is still running. Use block=true to wait for completion.",
			task.ID, task.Description, task.StartedAt.Format(time.RFC3339))
	case TaskStatusCompleted:
		duration := ""
		if task.CompletedAt != nil {
			duration = fmt.Sprintf(" (took %s)", task.CompletedAt.Sub(task.StartedAt).Round(time.Second))
		}
		result = fmt.Sprintf("**Task Status**: Completed%s\n**Task ID**: %s\n**Description**: %s\n\n---\n\n%s",
			duration, task.ID, task.Description, task.Result)
	case TaskStatusFailed:
		result = fmt.Sprintf("**Task Status**: Failed\n**Task ID**: %s\n**Description**: %s\n**Error**: %s",
			task.ID, task.Description, task.Error)
	case TaskStatusPending:
		result = fmt.Sprintf("**Task Status**: Pending\n**Task ID**: %s\n**Description**: %s\n\nTask has not started yet.",
			task.ID, task.Description)
	}

	return result, nil
}

// RequiresApproval returns whether this tool requires user approval.
func (t *TaskOutputTool) RequiresApproval() bool {
	return false
}
