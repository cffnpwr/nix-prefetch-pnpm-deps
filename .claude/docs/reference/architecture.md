# Architecture Reference

## Entry Point

`main.go` calls `cli.Execute()`. All application code lives under `internal/` (unexported).

## Package Details

### `cli/`

CLI layer using cobra.

- Root command requires exactly 1 arg (path to `pnpm-lock.yaml`)
- Flags defined in `flags.go` via `cobraflags` package:
  - `--fetcher-version` (required, 1-3)
  - `--pnpm-path`
  - `--workspace` (repeatable)
  - `--pnpm-flag` (repeatable)
  - `--hash`
  - `--quiet`

### `common/`

Shared utilities:

- `BaseError` — Base error struct with `Message`/`Cause`/`Unwrap`. See `.claude/docs/reference/error-handling.md` for full pattern.
- `MajorVersion` — Semver parser that extracts major version number.

### `lockfile/`

Parses `pnpm-lock.yaml` files.

- `Load(fs afero.Fs, path string)` — Reads lockfile from filesystem. Returns `(*Lockfile, LockfileErrorIF)`.
- `Parse(data []byte)` — Unmarshals YAML into `Lockfile` struct.
- Uses `afero.Fs` for filesystem abstraction.
- Has its own `errors/` subpackage with `LockfileErrorIF` interface.

### `path/`

Path string validation.

- OS-specific implementations via build tags (`//go:build unix` / `//go:build windows`).
- `IsPath(s string)` — Validates whether a string is a valid path.

### `pnpm/`

Controls pnpm execution.

- `New(fs afero.Fs, path string)` — Constructor with explicit path.
- `WithPathEnvVar(fs afero.Fs)` — Constructor that finds pnpm from `PATH`.
- `Install(opts InstallOpts)` — Configures pnpm settings then runs install with `--force --ignore-scripts --frozen-lockfile`.
- Uses `afero.Fs` for filesystem abstraction.
- Has its own `errors/` subpackage with `PnpmErrorIF` interface.

### `store/`

Processes the pnpm store for reproducible hashing and packaging.

- `Normalize(afs afero.Fs, opts NormalizeOptions)` — Normalizes store for reproducible hashing (removes tmp/projects dirs, normalizes JSON, sets permissions for v2+). `NormalizeOptions` has `StorePath string` and `FetcherVersion int`.
- `Hash(afs afero.Fs, storePath string)` — Computes NAR hash in SRI format (`sha256-<base64>`) using `go-nix`.
- `CreateTarball(afs afero.Fs, storePath string, outputPath string)` — Creates reproducible zstd-compressed tarball (fetcher v3+). Byte-identical to `tar --sort=name --mtime="@315532800" --owner=0 --group=0 --numeric-owner --zstd`.
- Internal: `gnuTarWriter` (GNU tar PAX format), `zstdWriter` (CGo wrapper for C zstd library, level 3, content checksum).
- Uses `afero.Fs` for filesystem abstraction.
- Has its own `errors/` subpackage with `StoreErrorIF` interface.
