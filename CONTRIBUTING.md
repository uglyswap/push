# Contributing to Crush

Thank you for your interest in contributing to Crush! This document provides guidelines and information for contributors.

## Code of Conduct

Please be respectful and constructive in all interactions. We're building something great together.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- An API key for Anthropic or OpenAI

### Development Setup

```bash
# Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/crush.git
cd crush

# Create a branch for your changes
git checkout -b feature/your-feature-name

# Install dependencies
go mod download

# Run tests to make sure everything works
go test ./...
```

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in Issues
2. If not, create a new issue with:
   - A clear, descriptive title
   - Steps to reproduce the bug
   - Expected behavior
   - Actual behavior
   - Your environment (OS, Go version, etc.)

### Suggesting Features

1. Check if the feature has already been suggested
2. Create a new issue with:
   - A clear description of the feature
   - Why it would be useful
   - Possible implementation approach

### Submitting Code

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Write or update tests
5. Ensure all tests pass
6. Submit a pull request

## Code Standards

### Go Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `go vet` before committing
- Add comments for exported functions and types

### Commit Messages

Use conventional commits:

```
feat: add new feature
fix: fix a bug
docs: update documentation
test: add or update tests
refactor: code refactoring
chore: maintenance tasks
```

### Testing

- Write tests for new functionality
- Maintain or improve code coverage
- Test edge cases

### Documentation

- Update README.md if needed
- Add comments for complex logic
- Update CHANGELOG.md for significant changes

## Project Structure

```
crush/
├── cmd/crush/          # Main application
├── internal/           # Internal packages
│   ├── agent/          # Agent implementation
│   │   └── tools/      # Tool implementations
│   ├── orchestrator/   # Multi-agent orchestration
│   ├── planmode/       # Plan mode system
│   ├── skills/         # Skills system
│   ├── lsp/            # LSP integration
│   └── cache/          # Caching system
├── docs/               # Documentation
└── tests/              # Integration tests
```

## Adding New Tools

1. Create a new file in `internal/agent/tools/`
2. Implement the tool interface:
   - `Name() string`
   - `Description() string`
   - `Parameters() map[string]interface{}`
   - `Execute(ctx context.Context, params json.RawMessage) (string, error)`
   - `RequiresApproval() bool`
3. Register the tool in the agent
4. Add tests
5. Update documentation

## Adding New Agents

1. Add the agent definition in `internal/orchestrator/agent.go`
2. Define the agent's:
   - Name and description
   - Squad assignment
   - Model preference (sonnet/opus/haiku)
   - Expertise keywords
   - Collaboration patterns
3. Update the agent registry
4. Add tests

## Adding New Skills

1. Create a markdown file in the appropriate skills directory
2. Add YAML frontmatter with:
   - `name`
   - `description`
   - `allowed-tools` (optional)
3. Write the skill content
4. Test the skill

## Pull Request Process

1. Ensure your code follows the style guidelines
2. Update documentation as needed
3. Add tests for new functionality
4. Ensure all tests pass
5. Request review from maintainers
6. Address review feedback
7. Once approved, your PR will be merged

## Questions?

Open an issue or reach out to the maintainers.

Thank you for contributing! 🎉
