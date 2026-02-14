package pnpm

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/spf13/afero"

	pnpm_err "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/pnpm/errors"
)

// InstallOptions contains options for pnpm install command.
type InstallOptions struct {
	StorePath          string   // store-dir path
	Workspaces         []string // --filter flags for workspace filtering
	Registry           string   // --registry flag (from NIX_NPM_REGISTRY)
	ExtraFlags         []string // additional flags passed to pnpm install
	PreInstallCommands []string // shell commands to run before pnpm install (after config)
	WorkingDir         string   // directory containing pnpm-lock.yaml
}

// Install runs pnpm install with the specified options.
// It configures pnpm settings and runs install with --force, --ignore-scripts, --frozen-lockfile.
//
//nolint:funlen,cyclop // sequential steps (configure → pre-install commands → install) kept together for readability
func (p *Pnpm) Install(fs afero.Fs, opts InstallOptions) pnpm_err.PnpmErrorIF {
	cmdLogger := p.logger.CommandLogger(slog.LevelInfo, "pnpm install")

	// Disable manage-package-manager-versions first, from a temporary directory.
	// If package.json contains a "packageManager" field, pnpm checks it on every command.
	// Running this config set in the source directory would fail because pnpm tries to
	// download the specified version before the config is applied (chicken-and-egg problem).
	// Using a temp directory avoids triggering the packageManager check.
	tmpDir, tmpErr := afero.TempDir(fs, "", "pnpm-config-")
	if tmpErr != nil {
		return pnpm_err.NewPnpmError(
			&pnpm_err.FailedToExecuteError{},
			"failed to create temp directory for pnpm config",
			tmpErr,
		)
	}
	defer func() { _ = fs.RemoveAll(tmpDir) }()

	if err := p.configSet("manage-package-manager-versions", "false", tmpDir); err != nil {
		return err
	}

	// Configure remaining pnpm settings in the source directory
	configSettings := map[string]string{
		"store-dir":          opts.StorePath,
		"side-effects-cache": "false",
		"update-notifier":    "false",
	}

	for key, value := range configSettings {
		if err := p.configSet(key, value, opts.WorkingDir); err != nil {
			return err
		}
	}

	// Run pre-install commands (equivalent to nixpkgs' prePnpmInstall)
	for _, command := range opts.PreInstallCommands {
		cmd := exec.Command("sh", "-c", command)
		cmd.Dir = opts.WorkingDir
		cmd.Stdout = cmdLogger
		cmd.Stderr = cmdLogger

		err := cmd.Run()
		if err != nil {
			return pnpm_err.NewPnpmError(
				&pnpm_err.FailedToExecuteError{},
				"failed to execute pre-install command: "+command,
				err,
			)
		}
	}

	// Build pnpm install arguments
	args := []string{
		"install",
		"--force",
		"--ignore-scripts",
		"--frozen-lockfile",
	}

	// Add registry if specified
	if opts.Registry != "" {
		args = append(args, "--registry="+opts.Registry)
	}

	// Add workspace filters
	for _, ws := range opts.Workspaces {
		args = append(args, "--filter="+ws)
	}

	// Add extra flags
	args = append(args, opts.ExtraFlags...)

	// Run pnpm install
	cmd := exec.Command(p.path, args...)
	cmd.Dir = opts.WorkingDir
	cmd.Stdout = cmdLogger
	cmd.Stderr = cmdLogger

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			cmdLogger.Fail(exitErr.ExitCode())
		} else {
			cmdLogger.Fail(-1)
		}

		return pnpm_err.NewPnpmError(
			&pnpm_err.FailedToExecuteError{},
			"failed to execute pnpm install",
			err,
		)
	}

	cmdLogger.Done()

	return nil
}

// configSet runs pnpm config set <key> <value>.
func (p *Pnpm) configSet(key, value, workingDir string) pnpm_err.PnpmErrorIF {
	cmd := exec.Command(p.path, "config", "set", key, value)
	cmd.Dir = workingDir

	if err := cmd.Run(); err != nil {
		return pnpm_err.NewPnpmError(
			&pnpm_err.FailedToExecuteError{},
			fmt.Sprintf("failed to set pnpm config %s=%s", key, value),
			err,
		)
	}

	return nil
}
