package common

type BaseError struct {
	Message string
	Cause   error
}

func (e *BaseError) Unwrap() error {
	return e.Cause
}

func (e *BaseError) SetMessage(msg string) {
	e.Message = msg
}

func (e *BaseError) SetCause(cause error) {
	e.Cause = cause
}
