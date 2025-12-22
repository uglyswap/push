package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// TodoStatus represents the status of a todo item.
type TodoStatus string

const (
	TodoStatusPending    TodoStatus = "pending"
	TodoStatusInProgress TodoStatus = "in_progress"
	TodoStatusCompleted  TodoStatus = "completed"
)

// TodoItem represents a single todo item.
type TodoItem struct {
	Content    string     `json:"content"`
	Status     TodoStatus `json:"status"`
	ActiveForm string     `json:"activeForm"`
	CreatedAt  time.Time  `json:"created_at,omitempty"`
	UpdatedAt  time.Time  `json:"updated_at,omitempty"`
}

// TodoManager manages the todo list state.
type TodoManager struct {
	mu       sync.RWMutex
	todos    []TodoItem
	onChange func([]TodoItem)
}

// NewTodoManager creates a new todo manager.
func NewTodoManager() *TodoManager {
	return &TodoManager{
		todos: []TodoItem{},
	}
}

// SetOnChange sets a callback for todo changes.
func (tm *TodoManager) SetOnChange(fn func([]TodoItem)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.onChange = fn
}

// GetTodos returns a copy of the current todos.
func (tm *TodoManager) GetTodos() []TodoItem {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := make([]TodoItem, len(tm.todos))
	copy(result, tm.todos)
	return result
}

// SetTodos replaces the entire todo list.
func (tm *TodoManager) SetTodos(todos []TodoItem) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	now := time.Now()
	for i := range todos {
		if todos[i].CreatedAt.IsZero() {
			todos[i].CreatedAt = now
		}
		todos[i].UpdatedAt = now
	}

	tm.todos = todos

	if tm.onChange != nil {
		tm.onChange(todos)
	}
}

// GetProgress returns progress statistics.
func (tm *TodoManager) GetProgress() (completed, inProgress, pending, total int) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	total = len(tm.todos)
	for _, t := range tm.todos {
		switch t.Status {
		case TodoStatusCompleted:
			completed++
		case TodoStatusInProgress:
			inProgress++
		case TodoStatusPending:
			pending++
		}
	}
	return
}

// TodoWriteTool allows the agent to manage a todo list.
type TodoWriteTool struct {
	manager *TodoManager
}

// NewTodoWriteTool creates a new TodoWrite tool.
func NewTodoWriteTool(manager *TodoManager) *TodoWriteTool {
	return &TodoWriteTool{
		manager: manager,
	}
}

// Name returns the tool name.
func (t *TodoWriteTool) Name() string {
	return "TodoWrite"
}

// Description returns the tool description.
func (t *TodoWriteTool) Description() string {
	return `Use this tool to create and manage a structured task list for your current coding session. This helps you track progress, organize complex tasks, and demonstrate thoroughness to the user.

## When to Use This Tool
Use this tool proactively in these scenarios:

1. **Complex multi-step tasks** - When a task requires 3 or more distinct steps
2. **Non-trivial tasks** - Tasks that require careful planning or multiple operations
3. **User explicitly requests todo list** - When the user directly asks for task tracking
4. **User provides multiple tasks** - When users provide a list of things to be done
5. **After receiving new instructions** - Immediately capture user requirements as todos
6. **When you start working on a task** - Mark it as in_progress BEFORE beginning work
7. **After completing a task** - Mark it as completed immediately

## When NOT to Use This Tool

Skip using this tool when:
1. There is only a single, straightforward task
2. The task is trivial and tracking it provides no organizational benefit
3. The task can be completed in less than 3 trivial steps

## Task States
- **pending**: Task not yet started
- **in_progress**: Currently working on (limit to ONE task at a time)
- **completed**: Task finished successfully

## Task Descriptions
Each task needs two forms:
- **content**: Imperative form (e.g., "Run tests", "Build the project")
- **activeForm**: Present continuous form (e.g., "Running tests", "Building the project")`
}

// TodoWriteParams represents the parameters for TodoWrite.
type TodoWriteParams struct {
	Todos []TodoItem `json:"todos"`
}

// Parameters returns the JSON schema for the tool parameters.
func (t *TodoWriteTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"todos": map[string]interface{}{
				"type":        "array",
				"description": "The updated todo list",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"content": map[string]interface{}{
							"type":      "string",
							"minLength": 1,
						},
						"status": map[string]interface{}{
							"type": "string",
							"enum": []string{"pending", "in_progress", "completed"},
						},
						"activeForm": map[string]interface{}{
							"type":      "string",
							"minLength": 1,
						},
					},
					"required": []string{"content", "status", "activeForm"},
				},
			},
		},
		"required": []string{"todos"},
	}
}

// Execute runs the TodoWrite tool.
func (t *TodoWriteTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
	var p TodoWriteParams
	if err := json.Unmarshal(params, &p); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Validate todos
	inProgressCount := 0
	for i, todo := range p.Todos {
		if todo.Content == "" {
			return "", fmt.Errorf("todo %d: content is required", i+1)
		}
		if todo.ActiveForm == "" {
			return "", fmt.Errorf("todo %d: activeForm is required", i+1)
		}
		if todo.Status != TodoStatusPending && todo.Status != TodoStatusInProgress && todo.Status != TodoStatusCompleted {
			return "", fmt.Errorf("todo %d: invalid status '%s'", i+1, todo.Status)
		}
		if todo.Status == TodoStatusInProgress {
			inProgressCount++
		}
	}

	// Warn if multiple tasks are in progress
	if inProgressCount > 1 {
		// We allow it but log a warning
		fmt.Printf("Warning: %d tasks are in_progress. Best practice is to have only one.\n", inProgressCount)
	}

	// Update the todo list
	t.manager.SetTodos(p.Todos)

	// Generate response
	completed, inProgress, pending, total := t.manager.GetProgress()

	var result strings.Builder
	result.WriteString("Todos have been modified successfully.\n\n")
	result.WriteString(fmt.Sprintf("**Progress**: %d/%d completed", completed, total))
	if inProgress > 0 {
		result.WriteString(fmt.Sprintf(", %d in progress", inProgress))
	}
	if pending > 0 {
		result.WriteString(fmt.Sprintf(", %d pending", pending))
	}
	result.WriteString("\n\n")

	// Show current todos
	result.WriteString("**Current Tasks:**\n")
	for i, todo := range p.Todos {
		var status string
		switch todo.Status {
		case TodoStatusCompleted:
			status = "✅"
		case TodoStatusInProgress:
			status = "⏳"
		case TodoStatusPending:
			status = "⬜"
		}
		result.WriteString(fmt.Sprintf("%d. %s %s\n", i+1, status, todo.Content))
	}

	if inProgress > 0 {
		result.WriteString("\nPlease proceed with the current tasks if applicable.")
	}

	return result.String(), nil
}

// RequiresApproval returns whether this tool requires user approval.
func (t *TodoWriteTool) RequiresApproval() bool {
	return false
}
