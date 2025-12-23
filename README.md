# Crush 🚀

**The Ultimate AI-Powered CLI Tool**

Crush is a next-generation AI CLI tool that combines the power of large language models with advanced orchestration, multi-agent collaboration, and intelligent code assistance. Built to be the most comprehensive AI development assistant available.

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
git clone https://github.com/uglyswap/crush.git
cd crush

# Build
go build -o crush ./cmd/crush

# Install globally (Linux/macOS)
sudo mv crush /usr/local/bin/

# Or add to PATH (Windows)
# Move crush.exe to a directory in your PATH
```

### Using Go Install

```bash
go install github.com/uglyswap/crush/cmd/crush@latest
```

### From Releases

Download pre-built binaries from the [Releases](https://github.com/uglyswap/crush/releases) page.

---

## 🚀 Quick Start

### Basic Usage

```bash
# Start interactive session
crush

# Run with a prompt
crush "Explain this codebase"

# Run in a specific directory
crush --cwd /path/to/project "Add tests for the API"

# Use extended thinking for complex tasks
crush --thinking ultrathink "Refactor the authentication system"
```

### Configuration

Create `~/.crush/config.yaml`:

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
  - ~/.crush/skills
  - ./.crush/skills

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

---

## 🧠 Extended Thinking

Crush supports extended thinking mode for deeper reasoning on complex tasks.

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

## 🏗️ Architecture

```
crush/
├── cmd/
│   └── crush/              # Main entry point
├── internal/
│   ├── agent/              # Core agent implementation
│   │   └── tools/          # All tool implementations
│   ├── orchestrator/       # Multi-agent orchestration
│   │   ├── orchestrator.go # Core orchestrator
│   │   ├── agent.go        # 28 specialized agents
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
crush "Build a REST API with authentication, tests, and documentation"
```

Crush will:
1. Analyze the task and select appropriate agents
2. Create a session with snapshots for safety
3. Execute agents in optimal sequence
4. Validate quality at each step
5. Generate comprehensive output

### Plan Mode

```bash
crush "Refactor the authentication system to use JWT"
```

Crush will:
1. Enter plan mode automatically for complex tasks
2. Explore the codebase
3. Present an implementation plan for approval
4. Execute only after user approval

### Using Skills

```bash
# Invoke a skill
crush "/commit -m 'Add new feature'"

# Review a PR
crush "/review-pr 123"

# Use a specialized skill
crush "/supabase-patterns Create RLS policies for user table"
```

### Interactive Questions

Crush will ask clarifying questions when needed:

```
? Which authentication method do you prefer?
  > JWT (Recommended)
    Session-based
    OAuth 2.0
    Other
```

---

## 🔧 Creating Custom Skills

Create a markdown file in `~/.crush/skills/my-skill.md`:

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
git clone https://github.com/uglyswap/crush.git
cd crush

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o crush ./cmd/crush

# Run locally
./crush
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

**Made with ❤️ by the Crush Team**
