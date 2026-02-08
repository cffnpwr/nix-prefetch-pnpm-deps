package pnpm

import (
	"fmt"
	"os"
	"os/exec"

	pnpm_err "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/pnpm/errors"
)

// InstallOptions contains options for pnpm install command.
type InstallOptions struct {
	StorePath  string   // store-dir path
	Workspaces []string // --filter flags for workspace filtering
	Registry   string   // --registry flag (from NIX_NPM_REGISTRY)
	ExtraFlags []string // additional flags passed to pnpm install
	WorkingDir string   // directory containing pnpm-lock.yaml
}

// Install runs pnpm install with the specified options.
// It configures pnpm settings and runs install with --force, --ignore-scripts, --frozen-lockfile.
func (p *Pnpm) Install(opts InstallOptions) pnpm_err.PnpmErrorIF {
	// Configure pnpm settings
	configSettings := map[string]string{
		"store-dir":                       opts.StorePath,
		"side-effects-cache":              "false",
		"manage-package-manager-versions": "false",
		"update-notifier":                 "false",
	}

	for key, value := range configSettings {
		if err := p.configSet(key, value, opts.WorkingDir); err != nil {
			return err
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return pnpm_err.NewPnpmError(
			&pnpm_err.FailedToExecuteError{},
			"failed to execute pnpm install",
			err,
		)
	}

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
