package pnpm_err

import (
	"fmt"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/common"
)

type OtherError struct{ common.BaseError }

var _ PnpmErrorIF = (*OtherError)(nil)

func (e *OtherError) Error() string {
	errMsg := "an unspecified error occurred"

	if e.Message != "" {
		errMsg = e.Message
	}

	if e.Cause != nil {
		errMsg = fmt.Sprintf("%s\ncaused by: %s", errMsg, e.Cause.Error())
	}
	return errMsg
}

func (e *OtherError) Is(target error) bool {
	_, ok := target.(*OtherError)
	return ok
}

func (e *OtherError) As(target any) bool {
	if t, ok := target.(**OtherError); ok {
		*t = e
		return true
	}
	return false
}
