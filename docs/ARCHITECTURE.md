# Crush Architecture

This document provides an overview of Crush's architecture and design decisions.

## Overview

Crush is built as a modular, extensible AI CLI tool with several key components:

1. **Core Agent** - The main agent that processes user requests
2. **Orchestrator** - Coordinates multi-agent collaboration
3. **Tools** - Individual capabilities (read, write, search, etc.)
4. **Skills** - Specialized domain knowledge
5. **Plan Mode** - Structured implementation planning

## Core Components

### Agent (`internal/agent/`)

The core agent handles:
- User input processing
- LLM communication
- Tool execution
- Response streaming

```go
type Agent struct {
    llm       LLMClient
    tools     *ToolRegistry
    session   *Session
    config    *Config
}
```

### Orchestrator (`internal/orchestrator/`)

The orchestrator manages multi-agent collaboration:

#### Components:

1. **Agent Registry** - 28 specialized agents across 10 squads
2. **Scoring Engine** - Quality assessment with 4 dimensions
3. **Trust Manager** - Progressive autonomy levels (L0-L4)
4. **Session Manager** - Snapshots and rollback capability
5. **Handoff Protocol** - Inter-agent communication

#### Trust Levels:

| Level | Name | Validation | Autonomy |
|-------|------|------------|----------|
| L0 | Quarantine | Every change | Minimal |
| L1 | Supervised | Every commit | Low |
| L2 | Validated | Major changes | Medium |
| L3 | Trusted | End of task | High |
| L4 | Autonomous | End of session | Full |

#### Quality Scoring:

```go
type ScoringWeights struct {
    Completeness     float64 // 30%
    Precision        float64 // 30%
    Coherence        float64 // 25%
    ContextRetention float64 // 15%
}
```

### Tools (`internal/agent/tools/`)

Tools are individual capabilities with a standard interface:

```go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{}
    Execute(ctx context.Context, params json.RawMessage) (string, error)
    RequiresApproval() bool
}
```

#### Tool Categories:

1. **File Operations**: Read, Write, Edit, Glob
2. **Search**: Grep, LSP
3. **Execution**: Bash, KillShell
4. **Web**: WebFetch, WebSearch
5. **Agent**: Task, AskUserQuestion, TodoWrite
6. **Planning**: EnterPlanMode, ExitPlanMode, AddPlanStep
7. **Skills**: Skill
8. **Notebook**: NotebookEdit

### Skills (`internal/skills/`)

Skills provide specialized domain knowledge:

```go
type Skill struct {
    Name         string
    Description  string
    AllowedTools []string
    Content      string
    Source       string // local, user, project
}
```

#### Skill Loading:

1. **Local Skills** - Built-in skills
2. **User Skills** - `~/.crush/skills/`
3. **Project Skills** - `./.crush/skills/`

### Plan Mode (`internal/planmode/`)

Structured implementation planning:

```go
type Plan struct {
    Title     string
    Objective string
    Status    PlanStatus
    Steps     []PlanStep
    Decisions []ArchitecturalDecision
    Risks     []RiskAssessment
}
```

### Cache (`internal/cache/`)

LRU cache with TTL:

```go
type Cache struct {
    entries    map[string]*Entry
    lruList    *list.List
    maxSize    int
    defaultTTL time.Duration
}
```

## Data Flow

```
User Input
    ↓
[Agent] → Parse Request
    ↓
[Orchestrator] → Select Agents, Create Session
    ↓
[LLM] → Generate Response with Tool Calls
    ↓
[Tools] → Execute Actions
    ↓
[Quality Scoring] → Validate Results
    ↓
[Handoff] → Pass to Next Agent (if needed)
    ↓
User Output
```

## Design Principles

### 1. Modularity
Each component is independent and replaceable.

### 2. Safety First
- Trust cascade prevents uncontrolled changes
- Snapshots enable rollback
- Quality gates catch issues early

### 3. Extensibility
- Easy to add new tools
- Skills system for domain knowledge
- Agent registry for specialized capabilities

### 4. Performance
- Caching for repeated operations
- Parallel tool execution where possible
- Efficient handoff protocol

## Configuration

```yaml
# ~/.crush/config.yaml
orchestrator:
  trust_level: L2
  quality_threshold: 60
  
cache:
  enabled: true
  max_size: 1000
  ttl: 15m

skills:
  directories:
    - ~/.crush/skills
    - ./.crush/skills
```

## Extension Points

### Adding a New Tool

1. Implement the `Tool` interface
2. Register in tool registry
3. Add tests

### Adding a New Agent

1. Define in `orchestrator/agent.go`
2. Specify squad, model, expertise
3. Add to registry

### Adding a New Skill

1. Create markdown file with frontmatter
2. Place in skills directory
3. Skill is auto-loaded

## Future Considerations

- Remote MCP server support
- Plugin system for third-party extensions
- Web UI for visual interaction
- Team collaboration features
