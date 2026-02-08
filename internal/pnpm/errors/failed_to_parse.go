package pnpm_err

import "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/common"

type FailedToParseError struct{ common.BaseError }

var _ PnpmErrorIF = (*FailedToParseError)(nil)

func (e *FailedToParseError) Error() string {
	errMsg := "failed to parse pnpm output"

	if e.Message != "" {
		errMsg = e.Message
	}

	if e.Cause != nil {
		errMsg = errMsg + "\ncaused by: " + e.Cause.Error()
	}
	return errMsg
}

func (e *FailedToParseError) Is(target error) bool {
	_, ok := target.(*FailedToParseError)
	return ok
}

func (e *FailedToParseError) As(target any) bool {
	if t, ok := target.(**FailedToParseError); ok {
		*t = e
		return true
	}
	return false
}
