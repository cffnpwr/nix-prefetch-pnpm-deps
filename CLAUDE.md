# CLAUDE.md

## Project Overview

A CLI tool written in Go that prefetches pnpm dependencies for nixpkgs. It reads `pnpm-lock.yaml` and produces dependency hashes for Nix builds. Supports pnpm fetcher versions 1-3. All application code lives under `internal/` (unexported).

## Critical Rules

- **Always use `afero.Fs`** for filesystem operations. Never use `os` package directly. Pass `afero.Fs` to constructors/functions.
- **Return domain error interfaces** (`LockfileErrorIF`, `PnpmErrorIF`, `StoreErrorIF`), not `error`. Use `NewXxxError()` factory functions.
- **Import ordering**: stdlib → external → local (`github.com/cffnpwr/nix-prefetch-pnpm-deps`). Enforced by `goimports`.
- **CGo required**: `store` package uses CGo for zstd (`CGO_ENABLED=1`, `pkg-config: libzstd`). The Nix dev environment provides these.
- **Go version**: 1.25.5.

→ Detailed rules: `.claude/docs/critical/rules.md`

## Commands

```bash
go build .                              # Build
go test ./...                           # Test (all)
go test ./internal/path/ -run "TestIsPath"  # Test (single)
golangci-lint run ./...                 # Lint (v2, 90+ linters)
treefmt                                 # Format
nix develop                             # Dev environment (Nix Flakes)
mise install                            # Dev environment (mise)
```

Pre-commit hooks via lefthook run `golangci-lint run`, `go mod tidy`, and `treefmt --fail-on-change` in parallel.

## Architecture

```
internal/
├── cli/       — CLI layer (cobra), flags, root command
├── common/    — BaseError, MajorVersion semver parser
├── lockfile/  — pnpm-lock.yaml parsing (Load, Parse)
│   └── errors/
├── path/      — Path string validation (OS-specific build tags)
├── pnpm/      — pnpm execution (New, WithPathEnvVar, Install)
│   └── errors/
└── store/     — Store normalization, NAR hashing, tarball creation (fetcher v3+)
    └── errors/
```

→ Details: `.claude/docs/reference/architecture.md`

## Task Navigation

### Error Handling
**When**: Adding or modifying error types in `lockfile/errors/`, `pnpm/errors/`, or `store/errors/`, or creating a new domain error hierarchy.

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
