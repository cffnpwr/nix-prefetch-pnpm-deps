# Task Completion Checklist

When a coding task is completed, run these checks before considering it done:

## 1. Tests
```bash
go test ./...
```
Ensure all tests pass. If new functionality was added, verify that appropriate tests exist.

## 2. Linting
```bash
golangci-lint run ./...
```
Must pass with zero issues. The configuration is very strict (90+ linters).

## 3. Formatting
```bash
treefmt
```
Ensure code is properly formatted (gofmt for Go, nixfmt for Nix).

## 4. Module Tidying
```bash
go mod tidy
```
Ensure `go.mod` and `go.sum` are clean.

## 5. Build Verification
```bash
go build .
```
Ensure the project builds without errors.

## Notes
- Pre-commit hooks (lefthook) run linting, formatting, and module tidy automatically
- These checks run in parallel during pre-commit
- The `treefmt --fail-on-change` check in pre-commit will fail if files need formatting
