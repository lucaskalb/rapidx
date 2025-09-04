# GitHub Actions Workflows

This directory contains GitHub Actions workflows for the RapidX project.

## Workflows

### 🔄 CI (`ci.yml`)
**Triggers:** Push to `main`/`develop`, Pull Requests to `main`/`develop`

**Jobs:**
- **Test**: Runs tests across Go 1.21, 1.22, and 1.23
  - Functional tests (gen, prop, quick packages)
  - Demonstration tests (expected to fail)
  - Examples tests (expected to fail)
  - Race condition detection
  - Coverage reporting
- **Build**: Compiles the project and examples
- **Lint**: Runs golangci-lint for code quality
- **Security**: Scans for security vulnerabilities

### 📋 PR Validation (`pr-validation.yml`)
**Triggers:** Pull Request events

**Features:**
- Validates PR title follows conventional commits format
- Checks for breaking changes
- Runs tests and linting
- Validates test coverage (minimum 80%)
- Checks ADR changes
- Warns about TODO/FIXME comments

### 🚀 Release (`release.yml`)
**Triggers:** Git tags (`v*`), Manual dispatch

**Features:**
- Runs full test suite
- Builds the project
- Creates GitHub releases with changelog
- Generates installation instructions

### 🔄 Dependencies (`dependencies.yml`)
**Triggers:** Weekly schedule (Mondays), Manual dispatch

**Features:**
- Checks for outdated dependencies
- Updates dependencies automatically
- Creates PR with updates
- Runs tests with updated dependencies

### 📚 Documentation (`docs.yml`)
**Triggers:** Push/PR to `main`/`develop` (when docs change)

**Features:**
- Validates markdown formatting
- Checks ADR format compliance
- Validates Go code examples in documentation
- Checks for broken internal links

### 🔍 Code Quality (`code-quality.yml`)
**Triggers:** Push/PR to `main`/`develop`

**Features:**
- Runs `go vet`
- Checks code formatting
- Static analysis with staticcheck
- Security scanning with gosec
- License header validation
- Dependency tidiness check

## Configuration Files

### `.golangci.yml`
Configuration for golangci-lint with:
- Comprehensive linter rules
- Exclusions for test files and domain-specific code
- Custom thresholds and settings

### `codecov.yml`
Code coverage configuration:
- 80% coverage target
- Patch and project coverage thresholds
- Exclusions for documentation and test files

## Usage

### Running Workflows Locally

While you can't run GitHub Actions locally without additional tools, you can test the commands:

```bash
# Test functional code
go test -v ./gen/... ./quick/...

# Check formatting
gofmt -s -l .

# Run linting (requires golangci-lint)
golangci-lint run

# Check security (requires gosec)
gosec ./...

# Run static analysis (requires staticcheck)
staticcheck ./...
```

### Workflow Status

All workflows are designed to:
- ✅ Pass on clean code
- ❌ Fail on issues that need attention
- ⚠️ Warn about potential problems

### Troubleshooting

**Common Issues:**

1. **Tests failing**: Check if they're demonstration tests (expected to fail)
2. **Formatting issues**: Run `go fmt ./...` to fix
3. **Linting errors**: Review `.golangci.yml` configuration
4. **Coverage below 80%**: Add more tests or adjust thresholds

**Getting Help:**

- Check workflow logs in GitHub Actions tab
- Review configuration files in this directory
- Consult Go documentation for specific tool issues

## Contributing

When adding new workflows:

1. Follow the existing naming conventions
2. Include proper triggers and conditions
3. Add documentation to this README
4. Test workflows thoroughly
5. Consider security implications

## Security

- All workflows use official actions when possible
- Dependencies are pinned to specific versions
- Secrets are properly scoped and used
- No sensitive data is logged or exposed