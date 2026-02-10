# Nix prefetch pnpm dependencies

[![GitHub License](https://img.shields.io/github/license/cffnpwr/nix-prefetch-pnpm-deps?style=flat)](./LICENSE)

Prefetch pnpm dependencies for nixpkgs

[README.md for English is available here](./README.md).

## How to Install

### Nix (Flakes)

```bash
# 直接実行
nix run github:cffnpwr/nix-prefetch-pnpm-deps

# インストール
nix profile install github:cffnpwr/nix-prefetch-pnpm-deps
```

### Nix Flake Overlay

nixpkgsのoverlayとして使用可能です。
overlayを適用すると`pkgs.nix-prefetch-pnpm-deps`としてパッケージを使用できます。

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
          # pkgs.nix-prefetch-pnpm-deps が利用可能になる
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

[Github Release](https://github.com/cffnpwr/nix-prefetch-pnpm-deps/releases)からダウンロード

### Go install

```bash
go install github.com/cffnpwr/nix-prefetch-pnpm-deps
```

### Build from Source

#### Prerequisites

以下のいずれかの環境を用意してください。

- [Nix](https://nixos.org/) - Nix FlakesをサポートするNix環境
- [mise](https://mise.jdx.dev/) - miseがインストールされている環境
- [go](https://go.dev/) - Go v1.25.5がインストールされている環境

#### How to Build

1. リポジトリをクローン

```bash
git clone https://github.com/cffnpwr/nix-prefetch-pnpm-deps.git
cd nix-prefetch-pnpm-deps
```

2. 開発環境のセットアップ

<details>
<summary>Nixを使用する場合</summary>

```bash
nix develop
```

</details>

<details>
<summary>miseを使用する場合</summary>

```bash
mise install
```

</details>

<details>
<summary>Goを直接使用する場合</summary>

この手順はスキップしてください。

</details>

3. ビルド

```bash
go build .
```

または、Nixを使用する場合：

```bash
nix build
```

4. 実行

```bash
./nix-prefetch-pnpm-deps --help

# Nixでビルドした場合
./result/bin/nix-prefetch-pnpm-deps --help
```

## How to Use

実行には[pnpm](https://pnpm.io/)が必要です。

```bash
nix-prefetch-pnpm-deps [options] <path to pnpm-lock.yaml>
```

## How to setup development environment

開発環境のセットアップは[「Build from Sources」のPrerequisitesセクション](#Prerequisites)を参照してください。

### Pre-commit Hook

このプロジェクトではpre-commit hookに[lefthook](https://github.com/evilmartians/lefthook)を使用しています。
Nix環境では`nix develop`時に自動でインストールされます。

### Linter / Formatter

| ツール | 用途 | 設定ファイル |
|--------|------|--------------|
| [golangci-lint](https://github.com/golangci/golangci-lint) | Go linter | `.golangci.yaml` |
| [gofmt](https://pkg.go.dev/cmd/gofmt) | Go formatter | - |
| [treefmt](https://github.com/numtide/treefmt) | Formatter multiplexer | `treefmt.toml` |

## License

[MIT License](./LICENSE)
