package pnpm_err

import "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/common"

type FailedToExecuteError struct{ common.BaseError }

var _ PnpmErrorIF = (*FailedToExecuteError)(nil)

func (e *FailedToExecuteError) Error() string {
	errMsg := "failed to execute pnpm command"

	if e.Message != "" {
		errMsg = e.Message
	}

	if e.Cause != nil {
		errMsg = errMsg + "\ncaused by: " + e.Cause.Error()
	}
	return errMsg
}

func (e *FailedToExecuteError) Is(target error) bool {
	_, ok := target.(*FailedToExecuteError)
	return ok
}

func (e *FailedToExecuteError) As(target any) bool {
	if t, ok := target.(**FailedToExecuteError); ok {
		*t = e
		return true
	}
	return false
}
