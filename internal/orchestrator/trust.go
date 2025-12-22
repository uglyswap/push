package orchestrator

import (
	"sync"
	"time"
)

// TrustLevelValue represents the numeric trust level.
type TrustLevelValue int

const (
	TrustLevelQuarantine TrustLevelValue = iota // L0: Every change validated
	TrustLevelSupervised                         // L1: Every commit validated
	TrustLevelValidated                          // L2: Major changes validated
	TrustLevelTrusted                            // L3: End of task validated
	TrustLevelAutonomous                         // L4: End of session validated
)

// TrustLevel represents the current trust level with its configuration.
type TrustLevel struct {
	Level             TrustLevelValue
	Name              string
	ValidationFreq    string
	SnapshotFreq      string
	RollbackBehavior  string
	TasksCompleted    int
	TasksRemaining    int // Tasks remaining for promotion
	LastPromotion     time.Time
	LastDemotion      time.Time
	SuccessStreak     int
	FailureCount      int
}

// TrustManager manages trust levels and transitions.
type TrustManager struct {
	mu    sync.RWMutex
	level TrustLevel
}

// NewTrustManager creates a new trust manager starting at Supervised level.
func NewTrustManager() *TrustManager {
	return &TrustManager{
		level: TrustLevel{
			Level:            TrustLevelSupervised,
			Name:             "Supervised",
			ValidationFreq:   "every_commit",
			SnapshotFreq:     "before_risky",
			RollbackBehavior: "prompt",
			TasksRemaining:   5,
		},
	}
}

// GetLevel returns the current trust level.
func (tm *TrustManager) GetLevel() TrustLevel {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.level
}

// SetLevel sets the trust level directly.
func (tm *TrustManager) SetLevel(level TrustLevel) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.level = level
}

// RecordSuccess records a successful task completion.
func (tm *TrustManager) RecordSuccess() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.level.TasksCompleted++
	tm.level.SuccessStreak++
	tm.level.FailureCount = 0

	if tm.level.TasksRemaining > 0 {
		tm.level.TasksRemaining--
	}

	// Auto-promote if conditions are met
	if tm.level.TasksRemaining == 0 && tm.level.Level < TrustLevelAutonomous {
		tm.promote("Completed required tasks with success")
	}
}

// RecordFailure records a task failure.
func (tm *TrustManager) RecordFailure() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.level.SuccessStreak = 0
	tm.level.FailureCount++

	// Auto-demote after multiple failures
	if tm.level.FailureCount >= 3 && tm.level.Level > TrustLevelQuarantine {
		tm.demote("Multiple consecutive failures")
	}
}

// Promote increases the trust level.
func (tm *TrustManager) Promote(reason string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.promote(reason)
}

func (tm *TrustManager) promote(reason string) {
	if tm.level.Level >= TrustLevelAutonomous {
		return
	}

	tm.level.Level++
	tm.level.LastPromotion = time.Now()
	tm.updateLevelConfig()
}

// Demote decreases the trust level.
func (tm *TrustManager) Demote(reason string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.demote(reason)
}

func (tm *TrustManager) demote(reason string) {
	if tm.level.Level <= TrustLevelQuarantine {
		return
	}

	tm.level.Level--
	tm.level.LastDemotion = time.Now()
	tm.level.FailureCount = 0
	tm.updateLevelConfig()
}

// updateLevelConfig updates the level configuration based on current level.
func (tm *TrustManager) updateLevelConfig() {
	switch tm.level.Level {
	case TrustLevelQuarantine:
		tm.level.Name = "Quarantine"
		tm.level.ValidationFreq = "every_change"
		tm.level.SnapshotFreq = "always"
		tm.level.RollbackBehavior = "auto"
		tm.level.TasksRemaining = 10
	case TrustLevelSupervised:
		tm.level.Name = "Supervised"
		tm.level.ValidationFreq = "every_commit"
		tm.level.SnapshotFreq = "before_risky"
		tm.level.RollbackBehavior = "prompt"
		tm.level.TasksRemaining = 5
	case TrustLevelValidated:
		tm.level.Name = "Validated"
		tm.level.ValidationFreq = "major_changes"
		tm.level.SnapshotFreq = "on_request"
		tm.level.RollbackBehavior = "available"
		tm.level.TasksRemaining = 10
	case TrustLevelTrusted:
		tm.level.Name = "Trusted"
		tm.level.ValidationFreq = "end_of_task"
		tm.level.SnapshotFreq = "session_start"
		tm.level.RollbackBehavior = "on_request"
		tm.level.TasksRemaining = 20
	case TrustLevelAutonomous:
		tm.level.Name = "Autonomous"
		tm.level.ValidationFreq = "end_of_session"
		tm.level.SnapshotFreq = "never"
		tm.level.RollbackBehavior = "manual"
		tm.level.TasksRemaining = 0
	}
}

// RequiresSnapshot returns whether the current level requires snapshots.
func (tl TrustLevel) RequiresSnapshot() bool {
	return tl.Level <= TrustLevelValidated
}

// AutoRollback returns whether the current level auto-rollbacks on failure.
func (tl TrustLevel) AutoRollback() bool {
	return tl.Level == TrustLevelQuarantine
}

// ValidateAfterAgent returns whether to validate after each agent.
func (tl TrustLevel) ValidateAfterAgent() bool {
	return tl.Level <= TrustLevelSupervised
}

// ValidateAfterTask returns whether to validate at end of task.
func (tl TrustLevel) ValidateAfterTask() bool {
	return tl.Level <= TrustLevelTrusted
}
