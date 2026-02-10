# Error Handling Reference

## Architecture

```
common.BaseError (base struct)
    ├── lockfile/errors/
    │   ├── LockfileErrorIF (interface)
    │   ├── LockfileNotFoundError
    │   ├── FailedToLoadError
    │   └── FailedToParseError
    ├── pnpm/errors/
    │   ├── PnpmErrorIF (interface)
    │   ├── PnpmNotFoundError
    │   ├── FailedToExecuteError
    │   ├── FailedToParseError
    │   └── OtherError
    └── store/errors/
        ├── StoreErrorIF (interface)
        ├── FailedToCleanupError
        ├── FailedToCreateTarballError
        ├── FailedToHashError
        ├── FailedToNormalizeJSONError
        └── FailedToSetPermissionsError
```

## BaseError (`internal/common/errors.go`)

```go
type BaseError struct {
	Message string
	Cause   error
}

func (e *BaseError) Unwrap() error    { return e.Cause }
func (e *BaseError) SetMessage(msg string) { e.Message = msg }
func (e *BaseError) SetCause(cause error)  { e.Cause = cause }
```

- `Message`: User-facing custom message
- `Cause`: Wrapped original error (for error chains)
- `Unwrap()`: Supports Go 1.13+ error unwrapping

## Error Interface Pattern

Each domain package defines its own error interface in `errors/main.go`:

```go
type XxxErrorIF interface {
	error
	Unwrap() error
	Is(target error) bool
	As(target any) bool

	SetMessage(string)
	SetCause(error)
}
```

Factory function in the same file:

```go
func NewXxxError(e XxxErrorIF, message string, cause error) XxxErrorIF {
	e.SetMessage(message)
	e.SetCause(cause)
	return e
}
```

## Concrete Error Type Pattern

Each concrete error type lives in its own file (e.g., `errors/not_found.go`):

```go
type LockfileNotFoundError struct{ common.BaseError }

// Compile-time interface check
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
```

### Key details

- `Error()`: Returns base message, appends `Message` if set, appends `Cause` with `\ncaused by: ` prefix
- `Is()`: Type-only check via `target.(*TypeName)`
- `As()`: Uses double pointer `**T` to set value
- Compile-time check: `var _ IF = (*Concrete)(nil)` ensures interface compliance

## Usage in Application Code

Error return type is the domain error interface, not `error`:

```go
func Load(fs afero.Fs, path string) (*Lockfile, lockfile_err.LockfileErrorIF) {
	f, err := fs.Stat(path)
	if err != nil {
		return nil, lockfile_err.NewLockfileError(
			&lockfile_err.LockfileNotFoundError{},
			"",
			err,
		)
	}
	if f.IsDir() {
		return nil, lockfile_err.NewLockfileError(
			&lockfile_err.FailedToLoadError{},
			"path is a directory",
			nil,
		)
	}
	// ...
}
```

- Pass empty `""` for message when the base error message is sufficient
- Pass `nil` for cause when there is no underlying error
- Always use `NewXxxError()` factory — never construct the struct directly with fields set
