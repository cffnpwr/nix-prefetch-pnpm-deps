package store_err

import "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/common"

type FailedToCreateTarballError struct{ common.BaseError }

var _ StoreErrorIF = (*FailedToCreateTarballError)(nil)

func (e *FailedToCreateTarballError) Error() string {
	errMsg := "failed to create tarball"

	if e.Message != "" {
		errMsg = e.Message
	}

	if e.Cause != nil {
		errMsg = errMsg + "\ncaused by: " + e.Cause.Error()
	}
	return errMsg
}

func (e *FailedToCreateTarballError) Is(target error) bool {
	_, ok := target.(*FailedToCreateTarballError)
	return ok
}

func (e *FailedToCreateTarballError) As(target any) bool {
	if t, ok := target.(**FailedToCreateTarballError); ok {
		*t = e
		return true
	}
	return false
}
