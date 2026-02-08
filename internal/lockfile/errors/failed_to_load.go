package lockfile_err

import "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/common"

type FailedToLoadError struct{ common.BaseError }

var _ LockfileErrorIF = (*FailedToLoadError)(nil)

func (e *FailedToLoadError) Error() string {
	errMsg := "failed to load lockfile"

	if e.Message != "" {
		errMsg = e.Message
	}

	if e.Cause != nil {
		errMsg = errMsg + "\ncaused by: " + e.Cause.Error()
	}
	return errMsg
}

func (e *FailedToLoadError) Is(target error) bool {
	_, ok := target.(*FailedToLoadError)
	return ok
}

func (e *FailedToLoadError) As(target any) bool {
	if t, ok := target.(**FailedToLoadError); ok {
		*t = e
		return true
	}
	return false
}
