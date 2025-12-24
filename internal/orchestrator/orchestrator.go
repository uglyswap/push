// Package orchestrator provides multi-agent coordination for Crush.
//
// It implements a sophisticated agent orchestration system that manages
// specialized AI agents, handles task delegation, coordinates handoffs,
// and maintains quality through scoring and trust levels.
package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/uglyswap/push/internal/session"
)

// Orchestrator coordinates multiple specialized agents for complex tasks.
type Orchestrator struct {
	mu sync.RWMutex

	// Agent registry
	agents map[string]*Agent

	// Trust and scoring
	trustManager  *TrustManager
	scoringEngine *ScoringEngine

	// Session management
	sessionManager *SessionManager

	// Configuration
	config OrchestratorConfig

	// Active tasks
	activeTasks map[string]*Task
}

// OrchestratorConfig holds configuration for the orchestrator.
type OrchestratorConfig struct {
	// DefaultHandoffLevel is the default token level for handoffs
	DefaultHandoffLevel HandoffLevel

	// MaxConcurrentAgents limits parallel agent execution
	MaxConcurrentAgents int

	// EnableSnapshots enables session snapshots before risky operations
	EnableSnapshots bool

	// QualityThreshold is the minimum score for auto-approval
	QualityThreshold float64
}

// Task represents an active task being processed by agents.
type Task struct {
	ID          string
	Description string
	Status      TaskStatus
	Agents      []string
	CurrentAgent string
	StartTime   time.Time
	Handoffs    []Handoff
	Scores      map[string]AgentScore
	Artifacts   []Artifact
	Issues      []Issue
}

// TaskStatus represents the current status of a task.
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusRolledBack TaskStatus = "rolled_back"
)

// Artifact represents a file or resource created/modified during task execution.
type Artifact struct {
	Path        string
	Action      ArtifactAction
	Description string
	Snippet     string // First 50 lines for context
	CreatedAt   time.Time
}

// ArtifactAction represents what happened to an artifact.
type ArtifactAction string

const (
	ArtifactCreated  ArtifactAction = "created"
	ArtifactModified ArtifactAction = "modified"
	ArtifactDeleted  ArtifactAction = "deleted"
)

// Issue represents a problem identified during task execution.
type Issue struct {
	Severity      IssueSeverity
	Location      string
	Message       string
	FixSuggestion string
	AgentID       string
}

// IssueSeverity represents the severity of an issue.
type IssueSeverity string

const (
	IssueSeverityBlocker    IssueSeverity = "blocker"
	IssueSeverityCritical   IssueSeverity = "critical"
	IssueSeverityMajor      IssueSeverity = "major"
	IssueSeverityMinor      IssueSeverity = "minor"
	IssueSeveritySuggestion IssueSeverity = "suggestion"
)

// New creates a new Orchestrator with the given configuration.
func New(config OrchestratorConfig) *Orchestrator {
	if config.MaxConcurrentAgents == 0 {
		config.MaxConcurrentAgents = 3
	}
	if config.QualityThreshold == 0 {
		config.QualityThreshold = 0.75
	}
	if config.DefaultHandoffLevel == "" {
		config.DefaultHandoffLevel = HandoffStandard
	}

	o := &Orchestrator{
		agents:         make(map[string]*Agent),
		trustManager:   NewTrustManager(),
		scoringEngine:  NewScoringEngine(),
		sessionManager: NewSessionManager(),
		config:         config,
		activeTasks:    make(map[string]*Task),
	}

	// Register default agents
	o.registerDefaultAgents()

	return o
}

// RegisterAgent adds a new agent to the registry.
func (o *Orchestrator) RegisterAgent(agent *Agent) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, exists := o.agents[agent.ID]; exists {
		return fmt.Errorf("agent %s already registered", agent.ID)
	}

	o.agents[agent.ID] = agent
	return nil
}

// GetAgent retrieves an agent by ID.
func (o *Orchestrator) GetAgent(id string) (*Agent, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	agent, exists := o.agents[id]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", id)
	}
	return agent, nil
}

// SelectAgents chooses the best agents for a given task based on scoring.
func (o *Orchestrator) SelectAgents(ctx context.Context, taskDescription string) ([]*Agent, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Score all agents for this task
	scores := make(map[string]float64)
	for id, agent := range o.agents {
		score := o.scoringEngine.ScoreAgentForTask(agent, taskDescription)
		scores[id] = score
	}

	// Select top agents above threshold
	var selected []*Agent
	for id, score := range scores {
		if score >= 0.5 { // Minimum relevance threshold
			selected = append(selected, o.agents[id])
		}
	}

	// Sort by score (highest first)
	for i := 0; i < len(selected)-1; i++ {
		for j := i + 1; j < len(selected); j++ {
			if scores[selected[i].ID] < scores[selected[j].ID] {
				selected[i], selected[j] = selected[j], selected[i]
			}
		}
	}

	return selected, nil
}

// ExecuteTask runs a task through the selected agents.
func (o *Orchestrator) ExecuteTask(ctx context.Context, task *Task, sessionSvc session.Service) error {
	o.mu.Lock()
	o.activeTasks[task.ID] = task
	task.Status = TaskStatusInProgress
	task.StartTime = time.Now()
	o.mu.Unlock()

	defer func() {
		o.mu.Lock()
		delete(o.activeTasks, task.ID)
		o.mu.Unlock()
	}()

	// Get trust level for validation frequency
	trustLevel := o.trustManager.GetLevel()

	// Create snapshot if enabled and trust level requires it
	if o.config.EnableSnapshots && trustLevel.RequiresSnapshot() {
		if err := o.sessionManager.CreateSnapshot(ctx, task.ID); err != nil {
			return fmt.Errorf("failed to create snapshot: %w", err)
		}
	}

	// Execute each agent in sequence
	var lastHandoff *Handoff
	for _, agentID := range task.Agents {
		agent, err := o.GetAgent(agentID)
		if err != nil {
			return err
		}

		task.CurrentAgent = agentID

		// Prepare context for agent
		agentCtx := &AgentContext{
			Task:            task,
			PreviousHandoff: lastHandoff,
			TrustLevel:      trustLevel,
			HandoffLevel:    o.config.DefaultHandoffLevel,
		}

		// Execute agent
		result, err := agent.Execute(ctx, agentCtx)
		if err != nil {
			task.Status = TaskStatusFailed
			
			// Rollback if trust level requires it
			if trustLevel.AutoRollback() {
				if rbErr := o.sessionManager.Rollback(ctx, task.ID); rbErr != nil {
					return fmt.Errorf("execution failed and rollback failed: %v, %v", err, rbErr)
				}
				task.Status = TaskStatusRolledBack
			}
			return err
		}

		// Score agent performance
		score := o.scoringEngine.ScoreResult(result)
		task.Scores[agentID] = score

		// Check quality threshold
		if score.Total() < o.config.QualityThreshold {
			// Quality below threshold - may need intervention
			if trustLevel.Level <= TrustLevelSupervised {
				// For low trust levels, this is a failure
				task.Status = TaskStatusFailed
				return fmt.Errorf("agent %s scored %.2f, below threshold %.2f", 
					agentID, score.Total(), o.config.QualityThreshold)
			}
		}

		// Collect artifacts and issues
		task.Artifacts = append(task.Artifacts, result.Artifacts...)
		task.Issues = append(task.Issues, result.Issues...)

		// Prepare handoff for next agent
		if result.NextAgent != "" && result.NextAgent != "none" {
			lastHandoff = &Handoff{
				FromAgent:    agentID,
				ToAgent:      result.NextAgent,
				Context:      result.HandoffContext,
				Level:        o.config.DefaultHandoffLevel,
				PriorityItems: result.PriorityItems,
				Timestamp:    time.Now(),
			}
			task.Handoffs = append(task.Handoffs, *lastHandoff)
		}

		// Validate if trust level requires it
		if trustLevel.ValidateAfterAgent() {
			if err := o.validateTask(ctx, task); err != nil {
				return err
			}
		}
	}

	task.Status = TaskStatusCompleted

	// Update trust level based on outcome
	avgScore := o.calculateAverageScore(task)
	if avgScore >= 0.7 {
		o.trustManager.RecordSuccess()
	} else {
		o.trustManager.RecordFailure()
	}

	return nil
}

// validateTask runs validation checks on the current task state.
func (o *Orchestrator) validateTask(ctx context.Context, task *Task) error {
	// Check for blocker issues
	for _, issue := range task.Issues {
		if issue.Severity == IssueSeverityBlocker {
			return fmt.Errorf("blocker issue found: %s", issue.Message)
		}
	}
	return nil
}

// calculateAverageScore calculates the average score across all agents.
func (o *Orchestrator) calculateAverageScore(task *Task) float64 {
	if len(task.Scores) == 0 {
		return 0
	}

	var total float64
	for _, score := range task.Scores {
		total += score.Total()
	}
	return total / float64(len(task.Scores))
}

// GetActiveTask returns an active task by ID.
func (o *Orchestrator) GetActiveTask(id string) (*Task, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	task, exists := o.activeTasks[id]
	return task, exists
}

// ListAgents returns all registered agents.
func (o *Orchestrator) ListAgents() []*Agent {
	o.mu.RLock()
	defer o.mu.RUnlock()

	agents := make([]*Agent, 0, len(o.agents))
	for _, agent := range o.agents {
		agents = append(agents, agent)
	}
	return agents
}

// GetTrustLevel returns the current trust level.
func (o *Orchestrator) GetTrustLevel() TrustLevel {
	return o.trustManager.GetLevel()
}

// SetTrustLevel sets the trust level.
func (o *Orchestrator) SetTrustLevel(level TrustLevel) {
	o.trustManager.SetLevel(level)
}
