# Nix prefetch pnpm dependencies

[![GitHub License](https://img.shields.io/github/license/cffnpwr/nix-prefetch-pnpm-deps?style=flat)](./LICENSE)

Prefetch pnpm dependencies for nixpkgs

[日本語版のREADMEはこちら](./README-ja.md)

## How to Install

### Nix (Flakes)

```bash
# Run directly
nix run github:cffnpwr/nix-prefetch-pnpm-deps

# Install
nix profile install github:cffnpwr/nix-prefetch-pnpm-deps
```

### Nix Flake Overlay

This flake can be used as a nixpkgs overlay.
Applying the overlay makes the package available as `pkgs.nix-prefetch-pnpm-deps`.

```nix
# flake.nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    nix-prefetch-pnpm-deps.url = "github:cffnpwr/nix-prefetch-pnpm-deps";
  };

  outputs = { nixpkgs, nix-prefetch-pnpm-deps, ... }: {
    nixosConfigurations.myhost = nixpkgs.lib.nixosSystem {
      modules = [
        {
          nixpkgs.overlays = [ nix-prefetch-pnpm-deps.overlays.default ];
          # pkgs.nix-prefetch-pnpm-deps is now available
        }
      ];
    };
  };
}
```

### Nix (non-Flakes)

```bash
nix-env -if https://github.com/cffnpwr/nix-prefetch-pnpm-deps/archive/main.tar.gz
```

### Github Release

Download from [Github Release](https://github.com/cffnpwr/nix-prefetch-pnpm-deps/releases)

### Go install

```bash
go install github.com/cffnpwr/nix-prefetch-pnpm-deps
```

### Build from Source

#### Prerequisites

Please prepare one of the following environments:

- [Nix](https://nixos.org/) - Nix environment with Nix Flakes support
- [mise](https://mise.jdx.dev/) - Environment with mise installed
- [go](https://go.dev/) - Environment with Go v1.25.5 installed

#### How to Build

1. Clone the repository

```bash
git clone https://github.com/cffnpwr/nix-prefetch-pnpm-deps.git
cd nix-prefetch-pnpm-deps
```

2. Set up the development environment

<details>
<summary>Using Nix</summary>

```bash
nix develop
```

</details>

<details>
<summary>Using mise</summary>

```bash
mise install
```

</details>

<details>
<summary>Using Go directly</summary>

Skip this step.

</details>

3. Build

```bash
go build .
```

Or, if using Nix:

```bash
nix build
```

4. Run

```bash
./nix-prefetch-pnpm-deps --help

# If built with Nix
./result/bin/nix-prefetch-pnpm-deps --help
```

## How to Use

[pnpm](https://pnpm.io/) is required to run this tool.

```bash
nix-prefetch-pnpm-deps [options] <path to pnpm-lock.yaml>
```

## How to setup development environment

For setting up the development environment, please refer to the [Prerequisites section in "Build from Source"](#Prerequisites).

### Pre-commit Hook

This project uses [lefthook](https://github.com/evilmartians/lefthook) for pre-commit hooks.
In the Nix environment, it is automatically installed when running `nix develop`.

### Linter / Formatter

| Tool | Purpose | Config File |
|------|---------|-------------|
| [golangci-lint](https://github.com/golangci/golangci-lint) | Go linter | `.golangci.yaml` |
| [gofmt](https://pkg.go.dev/cmd/gofmt) | Go formatter | - |
| [treefmt](https://github.com/numtide/treefmt) | Formatter multiplexer | `treefmt.toml` |

## License

[MIT License](./LICENSE)