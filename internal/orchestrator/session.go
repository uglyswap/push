package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SessionSnapshot represents a point-in-time snapshot of files.
type SessionSnapshot struct {
	ID          string            `json:"id"`
	TaskID      string            `json:"task_id"`
	Label       string            `json:"label"`
	Files       map[string]string `json:"files"` // path -> content
	CreatedAt   time.Time         `json:"created_at"`
}

// SessionManager manages session state and snapshots.
type SessionManager struct {
	mu        sync.RWMutex
	snapshots map[string][]SessionSnapshot // taskID -> snapshots
	baseDir   string
}

// NewSessionManager creates a new session manager.
func NewSessionManager() *SessionManager {
	homeDir, _ := os.UserHomeDir()
	baseDir := filepath.Join(homeDir, ".crush", "sessions")
	os.MkdirAll(baseDir, 0755)

	return &SessionManager{
		snapshots: make(map[string][]SessionSnapshot),
		baseDir:   baseDir,
	}
}

// CreateSnapshot creates a snapshot of specified files.
func (sm *SessionManager) CreateSnapshot(ctx context.Context, taskID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	snapshot := SessionSnapshot{
		ID:        fmt.Sprintf("snap-%d", time.Now().UnixNano()),
		TaskID:    taskID,
		Label:     fmt.Sprintf("Pre-task snapshot"),
		Files:     make(map[string]string),
		CreatedAt: time.Now(),
	}

	sm.snapshots[taskID] = append(sm.snapshots[taskID], snapshot)

	// Persist snapshot
	return sm.persistSnapshot(snapshot)
}

// CreateSnapshotWithFiles creates a snapshot of specific files.
func (sm *SessionManager) CreateSnapshotWithFiles(ctx context.Context, taskID string, files []string, label string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	snapshot := SessionSnapshot{
		ID:        fmt.Sprintf("snap-%d", time.Now().UnixNano()),
		TaskID:    taskID,
		Label:     label,
		Files:     make(map[string]string),
		CreatedAt: time.Now(),
	}

	// Read file contents
	for _, filePath := range files {
		content, err := os.ReadFile(filePath)
		if err != nil {
			// File doesn't exist yet, that's okay
			continue
		}
		snapshot.Files[filePath] = string(content)
	}

	sm.snapshots[taskID] = append(sm.snapshots[taskID], snapshot)

	return sm.persistSnapshot(snapshot)
}

// Rollback restores files to a previous snapshot.
func (sm *SessionManager) Rollback(ctx context.Context, taskID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	snapshots, exists := sm.snapshots[taskID]
	if !exists || len(snapshots) == 0 {
		return fmt.Errorf("no snapshots found for task %s", taskID)
	}

	// Get the most recent snapshot
	snapshot := snapshots[len(snapshots)-1]

	// Restore files
	for filePath, content := range snapshot.Files {
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to restore file %s: %w", filePath, err)
		}
	}

	return nil
}

// RollbackToSnapshot restores files to a specific snapshot.
func (sm *SessionManager) RollbackToSnapshot(ctx context.Context, snapshotID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Find the snapshot
	for _, snapshots := range sm.snapshots {
		for _, snapshot := range snapshots {
			if snapshot.ID == snapshotID {
				// Restore files
				for filePath, content := range snapshot.Files {
					dir := filepath.Dir(filePath)
					if err := os.MkdirAll(dir, 0755); err != nil {
						return fmt.Errorf("failed to create directory %s: %w", dir, err)
					}
					if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
						return fmt.Errorf("failed to restore file %s: %w", filePath, err)
					}
				}
				return nil
			}
		}
	}

	return fmt.Errorf("snapshot %s not found", snapshotID)
}

// ListSnapshots returns all snapshots for a task.
func (sm *SessionManager) ListSnapshots(taskID string) []SessionSnapshot {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if snapshots, exists := sm.snapshots[taskID]; exists {
		return snapshots
	}
	return nil
}

// DiffSnapshot compares current files with a snapshot.
func (sm *SessionManager) DiffSnapshot(snapshotID string) (map[string]FileDiff, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Find the snapshot
	for _, snapshots := range sm.snapshots {
		for _, snapshot := range snapshots {
			if snapshot.ID == snapshotID {
				diffs := make(map[string]FileDiff)
				for filePath, snapshotContent := range snapshot.Files {
					currentContent, err := os.ReadFile(filePath)
					if err != nil {
						diffs[filePath] = FileDiff{
							Path:       filePath,
							Status:     "deleted",
							OldContent: snapshotContent,
						}
						continue
					}
					if string(currentContent) != snapshotContent {
						diffs[filePath] = FileDiff{
							Path:       filePath,
							Status:     "modified",
							OldContent: snapshotContent,
							NewContent: string(currentContent),
						}
					}
				}
				return diffs, nil
			}
		}
	}

	return nil, fmt.Errorf("snapshot %s not found", snapshotID)
}

// FileDiff represents the difference between snapshot and current file.
type FileDiff struct {
	Path       string
	Status     string // "modified", "deleted", "created"
	OldContent string
	NewContent string
}

// CleanSnapshots removes all snapshots for a task.
func (sm *SessionManager) CleanSnapshots(taskID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.snapshots, taskID)

	// Remove persisted snapshots
	snapshotDir := filepath.Join(sm.baseDir, taskID)
	return os.RemoveAll(snapshotDir)
}

// persistSnapshot saves a snapshot to disk.
func (sm *SessionManager) persistSnapshot(snapshot SessionSnapshot) error {
	snapshotDir := filepath.Join(sm.baseDir, snapshot.TaskID)
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	snapshotFile := filepath.Join(snapshotDir, snapshot.ID+".json")
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	return os.WriteFile(snapshotFile, data, 0644)
}

// LoadSnapshots loads snapshots from disk.
func (sm *SessionManager) LoadSnapshots(taskID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	snapshotDir := filepath.Join(sm.baseDir, taskID)
	entries, err := os.ReadDir(snapshotDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read snapshot directory: %w", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		snapshotFile := filepath.Join(snapshotDir, entry.Name())
		data, err := os.ReadFile(snapshotFile)
		if err != nil {
			continue
		}

		var snapshot SessionSnapshot
		if err := json.Unmarshal(data, &snapshot); err != nil {
			continue
		}

		sm.snapshots[taskID] = append(sm.snapshots[taskID], snapshot)
	}

	return nil
}
