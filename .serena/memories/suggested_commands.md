# Suggested Commands

## Development Environment Setup
```bash
# Option 1: Nix Flakes
nix develop

# Option 2: mise
mise install
```

## Build
```bash
go build .
```

## Testing
```bash
# Run all tests
go test ./...

# Run a specific test
go test ./internal/path/ -run "TestIsPath"

# Run tests with verbose output
go test -v ./...
```

## Linting
```bash
# Run golangci-lint (v2, very strict)
golangci-lint run ./...
```

## Formatting
```bash
# Format all files
treefmt

# Format Go files only
gofmt -w .
```

## Pre-commit Hooks
Managed by lefthook. Runs on pre-commit:
- `golangci-lint run ./...`
- `go mod tidy` (auto-stages fixes)
- `treefmt --fail-on-change` (auto-stages fixes)

```bash
# Install hooks
lefthook install
```

## Module Management
```bash
go mod tidy
```

## System Utilities (Darwin/macOS)
```bash
# Standard GNU-compatible utils available via nix develop shell
git, ls, cd, grep, find
```
