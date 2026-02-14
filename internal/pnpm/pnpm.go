package pnpm

import (
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/spf13/afero"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/logger"
	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/path"
	pnpm_err "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/pnpm/errors"
)

type config struct {
	Paths []string `env:"PATH" envSeparator:":"`
}

type Pnpm struct {
	fs     afero.Fs
	logger logger.Logger
	path   string
}

func New(fs afero.Fs, logger logger.Logger, path string) (*Pnpm, pnpm_err.PnpmErrorIF) {
	if err := validatePnpmExecutable(fs, path); err != nil {
		return nil, err
	}
	logger.Debugf("found pnpm executable at path: %s", path)

	return &Pnpm{fs, logger, path}, nil
}

// WithPathEnvVar searches for the pnpm executable in the PATH environment variable
// and returns a Pnpm instance if found.
// If not found, it returns a PnpmNotFoundError.
func WithPathEnvVar(fs afero.Fs, logger logger.Logger) (*Pnpm, pnpm_err.PnpmErrorIF) {
	var cfg config

	// Parse PATH environment variable
	err := env.Parse(&cfg)
	if err != nil {
		return nil, pnpm_err.NewPnpmError(
			&pnpm_err.OtherError{},
			"failed to parse environment variables",
			err,
		)
	}
	logger.Debugf("parsed PATH environment variable: %v", cfg.Paths)

	// Search executable pnpm in each path in PATH environment variable
	for _, p := range cfg.Paths {
		logger.Debugf("checking for pnpm executable in path: %s", p)

		// Check if p is a valid path
		if !path.IsPath(p) {
			// Skip invalid path
			logger.Debugf("skipping invalid path: %s", p)
			continue
		}

		pnpmPath := p + string(os.PathSeparator) + "pnpm"
		err := validatePnpmExecutable(fs, pnpmPath)
		if err == nil {
			// Found valid pnpm executable
			logger.Debugf("found pnpm executable at path: %s", pnpmPath)
			return &Pnpm{fs, logger, pnpmPath}, nil
		}

		logger.Debugf("pnpm executable not found at path: %s", pnpmPath)
	}

	// If pnpm executable is not found in any path, return an error
	logger.Debugf("pnpm executable not found in any of the paths in PATH environment variable")
	return nil, pnpm_err.NewPnpmError(&pnpm_err.PnpmNotFoundError{}, "", nil)
}

func (p *Pnpm) Path() string {
	return p.path
}

func validatePnpmExecutable(fs afero.Fs, path string) pnpm_err.PnpmErrorIF {
	f, err := fs.Stat(path)
	if err != nil {
		return pnpm_err.NewPnpmError(
			&pnpm_err.PnpmNotFoundError{},
			"pnpm executable not found at path: "+path,
			err,
		)
	}

	if !f.Mode().IsRegular() || (f.Mode().Perm()&0111 == 0) {
		return pnpm_err.NewPnpmError(
			&pnpm_err.PnpmNotFoundError{},
			"pnpm executable is not a executable file at path: "+path,
			nil,
		)
	}

	return nil
}
