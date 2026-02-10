# Project Overview

## Purpose
`nix-prefetch-pnpm-deps` is a CLI tool written in Go that prefetches pnpm dependencies for nixpkgs. It reads `pnpm-lock.yaml` and produces dependency hashes (NAR hashes) for Nix builds. Supports pnpm fetcher versions 1-3.

## Tech Stack
- **Language**: Go 1.25.5
- **CLI framework**: cobra (`github.com/spf13/cobra`)
- **Filesystem abstraction**: afero (`github.com/spf13/afero`) — enables testability via in-memory FS
- **YAML parsing**: `go.yaml.in/yaml/v3` and `go.yaml.in/yaml/v4`
- **NAR hashing**: `github.com/nix-community/go-nix`
- **Build system**: Nix Flakes (`flake.nix`)
- **Tool management**: mise (alternative to nix develop)
- **Pre-commit hooks**: lefthook
- **Linting**: golangci-lint v2 (very strict, 90+ linters enabled)
- **Formatting**: treefmt (gofmt for Go, nixfmt for Nix)
- **Flags library**: `github.com/go-extras/cobraflags`
- **Testing assertions**: `github.com/google/go-cmp`
- **Config/env**: `github.com/caarlos0/env/v11`

## Platform
- Targets: x86_64-linux, aarch64-linux, aarch64-darwin, x86_64-darwin
- Requires CGO_ENABLED=1
- System utilities use macOS (Darwin) variants
