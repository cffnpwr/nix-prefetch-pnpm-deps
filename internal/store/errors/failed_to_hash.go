package store_err

import "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/common"

type FailedToHashError struct{ common.BaseError }

var _ StoreErrorIF = (*FailedToHashError)(nil)

func (e *FailedToHashError) Error() string {
	errMsg := "failed to hash store"

	if e.Message != "" {
		errMsg = e.Message
	}

	if e.Cause != nil {
		errMsg = errMsg + "\ncaused by: " + e.Cause.Error()
	}
	return errMsg
}

func (e *FailedToHashError) Is(target error) bool {
	_, ok := target.(*FailedToHashError)
	return ok
}

func (e *FailedToHashError) As(target any) bool {
	if t, ok := target.(**FailedToHashError); ok {
		*t = e
		return true
	}
	return false
}
