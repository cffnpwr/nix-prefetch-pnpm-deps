package store_err

import "errors"

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

// AsStoreError is a helper function to extract StoreErrorIF from an error using errors.As.
func AsStoreError(err error, target *StoreErrorIF) bool {
	return errors.As(err, target)
}
