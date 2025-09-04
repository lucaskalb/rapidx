#!/bin/bash

# Setup script for development environment
# Installs all tools necessary to run CI locally

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Setting up development environment for rapidx${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go first.${NC}"
    echo "Visit: https://golang.org/doc/install"
    exit 1
fi

echo -e "${GREEN}✅ Go found: $(go version)${NC}"

# Check minimum Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.22"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo -e "${RED}❌ Go version $REQUIRED_VERSION or higher is required. Current version: $GO_VERSION${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Go version is compatible${NC}"

# Install tools
echo -e "${BLUE}📦 Installing development tools...${NC}"

# staticcheck
echo -e "${BLUE}  Installing staticcheck...${NC}"
go install honnef.co/go/tools/cmd/staticcheck@latest

# golangci-lint
echo -e "${BLUE}  Installing golangci-lint...${NC}"
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# gosec
echo -e "${BLUE}  Installing gosec...${NC}"
go install github.com/securego/gosec/v2/cmd/gosec@latest

# bc (for calculations in Makefile)
if ! command -v bc &> /dev/null; then
    echo -e "${YELLOW}⚠️  'bc' not found. Installing...${NC}"
    if command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y bc
    elif command -v yum &> /dev/null; then
        sudo yum install -y bc
    elif command -v brew &> /dev/null; then
        brew install bc
    else
        echo -e "${YELLOW}⚠️  Please install 'bc' manually for your system${NC}"
    fi
fi

# entr (for watch mode, optional)
if ! command -v entr &> /dev/null; then
    echo -e "${YELLOW}⚠️  'entr' not found. To use 'make watch-test', install entr:${NC}"
    if command -v apt-get &> /dev/null; then
        echo "  sudo apt-get install entr"
    elif command -v yum &> /dev/null; then
        echo "  sudo yum install entr"
    elif command -v brew &> /dev/null; then
        echo "  brew install entr"
    fi
fi

echo -e "${GREEN}✅ Tools installed${NC}"

# Download dependencies
echo -e "${BLUE}📥 Downloading dependencies...${NC}"
go mod download
go mod verify

echo -e "${GREEN}✅ Dependencies downloaded and verified${NC}"

# Run initial verification
echo -e "${BLUE}🔍 Running initial verification...${NC}"

# Check formatting
if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
    echo -e "${YELLOW}⚠️  Code is not formatted. Run 'make fmt-fix' to fix.${NC}"
else
    echo -e "${GREEN}✅ Code is properly formatted${NC}"
fi

# Check if dependencies are organized
if ! git diff --exit-code go.mod go.sum 2>/dev/null; then
    echo -e "${YELLOW}⚠️  Dependencies are not organized. Run 'make tidy' to fix.${NC}"
else
    echo -e "${GREEN}✅ Dependencies are organized${NC}"
fi

echo -e "${GREEN}🎉 Setup completed successfully!${NC}"
echo ""
echo -e "${BLUE}Useful commands:${NC}"
echo -e "  ${GREEN}make help${NC}        - Show all available commands"
echo -e "  ${GREEN}make test${NC}        - Run tests"
echo -e "  ${GREEN}make lint${NC}        - Run linting"
echo -e "  ${GREEN}make ci${NC}          - Run complete CI pipeline"
echo -e "  ${GREEN}make pr-validate${NC} - Validate PR"
echo -e "  ${GREEN}make dev-setup${NC}   - Reconfigure environment"
echo ""
echo -e "${BLUE}To run the complete CI pipeline:${NC}"
echo -e "  ${GREEN}make ci${NC}"
echo ""
echo -e "${BLUE}To run only quality checks:${NC}"
echo -e "  ${GREEN}make quality${NC}"