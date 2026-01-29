package cli

import (
	"fmt"

	"github.com/go-extras/cobraflags"
)

const (
	fetcherVersionFlagName = "fetcher-version"
	pnpmPathFlagName       = "pnpm-path"
	workspaceFlagName      = "workspace"
	pnpmFlagFlagName       = "pnpm-flag"
	hashFlagName           = "hash"
	quietFlagName          = "quiet"
)

var (
	fetcherVersionFlag = &cobraflags.IntFlag{
		Name: fetcherVersionFlagName,
		Usage: `pnpm fetcher version
Aviailable versions:
	1: First version. Here to preserve backwards compatibility
	2: Ensure consistent permissions. See https://github.com/NixOS/nixpkgs/pull/422975
	3: Build a reproducible tarball. See https://github.com/NixOS/nixpkgs/pull/469950`,
		Value:    0,
		Required: true,
		ValidateFunc: func(value int) error {
			if value < 1 || value > 3 {
				return fmt.Errorf(
					`"%d" is invalid value for --%s flag. (expected: 1, 2, or 3)`,
					value,
					fetcherVersionFlagName,
				)
			}
			return nil
		},
	}

	pnpmPathFlag = &cobraflags.StringFlag{
		Name:     pnpmPathFlagName,
		Usage:    "path to the pnpm executable",
		Value:    "",
		Required: false,
	}

	workspaceFlag = &cobraflags.StringSliceFlag{
		Name: workspaceFlagName,
		Usage: `filter to restrict to specific workspaces (can be specified multiple times)
if not specified, all workspaces are considered
supports rich selector syntax as described in https://pnpm.io/filtering`,
		Value:    []string{},
		Required: false,
	}

	pnpmFlagFlag = &cobraflags.StringSliceFlag{
		Name:     pnpmFlagFlagName,
		Usage:    "additional flag to pass to pnpm commands (can be specified multiple times)",
		Value:    []string{},
		Required: false,
	}

	hashFlag = &cobraflags.StringFlag{
		Name:     hashFlagName,
		Usage:    "expected hash of fetched dependencies",
		Value:    "",
		Required: false,
	}

	quietFlag = &cobraflags.BoolFlag{
		Name:     quietFlagName,
		Usage:    "suppress non-error output",
		Value:    false,
		Required: false,
	}
)
