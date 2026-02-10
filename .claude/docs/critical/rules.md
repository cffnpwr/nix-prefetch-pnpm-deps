# Critical Rules

## Filesystem Abstraction

All filesystem operations use `afero.Fs`. Never import or use the `os` package for file I/O.

- Pass `afero.Fs` as the first parameter to constructors and functions that access the filesystem.
- Production code uses `afero.NewOsFs()`.
- Test code uses `afero.NewMemMapFs()`.

```go
// Correct
func Load(fs afero.Fs, path string) (*Lockfile, LockfileErrorIF) {
    f, err := fs.Stat(path)
    // ...
}

// Wrong — never use os directly
func Load(path string) (*Lockfile, error) {
    f, err := os.Stat(path)
    // ...
}
```

## Error Handling

### Return Domain Error Interfaces

Functions return domain error interfaces, not `error`:

- `lockfile` → `lockfile_err.LockfileErrorIF`
- `pnpm` → `pnpm_err.PnpmErrorIF`
- `store` → `store_err.StoreErrorIF`

### Use Factory Functions

Always construct errors via `NewXxxError()`. Never set `Message`/`Cause` fields directly.

```go
// Correct
return nil, lockfile_err.NewLockfileError(
    &lockfile_err.LockfileNotFoundError{},
    "",
    err,
)

// Wrong
return nil, &lockfile_err.LockfileNotFoundError{
    BaseError: common.BaseError{Message: "not found", Cause: err},
}
```

### Error Type Files

Each concrete error type lives in its own file under `errors/`. All implement the domain `ErrorIF` interface with `Error()`, `Is()`, `As()` methods and a compile-time check.

→ Full pattern: `.claude/docs/reference/error-handling.md`

## Import Ordering

Three groups separated by blank lines, enforced by `goimports`:

```go
import (
    "fmt"           // 1. stdlib
    "os"

    "github.com/spf13/afero"  // 2. external

    "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/lockfile"  // 3. local
)
```

Local prefix: `github.com/cffnpwr/nix-prefetch-pnpm-deps`.

Import alias convention for error subpackages: `pkgname_err` (e.g., `lockfile_err`, `pnpm_err`, `store_err`).

## Build Requirements

- **Go version**: 1.25.5
- **CGo**: Required (`CGO_ENABLED=1`). The `store` package links against the C zstd library via `pkg-config: libzstd`.
- **Dev environment**: `nix develop` or `mise install` provides all dependencies including zstd, pkg-config, golangci-lint, lefthook, and treefmt.

## Code Style

- **Linting**: golangci-lint v2 with 90+ linters. Run `golangci-lint run ./...` before committing.
- **Formatting**: `treefmt` formats Go and Nix files. Run `treefmt` before committing.
- **Pre-commit hooks**: lefthook runs lint, `go mod tidy`, and format checks in parallel.
