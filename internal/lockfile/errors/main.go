package lockfile_err

type LockfileErrorIF interface {
	error
	Unwrap() error
	Is(target error) bool
	As(target any) bool

	SetMessage(string)
	SetCause(error)
}

func NewLockfileError(e LockfileErrorIF, message string, cause error) LockfileErrorIF {
	e.SetMessage(message)
	e.SetCause(cause)
	return e
}
