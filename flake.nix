{
  description = "Prefetch dependencies for pnpm";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    inputs@{ flake-parts, self, ... }:
    let
      mkCommonPackageAttrs =
        { pkgs, src }:
        {
          pname = "nix-prefetch-pnpm-deps";
          version = "1.0.0";

          inherit src;

          vendorHash = "sha256-yMIimiDH8J6iNTeTAAANT7frTBbgZm9o05Ga8VjdGgg=";

          env.CGO_ENABLED = "1";
          nativeBuildInputs = [ pkgs.pkg-config ];

          meta = {
            description = "Prefetch pnpm dependencies for Nix builds";
            homepage = "https://github.com/cffnpwr/nix-prefetch-pnpm-deps";
            license = pkgs.lib.licenses.mit;
            mainProgram = "nix-prefetch-pnpm-deps";
          };
        };
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
      ];

      flake = {
        overlays.default = final: prev: {
          nix-prefetch-pnpm-deps = final.buildGoModule (
            mkCommonPackageAttrs {
              pkgs = final;
              src = final.lib.cleanSource self;
            }
            // {
              buildInputs = [ final.zstd ];
            }
          );
        };
      };

      perSystem =
        {
          config,
          self',
          inputs',
          pkgs,
          lib,
          system,
          ...
        }:
        let
          commonPackageAttrs = mkCommonPackageAttrs {
            inherit pkgs;
            src = lib.cleanSource ./.;
          };
        in
        {
          packages = {
            default = pkgs.buildGoModule (
              commonPackageAttrs
              // {
                buildInputs = [ pkgs.zstd ];
              }
            );

            static = pkgs.buildGoModule (
              commonPackageAttrs
              // {
                buildInputs = [ (pkgs.zstd.override { static = true; }) ];
                tags = [
                  "netgo"
                  "osusergo"
                ];
                postFixup = lib.optionalString pkgs.stdenv.hostPlatform.isDarwin ''
                  install_name_tool -change \
                    ${pkgs.darwin.libresolv}/lib/libresolv.9.dylib \
                    /usr/lib/libresolv.9.dylib \
                    $out/bin/nix-prefetch-pnpm-deps
                '';
              }
            );
          };

          devShells.default = pkgs.mkShell {
            env.CGO_ENABLED = "1";

            buildInputs = with pkgs; [
              git
              go
              golangci-lint
              gopls
              lefthook
              nil
              nixd
              nixfmt
              pkg-config
              treefmt
              zstd
            ];

            shellHook = ''
              lefthook install

              # Only exec into user shell for interactive sessions
              # Skip for non-interactive commands (like VSCode env detection)
              if [ -t 0 ] && [ -z "$__NIX_SHELL_EXEC" ]; then
                export __NIX_SHELL_EXEC=1

                # Detect user's login shell (works on both macOS and Linux)
                if command -v dscl >/dev/null 2>&1; then
                  # macOS
                  USER_SHELL=$(dscl . -read ~/ UserShell | sed 's/UserShell: //')
                elif command -v getent >/dev/null 2>&1; then
                  # Linux
                  USER_SHELL=$(getent passwd $USER | cut -d: -f7)
                else
                  # Fallback: read /etc/passwd directly
                  USER_SHELL=$(grep "^$USER:" /etc/passwd | cut -d: -f7)
                fi

                exec ''${USER_SHELL:-$SHELL}
              fi
            '';
          };
        };
    };
}
