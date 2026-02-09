package store_err

type StoreErrorIF interface {
	error
	Unwrap() error
	Is(target error) bool
	As(target any) bool

	SetMessage(string)
	SetCause(error)
}

func NewStoreError(e StoreErrorIF, message string, cause error) StoreErrorIF {
	e.SetMessage(message)
	e.SetCause(cause)
	return e
}
