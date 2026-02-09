package store_err

import "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/common"

type FailedToNormalizeJSONError struct{ common.BaseError }

var _ StoreErrorIF = (*FailedToNormalizeJSONError)(nil)

func (e *FailedToNormalizeJSONError) Error() string {
	errMsg := "failed to normalize json"

	if e.Message != "" {
		errMsg = e.Message
	}

	if e.Cause != nil {
		errMsg = errMsg + "\ncaused by: " + e.Cause.Error()
	}
	return errMsg
}

func (e *FailedToNormalizeJSONError) Is(target error) bool {
	_, ok := target.(*FailedToNormalizeJSONError)
	return ok
}

func (e *FailedToNormalizeJSONError) As(target any) bool {
	if t, ok := target.(**FailedToNormalizeJSONError); ok {
		*t = e
		return true
	}
	return false
}
