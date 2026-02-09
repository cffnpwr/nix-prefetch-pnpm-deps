package store_err

import "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/common"

type FailedToCleanupError struct{ common.BaseError }

var _ StoreErrorIF = (*FailedToCleanupError)(nil)

func (e *FailedToCleanupError) Error() string {
	errMsg := "failed to cleanup store"

	if e.Message != "" {
		errMsg = e.Message
	}

	if e.Cause != nil {
		errMsg = errMsg + "\ncaused by: " + e.Cause.Error()
	}
	return errMsg
}

func (e *FailedToCleanupError) Is(target error) bool {
	_, ok := target.(*FailedToCleanupError)
	return ok
}

func (e *FailedToCleanupError) As(target any) bool {
	if t, ok := target.(**FailedToCleanupError); ok {
		*t = e
		return true
	}
	return false
}
