package pnpm_err

type baseError struct {
	message string
	cause   error
}

func (e *baseError) Unwrap() error {
	return e.cause
}

func (e *baseError) setMessage(msg string) {
	e.message = msg
}

func (e *baseError) setCause(cause error) {
	e.cause = cause
}

type PnpmErrorIF interface {
	error
	Unwrap() error
	Is(target error) bool
	As(target any) bool

	setMessage(string)
	setCause(error)
}

func NewPnpmError(e PnpmErrorIF, message string, cause error) PnpmErrorIF {
	e.setMessage(message)
	e.setCause(cause)
	return e
}
