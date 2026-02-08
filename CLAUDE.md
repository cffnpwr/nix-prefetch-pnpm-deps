# CLAUDE.md

## Project Overview

A CLI tool written in Go that prefetches pnpm dependencies for nixpkgs. It reads `pnpm-lock.yaml` and produces dependency hashes for Nix builds. Supports pnpm fetcher versions 1-3.

## Commands

```bash
# Build
go build .

# Run tests (all)
go test ./...

# Run a single test
go test ./internal/path/ -run "TestIsPath"

# Lint (golangci-lint v2, very strict config with 90+ linters)
golangci-lint run ./...

# Format
treefmt

# Dev environment setup (choose one)
nix develop          # Nix Flakes
mise install         # mise
```

Pre-commit hooks via lefthook run `golangci-lint run`, `go mod tidy`, and `treefmt --fail-on-change` in parallel.

## Architecture

Entry point: `main.go` → `cli.Execute()`. All application code lives under `internal/` (unexported):

```
internal/
├── cli/       — CLI layer (cobra), flags, root command
├── common/    — BaseError, MajorVersion semver parser
├── lockfile/  — pnpm-lock.yaml parsing (Load, Parse)
│   └── errors/
├── path/      — Path string validation (OS-specific build tags)
└── pnpm/      — pnpm execution (New, WithPathEnvVar, Install)
    └── errors/
```

→ Details: `.claude/docs/reference/architecture.md`

## Code Conventions

- **Filesystem abstraction**: `afero.Fs` is passed to constructors/functions for testability (in-memory FS in tests, real OS FS in production).
- **Import ordering**: `goimports` with local prefix `github.com/cffnpwr/nix-prefetch-pnpm-deps` (stdlib → external → local).
- **Go version**: 1.25.5.

## Task Navigation

### Error Handling
**When**: Adding or modifying error types in `lockfile/errors/` or `pnpm/errors/`, or creating a new domain error hierarchy.

→ Reference: `.claude/docs/reference/error-handling.md`
→ Skill: `implement-error-handling`

### Testing
**When**: Writing tests, adding test cases, or reviewing test patterns.

→ Reference: `.claude/docs/reference/testing.md`
→ Skill: `write-tests`

### Architecture
**When**: Understanding package responsibilities, function signatures, or relationships between packages.

→ Reference: `.claude/docs/reference/architecture.md`

## Available Skills

- `implement-error-handling` — Workflow for implementing error types in the domain error hierarchy
- `write-tests` — Workflow for writing tests (table-driven, parallel, afero, go-cmp, Japanese naming)
