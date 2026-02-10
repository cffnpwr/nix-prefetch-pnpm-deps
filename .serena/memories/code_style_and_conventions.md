# Code Style and Conventions

## Go Conventions
- **Go version**: 1.25.5
- **Indentation**: Tabs for Go files, 2 spaces for everything else
- **Import ordering**: goimports with local prefix `github.com/cffnpwr/nix-prefetch-pnpm-deps`
  - Order: stdlib → external packages → local packages
- **Formatting**: gofmt via treefmt
- **Line endings**: LF
- **Charset**: UTF-8
- **Trailing whitespace**: Trimmed
- **Final newline**: Required

## Filesystem Abstraction
- `afero.Fs` is passed to constructors/functions for testability
- In-memory FS (`afero.NewMemMapFs()`) in tests, real OS FS (`afero.NewOsFs()`) in production
- Never use `os` package directly for file operations in application code

## Error Handling Pattern
- Domain-specific error hierarchies per package (see `error-handling.md` memory)
- All domain errors embed `common.BaseError`
- Each domain has an error interface (e.g., `LockfileErrorIF`, `PnpmErrorIF`)
- Factory function `NewXxxError()` for creating errors
- Return types use domain error interface, not `error`
- Each concrete error type in its own file
- Compile-time interface checks: `var _ IF = (*Concrete)(nil)`

## Testing Conventions
- Table-driven tests with `struct` slices
- `t.Parallel()` on both test function and each subtest
- Test case names in **Japanese** with prefixes:
  - `[正常系]` for normal/positive cases
  - `[異常系]` for error/negative cases
- Assertions: `go-cmp` for struct comparison, `reflect.TypeOf` for error types
- Filesystem tests use `setupFs` callback pattern with afero
- Error subpackage imports aliased as `pkgname_err` (e.g., `lockfile_err`)

## Naming Conventions
- Error type names: suffix with `Error` (e.g., `LockfileNotFoundError`)
- Error interface names: suffix with `ErrorIF` (e.g., `LockfileErrorIF`)
- Error sentinel pattern NOT used — typed errors via interfaces instead
- Test functions: `Test_FunctionName` (underscore convention)

## Linting
- golangci-lint v2 with 90+ linters enabled
- Very strict configuration
- Some exclusions for test files (bodyclose, dupl, errcheck, funlen, goconst, gosec, noctx, wrapcheck)
- Package comments not required (revive exclusion)
- gosec G204 and noctx excluded for `internal/pnpm/` (intentional external command execution)
