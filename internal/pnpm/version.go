package pnpm

import (
	"os/exec"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/common"
	pnpm_err "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/pnpm/errors"
)

func (p *Pnpm) Version() (string, pnpm_err.PnpmErrorIF) {
	cmd := exec.Command(p.path, "--version")

	o, err := cmd.Output()
	if err != nil {
		return "", pnpm_err.NewPnpmError(
			&pnpm_err.FailedToExecuteError{},
			"failed to execute pnpm to get version",
			err,
		)
	}

	return string(o), nil
}

func (p *Pnpm) MajorVersion() (int, pnpm_err.PnpmErrorIF) {
	versionStr, err := p.Version()
	if err != nil {
		return 0, err
	}

	major, majorVerErr := common.MajorVersion(versionStr)
	if majorVerErr != nil {
		return 0, pnpm_err.NewPnpmError(
			&pnpm_err.FailedToParseError{},
			"invalid pnpm version format: "+versionStr,
			majorVerErr,
		)
	}

	return major, nil
}
