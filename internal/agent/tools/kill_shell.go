package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"syscall"
)

// ShellProcess represents a background shell process.
type ShellProcess struct {
	ID        string `json:"id"`
	Command   string `json:"command"`
	PID       int    `json:"pid"`
	StartedAt int64  `json:"started_at"`
	Status    string `json:"status"` // running, completed, killed
}

// ShellManager manages background shell processes.
type ShellManager struct {
	mu        sync.RWMutex
	processes map[string]*ShellProcess
	killFuncs map[string]func() error
}

// NewShellManager creates a new shell manager.
func NewShellManager() *ShellManager {
	return &ShellManager{
		processes: make(map[string]*ShellProcess),
		killFuncs: make(map[string]func() error),
	}
}

// Register registers a shell process.
func (m *ShellManager) Register(id, command string, pid int, killFunc func() error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.processes[id] = &ShellProcess{
		ID:      id,
		Command: command,
		PID:     pid,
		Status:  "running",
	}
	m.killFuncs[id] = killFunc
}

// Unregister removes a shell process.
func (m *ShellManager) Unregister(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.processes, id)
	delete(m.killFuncs, id)
}

// Get returns a shell process by ID.
func (m *ShellManager) Get(id string) *ShellProcess {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.processes[id]
}

// List returns all shell processes.
func (m *ShellManager) List() []*ShellProcess {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*ShellProcess, 0, len(m.processes))
	for _, p := range m.processes {
		result = append(result, p)
	}
	return result
}

// Kill kills a shell process.
func (m *ShellManager) Kill(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	killFunc, ok := m.killFuncs[id]
	if !ok {
		return fmt.Errorf("shell not found: %s", id)
	}

	if err := killFunc(); err != nil {
		return err
	}

	if proc, ok := m.processes[id]; ok {
		proc.Status = "killed"
	}

	return nil
}

// KillByPID kills a process by PID.
func (m *ShellManager) KillByPID(pid int) error {
	// Try to kill the process directly
	return syscall.Kill(pid, syscall.SIGTERM)
}

// KillShellTool kills background shell processes.
type KillShellTool struct {
	manager *ShellManager
}

// NewKillShellTool creates a new KillShell tool.
func NewKillShellTool(manager *ShellManager) *KillShellTool {
	return &KillShellTool{
		manager: manager,
	}
}

// Name returns the tool name.
func (t *KillShellTool) Name() string {
	return "KillShell"
}

// Description returns the tool description.
func (t *KillShellTool) Description() string {
	return `Kills a running background bash shell by its ID.

Usage:
- Takes a shell_id parameter identifying the shell to kill
- Returns a success or failure status
- Use this tool when you need to terminate a long-running shell
- Shell IDs can be found using the /tasks command`
}

// KillShellParams represents the parameters for KillShell.
type KillShellParams struct {
	ShellID string `json:"shell_id"`
}

// Parameters returns the JSON schema for the tool parameters.
func (t *KillShellTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"shell_id": map[string]interface{}{
				"type":        "string",
				"description": "The ID of the background shell to kill",
			},
		},
		"required": []string{"shell_id"},
	}
}

// Execute runs the KillShell tool.
func (t *KillShellTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
	var p KillShellParams
	if err := json.Unmarshal(params, &p); err != nil {
		return "", fmt.Errorf("failed to parse parameters: %w", err)
	}

	if p.ShellID == "" {
		return "", fmt.Errorf("shell_id is required")
	}

	// Get shell info before killing
	shell := t.manager.Get(p.ShellID)
	if shell == nil {
		return "", fmt.Errorf("shell not found: %s", p.ShellID)
	}

	// Kill the shell
	if err := t.manager.Kill(p.ShellID); err != nil {
		return "", fmt.Errorf("failed to kill shell: %w", err)
	}

	return fmt.Sprintf("Successfully killed shell %s (PID: %d, Command: %s)", p.ShellID, shell.PID, shell.Command), nil
}

// RequiresApproval returns whether this tool requires user approval.
func (t *KillShellTool) RequiresApproval() bool {
	return false
}
