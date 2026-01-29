package pnpm_err

import "fmt"

type OtherError struct{ baseError }

var _ PnpmErrorIF = (*OtherError)(nil)

func (e *OtherError) Error() string {
	errMsg := "an unspecified error occurred"

	if e.message != "" {
		errMsg = e.message
	}

	if e.cause != nil {
		errMsg = fmt.Sprintf("%s\ncaused by: %s", errMsg, e.cause.Error())
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
