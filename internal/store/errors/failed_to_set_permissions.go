package store_err

import "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/common"

type FailedToSetPermissionsError struct{ common.BaseError }

var _ StoreErrorIF = (*FailedToSetPermissionsError)(nil)

func (e *FailedToSetPermissionsError) Error() string {
	errMsg := "failed to set permissions"

	if e.Message != "" {
		errMsg = e.Message
	}

	if e.Cause != nil {
		errMsg = errMsg + "\ncaused by: " + e.Cause.Error()
	}
	return errMsg
}

func (e *FailedToSetPermissionsError) Is(target error) bool {
	_, ok := target.(*FailedToSetPermissionsError)
	return ok
}

func (e *FailedToSetPermissionsError) As(target any) bool {
	if t, ok := target.(**FailedToSetPermissionsError); ok {
		*t = e
		return true
	}
	return false
}
