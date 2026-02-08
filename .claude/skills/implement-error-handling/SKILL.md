---
name: implement-error-handling
description: Workflow for implementing error types in this project's domain error hierarchy. Use when adding a new error type to lockfile/errors/ or pnpm/errors/, or creating a new domain package with its own error hierarchy. Also use when modifying existing error types or reviewing error handling patterns.
---

# Implement Error Handling

## Reference

Full pattern details: `.claude/docs/reference/error-handling.md`

## Adding a New Error Type to an Existing Domain

### Step 1: Create the error file

Create `internal/<domain>/errors/<error_name>.go`:

```go
package errors

import (
	"fmt"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/common"
)

type NewErrorName struct{ common.BaseError }

var _ <Domain>ErrorIF = (*NewErrorName)(nil)

func (e *NewErrorName) Error() string {
	errMsg := "<base error message>"
	if e.Message != "" {
		errMsg = fmt.Sprintf("%s: %s", errMsg, e.Message)
	}
	if e.Cause != nil {
		errMsg = fmt.Sprintf("%s\ncaused by: %s", errMsg, e.Cause.Error())
	}
	return errMsg
}

func (e *NewErrorName) Is(target error) bool {
	_, ok := target.(*NewErrorName)
	return ok
}

func (e *NewErrorName) As(target any) bool {
	if t, ok := target.(**NewErrorName); ok {
		*t = e
		return true
	}
	return false
}
```

### Step 2: Use in application code

```go
return nil, domain_err.NewDomainError(
	&domain_err.NewErrorName{},
	"optional detail message",  // "" if base message is sufficient
	originalErr,                // nil if no underlying error
)
```

## Creating a New Domain Error Hierarchy

### Step 1: Create `internal/<domain>/errors/main.go`

```go
package errors

type <Domain>ErrorIF interface {
	error
	Unwrap() error
	Is(target error) bool
	As(target any) bool

	SetMessage(string)
	SetCause(error)
}

func New<Domain>Error(e <Domain>ErrorIF, message string, cause error) <Domain>ErrorIF {
	e.SetMessage(message)
	e.SetCause(cause)
	return e
}
```

### Step 2: Add concrete error types

Follow "Adding a New Error Type" above for each type.

## Checklist

- [ ] Error struct embeds `common.BaseError`
- [ ] Compile-time check: `var _ IF = (*Concrete)(nil)`
- [ ] `Error()` builds message: base + optional `Message` + optional `Cause`
- [ ] `Is()` checks type via `target.(*TypeName)`
- [ ] `As()` uses double pointer `**T`
- [ ] Each error type in its own file
- [ ] Return type is domain error interface, not `error`
- [ ] Use `NewXxxError()` factory for construction
