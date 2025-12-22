// Package planmode provides structured implementation planning capabilities.
// Plan mode allows the agent to explore the codebase, design implementation
// approaches, and get user approval before writing code.
package planmode

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// PlanStatus represents the current status of a plan.
type PlanStatus string

const (
	PlanStatusDraft      PlanStatus = "draft"
	PlanStatusPending    PlanStatus = "pending_approval"
	PlanStatusApproved   PlanStatus = "approved"
	PlanStatusRejected   PlanStatus = "rejected"
	PlanStatusExecuting  PlanStatus = "executing"
	PlanStatusCompleted  PlanStatus = "completed"
	PlanStatusAbandoned  PlanStatus = "abandoned"
)

// PlanStep represents a single step in an implementation plan.
type PlanStep struct {
	ID          string     `yaml:"id" json:"id"`
	Title       string     `yaml:"title" json:"title"`
	Description string     `yaml:"description" json:"description"`
	Files       []string   `yaml:"files,omitempty" json:"files,omitempty"`
	DependsOn   []string   `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	Completed   bool       `yaml:"completed" json:"completed"`
	Notes       string     `yaml:"notes,omitempty" json:"notes,omitempty"`
}

// ArchitecturalDecision records a decision made during planning.
type ArchitecturalDecision struct {
	ID                   string   `yaml:"id" json:"id"`
	Decision             string   `yaml:"decision" json:"decision"`
	Rationale            string   `yaml:"rationale" json:"rationale"`
	AlternativesRejected []string `yaml:"alternatives_rejected,omitempty" json:"alternatives_rejected,omitempty"`
	Impact               string   `yaml:"impact,omitempty" json:"impact,omitempty"`
}

// RiskAssessment captures potential risks identified during planning.
type RiskAssessment struct {
	Risk       string `yaml:"risk" json:"risk"`
	Severity   string `yaml:"severity" json:"severity"` // low, medium, high, critical
	Mitigation string `yaml:"mitigation" json:"mitigation"`
}

// Plan represents a complete implementation plan.
type Plan struct {
	ID          string                  `yaml:"id" json:"id"`
	Title       string                  `yaml:"title" json:"title"`
	Objective   string                  `yaml:"objective" json:"objective"`
	Scope       string                  `yaml:"scope" json:"scope"`
	Approach    string                  `yaml:"approach" json:"approach"`
	Status      PlanStatus              `yaml:"status" json:"status"`
	Steps       []PlanStep              `yaml:"steps" json:"steps"`
	Decisions   []ArchitecturalDecision `yaml:"decisions,omitempty" json:"decisions,omitempty"`
	Risks       []RiskAssessment        `yaml:"risks,omitempty" json:"risks,omitempty"`
	FilesAffected []string              `yaml:"files_affected,omitempty" json:"files_affected,omitempty"`
	CreatedAt   time.Time               `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time               `yaml:"updated_at" json:"updated_at"`
	ApprovedAt  *time.Time              `yaml:"approved_at,omitempty" json:"approved_at,omitempty"`
}

// PlanModeManager manages the plan mode state and operations.
type PlanModeManager struct {
	mu           sync.RWMutex
	active       bool
	currentPlan  *Plan
	planDir      string
	onPlanChange func(*Plan)
}

// NewPlanModeManager creates a new plan mode manager.
func NewPlanModeManager(planDir string) *PlanModeManager {
	return &PlanModeManager{
		planDir: planDir,
	}
}

// SetOnPlanChange sets a callback for plan changes.
func (pm *PlanModeManager) SetOnPlanChange(fn func(*Plan)) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.onPlanChange = fn
}

// IsActive returns whether plan mode is currently active.
func (pm *PlanModeManager) IsActive() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.active
}

// GetCurrentPlan returns the current plan if one exists.
func (pm *PlanModeManager) GetCurrentPlan() *Plan {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.currentPlan
}

// EnterPlanMode activates plan mode and creates a new draft plan.
func (pm *PlanModeManager) EnterPlanMode(ctx context.Context, title, objective string) (*Plan, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.active {
		return nil, fmt.Errorf("plan mode is already active with plan: %s", pm.currentPlan.Title)
	}

	now := time.Now()
	plan := &Plan{
		ID:        fmt.Sprintf("plan_%d", now.UnixNano()),
		Title:     title,
		Objective: objective,
		Status:    PlanStatusDraft,
		Steps:     []PlanStep{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	pm.active = true
	pm.currentPlan = plan

	if pm.onPlanChange != nil {
		pm.onPlanChange(plan)
	}

	return plan, nil
}

// UpdatePlan updates the current plan with new information.
func (pm *PlanModeManager) UpdatePlan(ctx context.Context, updates func(*Plan)) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.active || pm.currentPlan == nil {
		return fmt.Errorf("no active plan")
	}

	updates(pm.currentPlan)
	pm.currentPlan.UpdatedAt = time.Now()

	if pm.onPlanChange != nil {
		pm.onPlanChange(pm.currentPlan)
	}

	return nil
}

// AddStep adds a new step to the current plan.
func (pm *PlanModeManager) AddStep(step PlanStep) error {
	return pm.UpdatePlan(context.Background(), func(p *Plan) {
		if step.ID == "" {
			step.ID = fmt.Sprintf("step_%d", len(p.Steps)+1)
		}
		p.Steps = append(p.Steps, step)
	})
}

// AddDecision adds an architectural decision to the current plan.
func (pm *PlanModeManager) AddDecision(decision ArchitecturalDecision) error {
	return pm.UpdatePlan(context.Background(), func(p *Plan) {
		if decision.ID == "" {
			decision.ID = fmt.Sprintf("decision_%d", len(p.Decisions)+1)
		}
		p.Decisions = append(p.Decisions, decision)
	})
}

// AddRisk adds a risk assessment to the current plan.
func (pm *PlanModeManager) AddRisk(risk RiskAssessment) error {
	return pm.UpdatePlan(context.Background(), func(p *Plan) {
		p.Risks = append(p.Risks, risk)
	})
}

// SetScope sets the scope of the current plan.
func (pm *PlanModeManager) SetScope(scope string) error {
	return pm.UpdatePlan(context.Background(), func(p *Plan) {
		p.Scope = scope
	})
}

// SetApproach sets the implementation approach.
func (pm *PlanModeManager) SetApproach(approach string) error {
	return pm.UpdatePlan(context.Background(), func(p *Plan) {
		p.Approach = approach
	})
}

// SetFilesAffected sets the list of files that will be affected.
func (pm *PlanModeManager) SetFilesAffected(files []string) error {
	return pm.UpdatePlan(context.Background(), func(p *Plan) {
		p.FilesAffected = files
	})
}

// SubmitForApproval marks the plan as pending approval.
func (pm *PlanModeManager) SubmitForApproval() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.active || pm.currentPlan == nil {
		return fmt.Errorf("no active plan")
	}

	if len(pm.currentPlan.Steps) == 0 {
		return fmt.Errorf("plan must have at least one step before submission")
	}

	pm.currentPlan.Status = PlanStatusPending
	pm.currentPlan.UpdatedAt = time.Now()

	// Save plan to file for persistence
	if err := pm.savePlan(); err != nil {
		return fmt.Errorf("failed to save plan: %w", err)
	}

	if pm.onPlanChange != nil {
		pm.onPlanChange(pm.currentPlan)
	}

	return nil
}

// ApprovePlan marks the plan as approved.
func (pm *PlanModeManager) ApprovePlan() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.active || pm.currentPlan == nil {
		return fmt.Errorf("no active plan")
	}

	if pm.currentPlan.Status != PlanStatusPending {
		return fmt.Errorf("plan must be pending approval, current status: %s", pm.currentPlan.Status)
	}

	now := time.Now()
	pm.currentPlan.Status = PlanStatusApproved
	pm.currentPlan.ApprovedAt = &now
	pm.currentPlan.UpdatedAt = now

	if err := pm.savePlan(); err != nil {
		return fmt.Errorf("failed to save plan: %w", err)
	}

	if pm.onPlanChange != nil {
		pm.onPlanChange(pm.currentPlan)
	}

	return nil
}

// RejectPlan marks the plan as rejected.
func (pm *PlanModeManager) RejectPlan(reason string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.active || pm.currentPlan == nil {
		return fmt.Errorf("no active plan")
	}

	pm.currentPlan.Status = PlanStatusRejected
	pm.currentPlan.UpdatedAt = time.Now()

	// Add rejection reason as a note on the last step or create new step
	if len(pm.currentPlan.Steps) > 0 {
		pm.currentPlan.Steps[len(pm.currentPlan.Steps)-1].Notes = "Rejected: " + reason
	}

	if err := pm.savePlan(); err != nil {
		return fmt.Errorf("failed to save plan: %w", err)
	}

	if pm.onPlanChange != nil {
		pm.onPlanChange(pm.currentPlan)
	}

	return nil
}

// MarkStepComplete marks a step as completed.
func (pm *PlanModeManager) MarkStepComplete(stepID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.active || pm.currentPlan == nil {
		return fmt.Errorf("no active plan")
	}

	for i := range pm.currentPlan.Steps {
		if pm.currentPlan.Steps[i].ID == stepID {
			pm.currentPlan.Steps[i].Completed = true
			pm.currentPlan.UpdatedAt = time.Now()

			if pm.onPlanChange != nil {
				pm.onPlanChange(pm.currentPlan)
			}
			return nil
		}
	}

	return fmt.Errorf("step not found: %s", stepID)
}

// ExitPlanMode exits plan mode and returns the final plan.
func (pm *PlanModeManager) ExitPlanMode(ctx context.Context) (*Plan, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.active {
		return nil, fmt.Errorf("plan mode is not active")
	}

	plan := pm.currentPlan

	// Mark as completed if all steps are done
	allComplete := true
	for _, step := range plan.Steps {
		if !step.Completed {
			allComplete = false
			break
		}
	}

	if allComplete && plan.Status == PlanStatusApproved {
		plan.Status = PlanStatusCompleted
	} else if plan.Status == PlanStatusDraft {
		plan.Status = PlanStatusAbandoned
	}

	plan.UpdatedAt = time.Now()

	if err := pm.savePlan(); err != nil {
		return nil, fmt.Errorf("failed to save final plan: %w", err)
	}

	pm.active = false
	pm.currentPlan = nil

	return plan, nil
}

// GetPlanSummary returns a markdown summary of the current plan.
func (pm *PlanModeManager) GetPlanSummary() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.currentPlan == nil {
		return "No active plan."
	}

	p := pm.currentPlan
	summary := fmt.Sprintf("## %s\n\n", p.Title)
	summary += fmt.Sprintf("**Status:** %s\n", p.Status)
	summary += fmt.Sprintf("**Objective:** %s\n\n", p.Objective)

	if p.Scope != "" {
		summary += fmt.Sprintf("**Scope:** %s\n\n", p.Scope)
	}

	if p.Approach != "" {
		summary += fmt.Sprintf("**Approach:** %s\n\n", p.Approach)
	}

	if len(p.Steps) > 0 {
		summary += "### Steps\n\n"
		for i, step := range p.Steps {
			status := "⬜"
			if step.Completed {
				status = "✅"
			}
			summary += fmt.Sprintf("%d. %s %s\n", i+1, status, step.Title)
			if step.Description != "" {
				summary += fmt.Sprintf("   %s\n", step.Description)
			}
		}
		summary += "\n"
	}

	if len(p.Decisions) > 0 {
		summary += "### Architectural Decisions\n\n"
		for _, d := range p.Decisions {
			summary += fmt.Sprintf("- **%s**: %s\n", d.Decision, d.Rationale)
		}
		summary += "\n"
	}

	if len(p.Risks) > 0 {
		summary += "### Risks\n\n"
		for _, r := range p.Risks {
			summary += fmt.Sprintf("- [%s] %s - Mitigation: %s\n", r.Severity, r.Risk, r.Mitigation)
		}
		summary += "\n"
	}

	if len(p.FilesAffected) > 0 {
		summary += "### Files Affected\n\n"
		for _, f := range p.FilesAffected {
			summary += fmt.Sprintf("- `%s`\n", f)
		}
	}

	return summary
}

// ToYAML returns the current plan as YAML.
func (pm *PlanModeManager) ToYAML() (string, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.currentPlan == nil {
		return "", fmt.Errorf("no active plan")
	}

	data, err := yaml.Marshal(pm.currentPlan)
	if err != nil {
		return "", fmt.Errorf("failed to marshal plan: %w", err)
	}

	return string(data), nil
}

// savePlan saves the current plan to a file.
func (pm *PlanModeManager) savePlan() error {
	if pm.currentPlan == nil {
		return nil
	}

	if pm.planDir == "" {
		return nil // No persistence configured
	}

	if err := os.MkdirAll(pm.planDir, 0755); err != nil {
		return fmt.Errorf("failed to create plan directory: %w", err)
	}

	filename := filepath.Join(pm.planDir, pm.currentPlan.ID+".yaml")

	data, err := yaml.Marshal(pm.currentPlan)
	if err != nil {
		return fmt.Errorf("failed to marshal plan: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write plan file: %w", err)
	}

	return nil
}

// LoadPlan loads a plan from file by ID.
func (pm *PlanModeManager) LoadPlan(planID string) (*Plan, error) {
	if pm.planDir == "" {
		return nil, fmt.Errorf("no plan directory configured")
	}

	filename := filepath.Join(pm.planDir, planID+".yaml")

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read plan file: %w", err)
	}

	var plan Plan
	if err := yaml.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plan: %w", err)
	}

	return &plan, nil
}

// ListPlans returns all saved plans.
func (pm *PlanModeManager) ListPlans() ([]*Plan, error) {
	if pm.planDir == "" {
		return nil, fmt.Errorf("no plan directory configured")
	}

	entries, err := os.ReadDir(pm.planDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Plan{}, nil
		}
		return nil, fmt.Errorf("failed to read plan directory: %w", err)
	}

	var plans []*Plan
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		planID := entry.Name()[:len(entry.Name())-5] // Remove .yaml
		plan, err := pm.LoadPlan(planID)
		if err != nil {
			continue // Skip invalid plans
		}
		plans = append(plans, plan)
	}

	return plans, nil
}
