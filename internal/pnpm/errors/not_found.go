package pnpm_err

import (
	"fmt"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/common"
)

type PnpmNotFoundError struct{ common.BaseError }

var _ PnpmErrorIF = (*PnpmNotFoundError)(nil)

func (e *PnpmNotFoundError) Error() string {
	errMsg := "pnpm not found"
	if e.Message != "" {
		errMsg = fmt.Sprintf("%s: %s", errMsg, e.Message)
	}

	if e.Cause != nil {
		errMsg = fmt.Sprintf("%s\ncaused by: %s", errMsg, e.Cause.Error())
	}
	return errMsg
}

func (e *PnpmNotFoundError) Is(target error) bool {
	_, ok := target.(*PnpmNotFoundError)
	return ok
}

func (e *PnpmNotFoundError) As(target any) bool {
	if t, ok := target.(**PnpmNotFoundError); ok {
		*t = e
		return true
	}
	return false
}
