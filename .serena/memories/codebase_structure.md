# Codebase Structure

## Entry Point
`main.go` → calls `cli.Execute()`. All application code under `internal/` (unexported).

## Package Layout
```
main.go              — Entry point
internal/
├── cli/             — CLI layer (cobra), flags, root command
│   ├── root.go      — Root command definition
│   └── flags.go     — Flag definitions via cobraflags
├── common/          — Shared utilities
│   ├── errors.go    — BaseError struct (base for all domain errors)
│   └── semver.go    — MajorVersion semver parser
├── lockfile/        — pnpm-lock.yaml parsing
│   ├── parse.go     — Load() and Parse() functions
│   └── errors/      — Domain errors (LockfileErrorIF interface)
│       ├── main.go
│       ├── not_found.go
│       ├── failed_to_load.go
│       └── failed_to_parse.go
├── path/            — Path string validation (OS-specific build tags)
│   ├── main.go
│   ├── main_unix.go
│   └── main_windows.go
├── pnpm/            — pnpm execution
│   ├── pnpm.go      — New(), WithPathEnvVar() constructors
│   ├── install.go   — Install() method
│   ├── version.go   — Version detection
│   └── errors/      — Domain errors (PnpmErrorIF interface)
│       ├── main.go
│       ├── not_found.go
│       ├── failed_to_execute.go
│       ├── failed_to_parse.go
│       └── other.go
└── store/           — NAR hash computation and store management
    ├── hash.go      — Hash computation
    ├── normalize.go — JSON normalization for pnpm dependencies
    ├── tarball.go   — Tarball creation (fetcher v3)
    ├── tarwriter.go — Tar archive writer
    ├── zstdwriter.go — Zstd compression writer
    └── errors/      — Domain errors
        ├── main.go
        ├── failed_to_cleanup.go
        ├── failed_to_create_tarball.go
        ├── failed_to_hash.go
        ├── failed_to_normalize_json.go
        └── failed_to_set_permissions.go
test/
└── integration/     — Integration tests
```

## CLI Flags
- `--fetcher-version` (required, 1-3)
- `--pnpm-path`
- `--workspace` (repeatable)
- `--pnpm-flag` (repeatable)
- `--hash`
- `--quiet`
