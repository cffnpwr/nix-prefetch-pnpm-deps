package store_err

import "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/common"

type FailedToNormalizeJsonError struct{ common.BaseError }

var _ StoreErrorIF = (*FailedToNormalizeJsonError)(nil)

func (e *FailedToNormalizeJsonError) Error() string {
	errMsg := "failed to normalize json"

	if e.Message != "" {
		errMsg = e.Message
	}

	if e.Cause != nil {
		errMsg = errMsg + "\ncaused by: " + e.Cause.Error()
	}
	return errMsg
}

func (e *FailedToNormalizeJsonError) Is(target error) bool {
	_, ok := target.(*FailedToNormalizeJsonError)
	return ok
}

func (e *FailedToNormalizeJsonError) As(target any) bool {
	if t, ok := target.(**FailedToNormalizeJsonError); ok {
		*t = e
		return true
	}
	return false
}
