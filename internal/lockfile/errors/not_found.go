package lockfile_err

import (
	"fmt"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/common"
)

type LockfileNotFoundError struct{ common.BaseError }

var _ LockfileErrorIF = (*LockfileNotFoundError)(nil)

func (e *LockfileNotFoundError) Error() string {
	errMsg := "lockfile not found"
	if e.Message != "" {
		errMsg = fmt.Sprintf("%s: %s", errMsg, e.Message)
	}

	if e.Cause != nil {
		errMsg = fmt.Sprintf("%s\ncaused by: %s", errMsg, e.Cause.Error())
	}
	return errMsg
}

func (e *LockfileNotFoundError) Is(target error) bool {
	_, ok := target.(*LockfileNotFoundError)
	return ok
}

func (e *LockfileNotFoundError) As(target any) bool {
	if t, ok := target.(**LockfileNotFoundError); ok {
		*t = e
		return true
	}
	return false
}
