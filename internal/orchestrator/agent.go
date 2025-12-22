package orchestrator

import (
	"context"
	"strings"
)

// AgentSquad represents a group of related agents.
type AgentSquad string

const (
	SquadFrontend      AgentSquad = "frontend"
	SquadBackend       AgentSquad = "backend"
	SquadData          AgentSquad = "data"
	SquadBusiness      AgentSquad = "business"
	SquadDevOps        AgentSquad = "devops"
	SquadQA            AgentSquad = "qa"
	SquadPerformance   AgentSquad = "performance"
	SquadDocumentation AgentSquad = "documentation"
	SquadAccessibility AgentSquad = "accessibility"
	SquadAI            AgentSquad = "ai"
)

// AgentModel represents the model tier for an agent.
type AgentModel string

const (
	ModelSonnet AgentModel = "sonnet"
	ModelOpus   AgentModel = "opus"
	ModelHaiku  AgentModel = "haiku"
)

// Agent represents a specialized AI agent.
type Agent struct {
	ID          string
	Name        string
	Description string
	Squad       AgentSquad
	Model       AgentModel
	Expertise   []string
	Keywords    []string
	Patterns    []string // Reference patterns to load
	Collaborates struct {
		ReceivesFrom []string
		TransmitsTo  []string
	}
}

// AgentContext provides context for agent execution.
type AgentContext struct {
	Task            *Task
	PreviousHandoff *Handoff
	TrustLevel      TrustLevel
	HandoffLevel    HandoffLevel
	ProjectContext  map[string]interface{}
}

// AgentResult represents the output from an agent execution.
type AgentResult struct {
	AgentID        string
	TaskCompleted  bool
	Summary        string
	Artifacts      []Artifact
	Decisions      []Decision
	CodeBlocks     []CodeBlock
	Issues         []Issue
	NextAgent      string
	HandoffContext string
	PriorityItems  []string
	Metrics        AgentMetrics
}

// Decision represents an architectural decision made by an agent.
type Decision struct {
	Decision            string
	Rationale           string
	AlternativesRejected []string
}

// CodeBlock represents generated code.
type CodeBlock struct {
	File    string
	Content string
}

// AgentMetrics tracks agent performance metrics.
type AgentMetrics struct {
	FilesCreated      int
	FilesModified     int
	LinesAdded        int
	TestsAdded        int
	TokensUsed        int64
	ExecutionTime     int64 // milliseconds
}

// Execute runs the agent with the given context.
// This is a placeholder that will be implemented by the actual agent execution.
func (a *Agent) Execute(ctx context.Context, agentCtx *AgentContext) (*AgentResult, error) {
	// This will be implemented to call the actual LLM with agent-specific prompts
	return &AgentResult{
		AgentID:       a.ID,
		TaskCompleted: true,
		Summary:       "Agent execution placeholder",
	}, nil
}

// MatchesTask checks if the agent is relevant for a given task description.
func (a *Agent) MatchesTask(description string) float64 {
	desc := strings.ToLower(description)
	var matches int
	var total int

	// Check keywords
	for _, kw := range a.Keywords {
		total++
		if strings.Contains(desc, strings.ToLower(kw)) {
			matches++
		}
	}

	// Check expertise
	for _, exp := range a.Expertise {
		total++
		if strings.Contains(desc, strings.ToLower(exp)) {
			matches++
		}
	}

	if total == 0 {
		return 0
	}
	return float64(matches) / float64(total)
}

// registerDefaultAgents registers all 28 specialized agents.
func (o *Orchestrator) registerDefaultAgents() {
	agents := []*Agent{
		// Frontend Squad
		{
			ID:          "ui-ux-designer",
			Name:        "UI/UX Designer",
			Description: "Expert in user interface design and user experience",
			Squad:       SquadFrontend,
			Model:       ModelSonnet,
			Expertise:   []string{"design", "ux", "ui", "wireframe", "prototype"},
			Keywords:    []string{"design", "style", "layout", "ux", "user experience", "interface"},
			Patterns:    []string{"ui-patterns", "design-system"},
		},
		{
			ID:          "frontend-developer",
			Name:        "Frontend Developer",
			Description: "Expert in React, TypeScript, and modern frontend development",
			Squad:       SquadFrontend,
			Model:       ModelSonnet,
			Expertise:   []string{"react", "typescript", "nextjs", "tailwind", "components"},
			Keywords:    []string{"react", "component", "ui", "tsx", "jsx", "frontend", "client"},
			Patterns:    []string{"react-patterns", "component-patterns"},
		},
		{
			ID:          "component-architect",
			Name:        "Component Architect",
			Description: "Expert in component architecture and design patterns",
			Squad:       SquadFrontend,
			Model:       ModelOpus,
			Expertise:   []string{"architecture", "patterns", "composition", "state management"},
			Keywords:    []string{"architecture", "pattern", "structure", "component", "design"},
			Patterns:    []string{"architecture-patterns", "state-patterns"},
		},

		// Backend Squad
		{
			ID:          "api-architect",
			Name:        "API Architect",
			Description: "Expert in REST/GraphQL API design and architecture",
			Squad:       SquadBackend,
			Model:       ModelOpus,
			Expertise:   []string{"api", "rest", "graphql", "schema", "contract"},
			Keywords:    []string{"api", "rest", "graphql", "schema", "contract", "endpoint"},
			Patterns:    []string{"api-patterns", "rest-patterns"},
		},
		{
			ID:          "backend-developer",
			Name:        "Backend Developer",
			Description: "Expert in server-side development and API implementation",
			Squad:       SquadBackend,
			Model:       ModelSonnet,
			Expertise:   []string{"nodejs", "typescript", "api", "server", "middleware"},
			Keywords:    []string{"api", "endpoint", "server", "route", "backend", "handler"},
			Patterns:    []string{"api-routes", "middleware-patterns"},
		},
		{
			ID:          "integration-specialist",
			Name:        "Integration Specialist",
			Description: "Expert in third-party integrations and webhooks",
			Squad:       SquadBackend,
			Model:       ModelSonnet,
			Expertise:   []string{"webhook", "oauth", "integration", "external api"},
			Keywords:    []string{"webhook", "integration", "external", "oauth", "api"},
			Patterns:    []string{"integration-patterns", "webhook-patterns"},
		},

		// Data Squad
		{
			ID:          "database-architect",
			Name:        "Database Architect",
			Description: "Expert in database design, schemas, and migrations",
			Squad:       SquadData,
			Model:       ModelOpus,
			Expertise:   []string{"database", "schema", "migration", "sql", "relations"},
			Keywords:    []string{"schema", "migration", "database", "sql", "relation", "table"},
			Patterns:    []string{"database-patterns", "migration-patterns"},
		},
		{
			ID:          "analytics-engineer",
			Name:        "Analytics Engineer",
			Description: "Expert in data analytics and business intelligence",
			Squad:       SquadData,
			Model:       ModelSonnet,
			Expertise:   []string{"analytics", "metrics", "dashboard", "reporting"},
			Keywords:    []string{"analytics", "metrics", "data", "dashboard", "report"},
			Patterns:    []string{"analytics-patterns"},
		},
		{
			ID:          "search-rag-specialist",
			Name:        "Search & RAG Specialist",
			Description: "Expert in search, vector databases, and RAG systems",
			Squad:       SquadData,
			Model:       ModelOpus,
			Expertise:   []string{"search", "vector", "rag", "embedding", "semantic"},
			Keywords:    []string{"search", "vector", "rag", "embedding", "semantic"},
			Patterns:    []string{"search-patterns", "rag-patterns"},
		},

		// Business Squad
		{
			ID:          "product-manager",
			Name:        "Product Manager",
			Description: "Expert in product requirements and specifications",
			Squad:       SquadBusiness,
			Model:       ModelSonnet,
			Expertise:   []string{"requirements", "specs", "user stories", "mvp"},
			Keywords:    []string{"requirement", "feature", "user story", "spec", "mvp"},
			Patterns:    []string{"product-patterns"},
		},
		{
			ID:          "copywriter",
			Name:        "Copywriter",
			Description: "Expert in UI copy and content writing",
			Squad:       SquadBusiness,
			Model:       ModelSonnet,
			Expertise:   []string{"copy", "content", "messaging", "tone"},
			Keywords:    []string{"copy", "text", "content", "message", "writing"},
			Patterns:    []string{"copy-patterns"},
		},
		{
			ID:          "pricing-strategist",
			Name:        "Pricing Strategist",
			Description: "Expert in pricing models and monetization",
			Squad:       SquadBusiness,
			Model:       ModelSonnet,
			Expertise:   []string{"pricing", "subscription", "monetization", "billing"},
			Keywords:    []string{"pricing", "subscription", "billing", "tier", "plan"},
			Patterns:    []string{"pricing-patterns"},
		},
		{
			ID:          "compliance-officer",
			Name:        "Compliance Officer",
			Description: "Expert in regulatory compliance and legal requirements",
			Squad:       SquadBusiness,
			Model:       ModelOpus,
			Expertise:   []string{"gdpr", "compliance", "privacy", "legal", "terms"},
			Keywords:    []string{"gdpr", "compliance", "privacy", "legal", "consent"},
			Patterns:    []string{"compliance-patterns"},
		},
		{
			ID:          "growth-engineer",
			Name:        "Growth Engineer",
			Description: "Expert in growth mechanics and user acquisition",
			Squad:       SquadBusiness,
			Model:       ModelSonnet,
			Expertise:   []string{"growth", "acquisition", "retention", "conversion"},
			Keywords:    []string{"growth", "onboarding", "conversion", "retention"},
			Patterns:    []string{"growth-patterns"},
		},

		// DevOps Squad
		{
			ID:          "infrastructure-engineer",
			Name:        "Infrastructure Engineer",
			Description: "Expert in CI/CD, deployment, and infrastructure",
			Squad:       SquadDevOps,
			Model:       ModelSonnet,
			Expertise:   []string{"cicd", "deploy", "docker", "kubernetes", "infrastructure"},
			Keywords:    []string{"deploy", "ci", "cd", "docker", "kubernetes", "infrastructure"},
			Patterns:    []string{"devops-patterns", "docker-patterns"},
		},
		{
			ID:          "security-engineer",
			Name:        "Security Engineer",
			Description: "Expert in application security and authentication",
			Squad:       SquadDevOps,
			Model:       ModelOpus,
			Expertise:   []string{"security", "auth", "owasp", "rls", "encryption"},
			Keywords:    []string{"security", "auth", "owasp", "rls", "vulnerability"},
			Patterns:    []string{"security-patterns", "auth-patterns"},
		},
		{
			ID:          "monitoring-engineer",
			Name:        "Monitoring Engineer",
			Description: "Expert in logging, monitoring, and observability",
			Squad:       SquadDevOps,
			Model:       ModelSonnet,
			Expertise:   []string{"logging", "monitoring", "alerting", "observability"},
			Keywords:    []string{"log", "monitor", "alert", "observability", "metrics"},
			Patterns:    []string{"monitoring-patterns"},
		},

		// QA Squad
		{
			ID:          "test-engineer",
			Name:        "Test Engineer",
			Description: "Expert in testing strategies and test implementation",
			Squad:       SquadQA,
			Model:       ModelSonnet,
			Expertise:   []string{"testing", "tdd", "unit test", "integration test", "e2e"},
			Keywords:    []string{"test", "coverage", "tdd", "vitest", "jest", "playwright"},
			Patterns:    []string{"testing-patterns"},
		},
		{
			ID:          "code-reviewer",
			Name:        "Code Reviewer",
			Description: "Expert in code quality and best practices",
			Squad:       SquadQA,
			Model:       ModelOpus,
			Expertise:   []string{"review", "quality", "refactor", "best practices"},
			Keywords:    []string{"review", "quality", "refactor", "clean code"},
			Patterns:    []string{"review-patterns"},
		},

		// Performance Squad
		{
			ID:          "performance-engineer",
			Name:        "Performance Engineer",
			Description: "Expert in performance optimization and profiling",
			Squad:       SquadPerformance,
			Model:       ModelOpus,
			Expertise:   []string{"performance", "optimization", "profiling", "benchmarking"},
			Keywords:    []string{"performance", "optimize", "speed", "benchmark", "profile"},
			Patterns:    []string{"performance-patterns"},
		},
		{
			ID:          "bundle-optimizer",
			Name:        "Bundle Optimizer",
			Description: "Expert in bundle size optimization and code splitting",
			Squad:       SquadPerformance,
			Model:       ModelSonnet,
			Expertise:   []string{"bundle", "webpack", "code splitting", "tree shaking"},
			Keywords:    []string{"bundle", "webpack", "split", "lazy", "chunk"},
			Patterns:    []string{"bundle-patterns"},
		},
		{
			ID:          "database-optimizer",
			Name:        "Database Optimizer",
			Description: "Expert in database query optimization and indexing",
			Squad:       SquadPerformance,
			Model:       ModelOpus,
			Expertise:   []string{"query optimization", "indexing", "n+1", "explain"},
			Keywords:    []string{"query", "index", "n+1", "optimize", "slow"},
			Patterns:    []string{"query-patterns"},
		},

		// Documentation Squad
		{
			ID:          "technical-writer",
			Name:        "Technical Writer",
			Description: "Expert in documentation and technical writing",
			Squad:       SquadDocumentation,
			Model:       ModelSonnet,
			Expertise:   []string{"documentation", "readme", "guide", "tutorial"},
			Keywords:    []string{"doc", "readme", "guide", "tutorial", "documentation"},
			Patterns:    []string{"doc-patterns"},
		},
		{
			ID:          "api-documenter",
			Name:        "API Documenter",
			Description: "Expert in API documentation and OpenAPI specs",
			Squad:       SquadDocumentation,
			Model:       ModelSonnet,
			Expertise:   []string{"openapi", "swagger", "api docs", "reference"},
			Keywords:    []string{"openapi", "swagger", "api doc", "reference"},
			Patterns:    []string{"api-doc-patterns"},
		},

		// Accessibility Squad
		{
			ID:          "accessibility-expert",
			Name:        "Accessibility Expert",
			Description: "Expert in WCAG compliance and accessibility",
			Squad:       SquadAccessibility,
			Model:       ModelSonnet,
			Expertise:   []string{"wcag", "a11y", "screen reader", "aria"},
			Keywords:    []string{"accessibility", "a11y", "wcag", "aria", "screen reader"},
			Patterns:    []string{"a11y-patterns"},
		},
		{
			ID:          "i18n-specialist",
			Name:        "i18n Specialist",
			Description: "Expert in internationalization and localization",
			Squad:       SquadAccessibility,
			Model:       ModelSonnet,
			Expertise:   []string{"i18n", "l10n", "translation", "locale"},
			Keywords:    []string{"i18n", "l10n", "translation", "locale", "language"},
			Patterns:    []string{"i18n-patterns"},
		},

		// AI/ML Squad
		{
			ID:          "ai-engineer",
			Name:        "AI Engineer",
			Description: "Expert in AI/LLM integration and prompt engineering",
			Squad:       SquadAI,
			Model:       ModelOpus,
			Expertise:   []string{"ai", "llm", "prompt", "openai", "anthropic"},
			Keywords:    []string{"ai", "llm", "prompt", "gpt", "claude", "model"},
			Patterns:    []string{"ai-patterns", "prompt-patterns"},
		},
		{
			ID:          "ml-ops-engineer",
			Name:        "ML Ops Engineer",
			Description: "Expert in ML operations and model deployment",
			Squad:       SquadAI,
			Model:       ModelOpus,
			Expertise:   []string{"mlops", "model deployment", "fine-tuning", "evaluation"},
			Keywords:    []string{"mlops", "model", "deploy", "fine-tune", "evaluate"},
			Patterns:    []string{"mlops-patterns"},
		},
	}

	for _, agent := range agents {
		o.agents[agent.ID] = agent
	}
}
