# Development Guide

This guide explains how to use the Makefile to run the same CI steps locally during development.

## 🚀 Initial Setup

### 1. Configure Environment
```bash
# Option 1: Automatic script
./scripts/setup.sh

# Option 2: Manual
make dev-setup
```

### 2. Verify Setup
```bash
make info
```

## 🔄 Development Workflow

### During Development

#### 1. Before starting to code
```bash
make deps          # Download updated dependencies
make tidy          # Organize dependencies
```

#### 2. During coding
```bash
# For quick tests
make test

# To check formatting
make fmt

# To check basic issues
make vet
```

#### 3. Watch Mode (optional)
```bash
# Install entr first
sudo apt-get install entr  # Ubuntu/Debian
# or
brew install entr          # macOS

# Run tests automatically when saving
make watch-test
```

### Before Committing

#### 1. Basic Checks
```bash
make quality
```

This command runs:
- `go vet` - Basic checks
- `go fmt` - Formatting verification
- `staticcheck` - Static analysis
- `golangci-lint` - Complete linting
- `gosec` - Security verification
- License header verification
- TODO/FIXME verification

#### 2. If there are problems
```bash
make fmt-fix       # Fix formatting
make lint-fix      # Fix linting issues automatically
```

### Before Pushing

#### 1. Complete Pipeline
```bash
make ci
```

This command runs the complete CI pipeline:
- Install tools
- Organize dependencies
- Run code checks
- Run tests
- Run tests with race detection
- Generate test coverage
- Compile the project

### Before Creating PR

#### 1. Complete PR Validation
```bash
make pr-validate
```

This command runs:
- All quality checks
- Tests with coverage
- Minimum coverage verification (80%)
- License header verification
- TODO/FIXME verification
- ADR change verification

## 🐛 Troubleshooting

### Common Problems

#### 1. Incorrect Formatting
```bash
make fmt-fix
```

#### 2. Linting Issues
```bash
make lint-fix
```

#### 3. Unorganized Dependencies
```bash
make tidy
```

#### 4. Low Coverage
```bash
make test-coverage
go tool cover -html=coverage.out  # Visualize in HTML
```

#### 5. Failing Tests
```bash
make test -v  # Verbose to see details
```

### Debug Commands

#### 1. Project Information
```bash
make info
```

#### 2. Go Version
```bash
make version
```

#### 3. Clean Generated Files
```bash
make clean
```

## 📊 Quality Monitoring

### Test Coverage
```bash
# See current coverage
make test-coverage

# Check if it's above threshold (80%)
make check-coverage

# Visualize coverage in HTML
go tool cover -html=coverage.out
```

### Code Analysis
```bash
# Static analysis
make staticcheck

# Complete linting
make lint

# Security verification
make security
```

## 🔧 Advanced Configuration

### Customize Thresholds
Edit the Makefile to change:
```makefile
COVERAGE_THRESHOLD := 80  # Change to your desired value
TIMEOUT := 5m            # Timeout for golangci-lint
```

### Configure Linting
Edit `.golangci.yml` to customize linting rules.

### Add New Commands
Add new targets to the Makefile following the existing pattern.

## 📝 Tips and Best Practices

### 1. Run Checks Frequently
```bash
# During development
make test

# Before commit
make quality

# Before push
make ci
```

### 2. Use Watch Mode for Development
```bash
make watch-test  # Automatic tests
```

### 3. Maintain High Coverage
- Run `make test-coverage` regularly
- Add tests for new code
- Keep coverage above 80%

### 4. Resolve Issues Immediately
- Don't let linting issues accumulate
- Use `make lint-fix` for automatic fixes
- Resolve TODOs/FIXMEs before merging

### 5. Use Specific Commands
```bash
# For specific problems
make vet           # Only go vet
make fmt           # Only formatting verification
make staticcheck   # Only staticcheck
make security      # Only security verification
```

## 🚨 Troubleshooting

### Error: "command not found"
```bash
make install-tools
```

### Error: "bc: command not found"
```bash
# Ubuntu/Debian
sudo apt-get install bc

# macOS
brew install bc
```

### Error: "entr: command not found"
```bash
# Ubuntu/Debian
sudo apt-get install entr

# macOS
brew install entr
```

### Timeout in golangci-lint
Increase timeout in Makefile:
```makefile
TIMEOUT := 10m  # Increase from 5m to 10m
```

### Memory Issues
```bash
# Clean Go cache
go clean -cache

# Clean generated files
make clean
```

## 📚 Additional Resources

- [MAKEFILE.md](./MAKEFILE.md) - Complete Makefile documentation
- [GitHub Actions Workflows](../.github/workflows/) - Original CI workflows
- [Go Testing](https://golang.org/pkg/testing/) - Official testing documentation
- [golangci-lint](https://golangci-lint.run/) - Linter documentation