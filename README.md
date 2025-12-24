# Push 🚀

**The Ultimate AI-Powered CLI Tool**

Push is a next-generation AI CLI tool that combines the power of large language models with advanced orchestration, multi-agent collaboration, and intelligent code assistance. Built to be the most comprehensive AI development assistant available.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

---

## ✨ Features

### 🧠 Extended Thinking Mode
- **4 Thinking Levels** for progressive reasoning depth:
  - `think` - Basic reasoning (1K token budget)
  - `think_hard` - Moderate analysis (4K token budget)
  - `think_harder` - Deep investigation (16K token budget)
  - `ultrathink` - Maximum thoroughness (32K token budget)
- Automatic activation based on task complexity
- Configurable per-model and per-request

### 🤖 Multi-Agent Orchestration
- **28 Specialized Agents** organized into squads (Frontend, Backend, Data, Security, QA, DevOps, AI/ML)
- **Trust Cascade System** (L0-L4) for progressive autonomy
- **Quality Scoring** with 4 dimensions: Completeness, Precision, Coherence, Context Retention
- **Session Snapshots** for safe rollback
- **Adaptive Handoff Protocol** with token levels (400/1000/2500)

### 📋 Plan Mode
- Enter plan mode for complex implementations
- Design approaches before writing code
- Get user approval on implementation plans
- Record architectural decisions and risks

### 🛠️ Comprehensive Tool Suite

| Tool | Description |
|------|-------------|
| **Read** | Read files with line number support |
| **Write** | Write files safely |
| **Edit** | Precise string replacement edits |
| **Bash** | Execute shell commands |
| **Glob** | Fast file pattern matching |
| **Grep** | Powerful content search with ripgrep |
| **LSP** | Full Language Server Protocol support |
| **WebFetch** | Fetch and process web content |
| **WebSearch** | Search the web with citations |
| **Task** | Launch specialized subagents |
| **AskUserQuestion** | Interactive user questioning |
| **TodoWrite** | Task tracking and management |
| **NotebookEdit** | Jupyter notebook manipulation |
| **KillShell** | Background process management |
| **Skill** | Invoke specialized skills |
| **EnterPlanMode** | Start implementation planning |
| **ExitPlanMode** | Finalize and submit plans |

### 🎯 Skills System
- **266+ Specialized Skills** covering:
  - Sciences & Bioinformatics
  - Cloud & DevOps (Cloudflare, AWS, etc.)
  - Frontend Frameworks (React, Next.js, Vue, etc.)
  - AI & ML (OpenAI, Claude, Gemini)
  - Document Processing
  - Business Automation
- Load skills from multiple sources (local, user, project)
- Markdown-based skill definitions with frontmatter

### 💡 Advanced LSP Integration
- Go to Definition
- Find References
- Hover Information
- Document Symbols
- Workspace Symbol Search
- Go to Implementation
- Call Hierarchy (Incoming/Outgoing)

### 🗄️ Smart Caching
- LRU eviction with configurable max size
- TTL-based expiration
- Persistent cache to disk
- Auto-cleanup of expired entries

---

## 📦 Installation

### Prerequisites
- Go 1.21 or later
- Git

### Quick Install

```bash
# Clone the repository
git clone https://github.com/uglyswap/push.git
cd push

# Download dependencies
go mod download

# Build and install
go install .

# The binary will be installed to $GOPATH/bin/push (or $HOME/go/bin/push)
# Make sure this directory is in your PATH
```

### Alternative: Build Locally

```bash
# Clone the repository
git clone https://github.com/uglyswap/push.git
cd push

# Download dependencies and build
go mod download
go build -o push .

# Install globally (Linux/macOS)
sudo mv push /usr/local/bin/

# Or add to PATH (Windows)
# Move push.exe to a directory in your PATH
```

### From Releases

Download pre-built binaries from the [Releases](https://github.com/uglyswap/push/releases) page.

### Compatibility Notes

This version uses a compatibility layer (`internal/compat/`) to work with bubbletea v1 and lipgloss v1. The codebase has been adapted from charm v2 libraries to ensure compatibility with the current stable releases of the Charm ecosystem.

---

## 🚀 Quick Start

### Basic Usage

```bash
# Start interactive session
push

# Run with a prompt
push "Explain this codebase"

# Run in a specific directory
crush --cwd /path/to/project "Add tests for the API"

# Use extended thinking for complex tasks
crush --thinking ultrathink "Refactor the authentication system"
```

### Configuration

Create `~/.push/config.yaml`:

```yaml
# API Configuration
provider: anthropic  # anthropic, openai, or custom
model: claude-sonnet-4-20250514

# Extended Thinking
thinking:
  default_level: think_hard  # think, think_hard, think_harder, ultrathink
  auto_escalate: true        # Automatically increase level for complex tasks

# Behavior
auto_approve: false
max_tokens: 8192

# Skills
skills_directories:
  - ~/.push/skills
  - ./.push/skills

# Cache
cache:
  enabled: true
  max_size: 1000
  ttl: 15m
  persist: true

# Orchestrator
orchestrator:
  trust_level: L2
  quality_threshold: 60
```

### Environment Variables

```bash
export ANTHROPIC_API_KEY="your-api-key"
# or
export OPENAI_API_KEY="your-api-key"
```

### Providers

Crush uses embedded AI provider configurations out of the box. No external network connection is required for provider setup - just set your API key and you're ready to go.

Supported providers:
- **Anthropic** (Claude models) - Recommended
- **OpenAI** (GPT models)
- **Custom** providers via configuration

---

## 🧠 Extended Thinking

Push supports extended thinking mode for deeper reasoning on complex tasks.

### Thinking Levels

| Level | Token Budget | Use Case |
|-------|-------------|----------|
| `think` | 1,024 | Simple analysis, quick fixes |
| `think_hard` | 4,096 | Bug investigation, code review |
| `think_harder` | 16,384 | Architecture decisions, complex debugging |
| `ultrathink` | 32,768 | Critical systems, security analysis |

### Usage Examples

```bash
# Command line
crush --thinking ultrathink "Design a microservices architecture"

# In configuration
models:
  large:
    model: claude-sonnet-4-20250514
    provider: anthropic
    thinking_level: think_hard
```

### Auto-Escalation Keywords

Crush automatically increases thinking depth when it detects:
- `architecture`, `design system`, `migration` → ultrathink
- `bug`, `debug`, `performance`, `optimize` → think_harder
- `security`, `auth`, `API design` → think_hard

---

## 🤖 Subagent Execution Engine

Push features a powerful subagent execution system that connects specialized agents to the LLM with squad-specific prompts and quality scoring.

### Executor Architecture

The executor (`internal/orchestrator/executor.go`) is the core component that:
1. Builds context-aware prompts from templates
2. Configures thinking levels based on task complexity
3. Parses structured YAML responses
4. Scores agent outputs on 4 dimensions

```go
// Create an executor with custom configuration
executor := NewExecutor(ExecutorConfig{
    ThinkingLevel:   ThinkingLevelThinkHard,
    MaxOutputTokens: 8192,
    Temperature:     0.7,
})

// Execute an agent
result, err := executor.ExecuteAgent(ctx, agent, agentContext)
```

### Thinking Levels (Internal)

```go
type ThinkingLevel string

const (
    ThinkingLevelNone        ThinkingLevel = ""           // No extended thinking
    ThinkingLevelThink       ThinkingLevel = "think"      // 1024 tokens
    ThinkingLevelThinkHard   ThinkingLevel = "think_hard" // 4096 tokens
    ThinkingLevelThinkHarder ThinkingLevel = "think_harder" // 16384 tokens
    ThinkingLevelUltrathink  ThinkingLevel = "ultrathink" // 32768 tokens
)
```

### Agent Quality Scoring

Every agent output is scored on 4 weighted dimensions:

| Dimension | Weight | Description |
|-----------|--------|-------------|
| **Completeness** | 30% | Did the agent address all aspects of the task? |
| **Precision** | 30% | Is the output accurate and free of errors? |
| **Coherence** | 25% | Is the reasoning logical and well-structured? |
| **Context Retention** | 15% | Does it maintain consistency with prior context? |

```go
type AgentScore struct {
    Completeness     float64 // 0-100
    Precision        float64 // 0-100
    Coherence        float64 // 0-100
    ContextRetention float64 // 0-100
}

// Weighted total calculation
total := (score.Completeness * 0.30) +
         (score.Precision * 0.30) +
         (score.Coherence * 0.25) +
         (score.ContextRetention * 0.15)
```

### Agent Output Format

Agents respond in a structured YAML format for reliable parsing:

```yaml
agent_output:
  status: success  # success, partial, blocked, failed
  confidence: 0.92
  summary: "Brief description of what was accomplished"
  artifacts:
    - type: code
      path: "src/components/Button.tsx"
      content: |
        // Generated code here
    - type: suggestion
      content: "Consider adding unit tests"
  next_steps:
    - "Run tests to verify"
    - "Review for edge cases"
  blockers: []
  handoff_context: "Context for next agent if needed"
```

### Squad-Specific Prompts

Each squad has a specialized system prompt (`internal/orchestrator/prompts.go`):

| Squad | Focus Areas |
|-------|-------------|
| **Frontend** | React, TypeScript, accessibility, responsive design, performance |
| **Backend** | API design, database optimization, security, scalability |
| **Data** | Schema design, query optimization, analytics, data pipelines |
| **Business** | Product strategy, user research, compliance, growth metrics |
| **DevOps** | Infrastructure, CI/CD, monitoring, security automation |
| **QA** | Test coverage, edge cases, regression prevention, quality gates |
| **Performance** | Profiling, optimization, caching, bundle size |
| **Documentation** | API docs, tutorials, code comments, architecture diagrams |
| **Accessibility** | WCAG compliance, screen readers, keyboard navigation, i18n |
| **AI/ML** | Model integration, prompt engineering, ML pipelines |

### Automatic Thinking Level Selection

The executor automatically selects thinking levels based on task complexity:

```go
func GetThinkingLevelForTask(task *Task, keywords []string) ThinkingLevel {
    taskLower := strings.ToLower(task.Title + " " + task.Description)
    
    // Ultra-complex tasks
    ultrathinkKeywords := []string{"architecture", "migration", "security audit", "refactor"}
    for _, kw := range ultrathinkKeywords {
        if strings.Contains(taskLower, kw) {
            return ThinkingLevelUltrathink
        }
    }
    
    // Complex tasks
    thinkHarderKeywords := []string{"debug", "performance", "optimize", "complex"}
    // ... and so on
}
```

---

## 🏗️ Architecture

```
push/
├── cmd/
│   └── push/              # Main entry point
├── internal/
│   ├── agent/              # Core agent implementation
│   │   └── tools/          # All tool implementations
│   ├── orchestrator/       # Multi-agent orchestration
│   │   ├── orchestrator.go # Core orchestrator
│   │   ├── agent.go        # 28 specialized agents
│   │   ├── executor.go     # Subagent execution engine
│   │   ├── prompts.go      # Squad-specific prompts
│   │   ├── handoff.go      # Handoff protocol
│   │   ├── scoring.go      # Quality scoring
│   │   ├── trust.go        # Trust cascade
│   │   └── session.go      # Session management
│   ├── planmode/           # Plan mode system
│   ├── skills/             # Skills system
│   │   ├── skill.go        # Skill definition
│   │   ├── loader.go       # Skill loading
│   │   ├── registry.go     # Skill registry
│   │   └── invoker.go      # Skill invocation
│   ├── config/             # Configuration
│   │   ├── config.go       # Main config
│   │   └── thinking.go     # Extended thinking
│   ├── lsp/                # LSP integration
│   ├── cache/              # Caching system
│   └── csync/              # Concurrent data structures
└── docs/                   # Documentation
```

### Agent Squads

| Squad | Agents |
|-------|--------|
| **Frontend** | ui-ux-designer, frontend-developer, component-architect |
| **Backend** | api-architect, backend-developer, integration-specialist |
| **Data** | database-architect, analytics-engineer, search-rag-specialist |
| **Business** | product-manager, copywriter, pricing-strategist, compliance-officer, growth-engineer |
| **DevOps** | infrastructure-engineer, security-engineer, monitoring-engineer |
| **QA** | test-engineer, code-reviewer |
| **Performance** | performance-engineer, bundle-optimizer, database-optimizer |
| **Documentation** | technical-writer, api-documenter |
| **Accessibility** | accessibility-expert, i18n-specialist |
| **AI/ML** | ai-engineer, ml-ops-engineer |

---

## 📖 Usage Examples

### Multi-Agent Task

```bash
push "Build a REST API with authentication, tests, and documentation"
```

Push will:
1. Analyze the task and select appropriate agents
2. Create a session with snapshots for safety
3. Execute agents in optimal sequence
4. Validate quality at each step
5. Generate comprehensive output

### Plan Mode

```bash
push "Refactor the authentication system to use JWT"
```

Push will:
1. Enter plan mode automatically for complex tasks
2. Explore the codebase
3. Present an implementation plan for approval
4. Execute only after user approval

### Using Skills

```bash
# Invoke a skill
push "/commit -m 'Add new feature'"

# Review a PR
push "/review-pr 123"

# Use a specialized skill
push "/supabase-patterns Create RLS policies for user table"
```

### Interactive Questions

Push will ask clarifying questions when needed:

```
? Which authentication method do you prefer?
  > JWT (Recommended)
    Session-based
    OAuth 2.0
    Other
```

### Subagent Execution

```go
// Programmatic usage example
import "github.com/uglyswap/push/internal/orchestrator"

// Create orchestrator
orch := orchestrator.NewOrchestrator()

// Define task
task := &orchestrator.Task{
    Title:       "Implement user authentication",
    Description: "Add JWT-based auth with refresh tokens",
    Type:        orchestrator.TaskTypeBackend,
}

// Execute with automatic agent selection
result, err := orch.ExecuteTask(ctx, task)
if err != nil {
    log.Fatal(err)
}

// Check quality score
if result.Score.WeightedTotal() < 70 {
    log.Warn("Quality below threshold, review recommended")
}
```

---

## 🔧 Creating Custom Skills

Create a markdown file in `~/.push/skills/my-skill.md`:

```markdown
---
name: my-skill
description: Description of what this skill does
allowed-tools: Read,Write,Edit,Bash
---

# My Custom Skill

Instructions for the skill...

## Usage

{{args}}
```

---

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone
git clone https://github.com/uglyswap/push.git
cd push

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o push .

# Run locally
./push
```

### Running Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/orchestrator/...
```

---

## 📋 Roadmap

- [x] Multi-agent orchestration
- [x] Plan mode system
- [x] Skills system
- [x] Full LSP integration
- [x] Web fetch and search
- [x] Caching system
- [x] Notebook editing
- [x] Extended thinking mode
- [x] Subagent execution engine
- [x] Squad-specific prompts
- [x] Quality scoring system
- [ ] Remote MCP server support
- [ ] Plugin system
- [ ] Web UI
- [ ] Team collaboration features

---

## 📜 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for TUI
- Uses [Fantasy](https://charm.sh/fantasy) for LLM abstraction
- Inspired by Claude Code CLI

---

**Made with ❤️ by the Push Team**
