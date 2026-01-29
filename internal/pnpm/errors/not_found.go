package pnpm_err

import "fmt"

type PnpmNotFoundError struct{ baseError }

var _ PnpmErrorIF = (*PnpmNotFoundError)(nil)

func (e *PnpmNotFoundError) Error() string {
	errMsg := "pnpm not found"
	if e.message != "" {
		errMsg = fmt.Sprintf("%s: %s", errMsg, e.message)
	}

	if e.cause != nil {
		errMsg = fmt.Sprintf("%s\ncaused by: %s", errMsg, e.cause.Error())
	}
	return errMsg
}

func (e *PnpmNotFoundError) Is(target error) bool {
	_, ok := target.(*PnpmNotFoundError)
	return ok
}

func (e *PnpmNotFoundError) As(target any) bool {
	if t, ok := target.(**PnpmNotFoundError); ok {
		*t = e
		return true
	}
	return false
}
