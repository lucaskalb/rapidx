# Makefile - Running CI Locally

This Makefile replicates all GitHub Actions CI steps locally, allowing you to run the same checks before pushing or creating a PR.

## 🚀 Quick Setup

### Option 1: Automatic Script
```bash
./scripts/setup.sh
```

### Option 2: Manual
```bash
make install-tools
make deps
```

## 📋 Main Commands

### Complete CI Pipeline
```bash
make ci                    # Run complete CI pipeline
make pr-validate          # Complete PR validation
make quality              # Only quality checks
```

### Tests
```bash
make test                 # Run all tests
make test-race           # Tests with race condition detection
make test-coverage       # Tests with coverage
make test-demo           # Demonstration tests (expects failures)
make test-examples       # Example tests (expects failures)
```

### Code Checks
```bash
make vet                 # go vet
make fmt                 # Check formatting
make fmt-fix            # Fix formatting
make staticcheck        # staticcheck
make lint               # golangci-lint
make lint-fix           # golangci-lint with fixes
make security           # Security verification (gosec)
```

### Quality and Maintenance
```bash
make tidy               # Check dependencies
make check-license      # Check license headers
make check-todos        # Check TODO/FIXME comments
make check-coverage     # Check minimum coverage (80%)
```

### Utilities
```bash
make build              # Compile the project
make clean              # Clean generated files
make deps               # Download and verify dependencies
make version            # Show Go version
make info               # Project information
```

### Development
```bash
make dev-setup          # Setup complete environment
make watch-test         # Tests in watch mode (requires entr)
```

## 🔧 Replicated Workflows

### CI (ci.yml)
- ✅ Tests on multiple Go versions
- ✅ Functional tests
- ✅ Demonstration tests
- ✅ Example tests
- ✅ Tests with race detection
- ✅ Test coverage
- ✅ Build
- ✅ go vet
- ✅ Formatting verification
- ✅ staticcheck
- ✅ golangci-lint
- ✅ Security verification

### Code Quality (code-quality.yml)
- ✅ go vet
- ✅ Formatting verification
- ✅ staticcheck
- ✅ golangci-lint
- ✅ Security verification
- ✅ License header verification
- ✅ TODO/FIXME verification
- ✅ Dependency verification

### PR Validation (pr-validation.yml)
- ✅ Tests
- ✅ Linting
- ✅ Coverage verification (80% minimum)
- ✅ ADR validation
- ✅ TODO/FIXME verification

## 🎯 Recommended Usage

### Before committing:
```bash
make quality
```

### Before pushing:
```bash
make ci
```

### Before creating PR:
```bash
make pr-validate
```

### During development:
```bash
make watch-test    # Automatic tests when saving files
```

## 📊 Test Coverage

The Makefile checks if coverage is above 80%. To see current coverage:

```bash
make test-coverage
```

To visualize coverage in HTML:
```bash
go tool cover -html=coverage.out
```

## 🛠️ Installed Tools

- **staticcheck**: Advanced static analysis
- **golangci-lint**: Aggregated linter with multiple rules
- **gosec**: Security verification
- **bc**: Calculator for Makefile checks
- **entr**: For watch mode (optional)

## ⚙️ Configuration

### .golangci.yml
golangci-lint configuration that replicates CI rules, including:
- Appropriate exclusions for test files
- Specific configurations for property-based testing
- Complexity thresholds
- Security rules

### Makefile Variables
```makefile
GO_VERSION := 1.24
COVERAGE_THRESHOLD := 80
TIMEOUT := 5m
```

## 🐛 Troubleshooting

### Error: "command not found: golangci-lint"
```bash
make install-tools
```

### Error: "bc: command not found"
```bash
# Ubuntu/Debian
sudo apt-get install bc

# CentOS/RHEL
sudo yum install bc

# macOS
brew install bc
```

### Error: "entr: command not found"
```bash
# Ubuntu/Debian
sudo apt-get install entr

# CentOS/RHEL
sudo yum install entr

# macOS
brew install entr
```

### Low coverage
```bash
make test-coverage
# Check which files need more tests
```

### Unorganized dependencies
```bash
make tidy
```

## 📝 Notes

- The Makefile uses colors for better readability
- All commands are idempotent (can be run multiple times)
- The complete pipeline takes a few minutes to run
- Use `make help` to see all available commands

## 🔗 Useful Links

- [Go Testing](https://golang.org/pkg/testing/)
- [golangci-lint](https://golangci-lint.run/)
- [staticcheck](https://staticcheck.io/)
- [gosec](https://securecodewarrior.github.io/gosec/)
- [GitHub Actions](https://docs.github.com/en/actions)