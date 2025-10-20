# Contributing to MIRRA

Thank you for your interest in contributing to MIRRA! This document provides guidelines and workflows for contributing to the project.

## Table of Contents

- [Development Setup](#development-setup)
- [Git Workflow](#git-workflow)
- [Testing Guidelines](#testing-guidelines)
- [Code Style](#code-style)
- [Pull Request Process](#pull-request-process)
- [Reporting Issues](#reporting-issues)

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git
- Make (optional but recommended)
- golangci-lint (optional, for linting)

### Clone and Setup

```bash
# Clone the repository
git clone https://github.com/llmite-ai/mirra.git
cd mirra

# Install dependencies
go mod download

# Install git hooks (recommended)
make install-hooks

# Build the binary
make build

# Run tests
make test
```

## Git Workflow

### Branching Strategy

- `main` - Production-ready code
- Feature branches - Use descriptive names like `feature/add-postgres-support`
- Bug fix branches - Use names like `fix/recording-race-condition`

### Commit Messages

Follow the Conventional Commits specification:

```
<type>: <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks
- `ci`: CI/CD changes

**Examples:**

```
feat: add PostgreSQL backend for recordings

- Implement PostgreSQL recorder
- Add connection pooling
- Update configuration schema

Closes #123
```

```
fix: resolve race condition in async recorder

The recorder's worker goroutine could access the file handle while
it was being closed, causing intermittent panics.

Fixes #456
```

### Git Hooks

MIRRA provides git hooks to maintain code quality:

**Install hooks:**
```bash
make install-hooks
```

**Pre-commit hook:**
- Runs `gofmt` on staged Go files
- Runs `go vet` on changed packages
- Prevents commit if checks fail

**Pre-push hook:**
- Runs full test suite
- Prevents push if tests fail

**Skip hooks temporarily:**
```bash
git commit --no-verify
git push --no-verify
```

## Testing Guidelines

### Writing Tests

MIRRA follows Go testing best practices:

**1. Table-Driven Tests**

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "EmptyString",
            input:    "",
            expected: "",
        },
        {
            name:     "ValidInput",
            input:    "hello",
            expected: "HELLO",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := MyFunction(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

**2. Test File Location**

Place test files next to the code they test:
```
internal/
  config/
    config.go
    config_test.go
  proxy/
    proxy.go
    proxy_test.go
```

**3. Test Naming**

- Test functions: `TestFunctionName_Scenario`
- Subtests: Use descriptive names in `t.Run()`

**4. Test Utilities**

Use helpers from `internal/testutil/`:

```go
func TestConfigLoading(t *testing.T) {
    tempDir, cleanup := testutil.TempDir(t)
    defer cleanup()

    configPath := filepath.Join(tempDir, "config.json")
    testutil.WriteJSONFile(t, configPath, myConfig)

    // Test code...
}
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run tests with race detector
make test-race

# Run tests with coverage
make test-coverage

# Generate HTML coverage report
make coverage-html
```

### Test Coverage Goals

- Aim for 70%+ overall coverage
- Critical paths (proxy routing, recording) should have 80%+ coverage
- Test both success and error cases
- Include edge cases and boundary conditions

## Code Style

### Formatting

- Use `gofmt` for code formatting
- Run `make fmt` before committing
- Configure your editor to run `gofmt` on save

### Linting

```bash
# Run all linters
make lint

# Run go vet
make vet
```

### Code Organization

Follow the project structure documented in `CLAUDE.md`:

- `internal/commands/` - CLI command handlers
- `internal/server/` - HTTP server setup
- `internal/proxy/` - Proxy routing logic
- `internal/recorder/` - Recording persistence
- `internal/config/` - Configuration loading
- `internal/logger/` - Logging utilities
- `internal/testutil/` - Test helpers

### Best Practices

1. **Error Handling**
   ```go
   // Good
   if err != nil {
       return fmt.Errorf("failed to load config: %w", err)
   }

   // Bad
   if err != nil {
       panic(err)
   }
   ```

2. **Logging**
   ```go
   // Use structured logging
   slog.Info("request completed",
       "id", rec.ID[:8],
       "provider", rec.Provider,
       "duration_ms", rec.Timing.DurationMs)
   ```

3. **Concurrency**
   - Use channels for goroutine communication
   - Always use `sync.WaitGroup` for worker goroutines
   - Document goroutine lifecycle in comments

## Pull Request Process

### Before Submitting

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make your changes:**
   - Write code
   - Add/update tests
   - Update documentation

3. **Run quality checks:**
   ```bash
   make fmt
   make vet
   make test
   ```

4. **Commit your changes:**
   ```bash
   git add .
   git commit -m "feat: add my feature"
   ```

5. **Push to your fork:**
   ```bash
   git push origin feature/my-feature
   ```

### Creating the PR

1. **Go to GitHub** and create a pull request

2. **Fill out the PR template:**
   - Title: Use conventional commit format
   - Description: Explain what and why
   - Link related issues

3. **PR checklist:**
   - [ ] Tests pass locally (`make test`)
   - [ ] Code is formatted (`make fmt`)
   - [ ] Linters pass (`make lint`)
   - [ ] Documentation updated
   - [ ] CHANGELOG updated (if applicable)
   - [ ] No sensitive data (API keys, credentials)

### PR Review Process

- Maintainers will review your PR
- Address feedback by pushing new commits
- Once approved, maintainers will merge

### CI/CD Pipeline

Your PR will trigger automated checks:

- **Test Job**: Runs tests on Go 1.21, 1.22, 1.23
- **Lint Job**: Runs gofmt, go vet, and golangci-lint
- **Build Job**: Builds for linux/darwin/windows
- **Coverage Job**: Generates coverage report (must be >70%)

All checks must pass before merging.

## Reporting Issues

### Bug Reports

Include:
- MIRRA version (`./mirra --version` or git commit)
- Go version (`go version`)
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs (with sensitive data redacted)

### Feature Requests

Include:
- Use case description
- Proposed API/interface
- Alternative solutions considered
- Willingness to contribute implementation

## Questions?

- Open an issue for bugs or feature requests
- Start a discussion for general questions
- Check existing issues before creating new ones

## License

By contributing to MIRRA, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to MIRRA! 🎉
