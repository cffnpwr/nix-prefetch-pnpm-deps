# Testing Reference

## Test Structure

All tests follow this structure:

```go
func Test_FunctionName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		// input fields
		// expected output fields
	}{
		{
			name: "[正常系] description of normal case",
			// ...
		},
		{
			name: "[異常系] description of error case",
			// ...
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// test logic
		})
	}
}
```

### Requirements

- `t.Parallel()` on **both** the test function and each subtest
- Table-driven tests via struct slice
- Loop with `for _, tt := range tests` and `t.Run(tt.name, ...)`

## Test Case Naming

Test case names use Japanese with prefixes:

- `[正常系]` — Normal/positive cases (valid input, expected behavior)
- `[異常系]` — Error/negative cases (invalid input, edge cases, failures)

Examples:

```go
{name: "[正常系] カレントディレクトリ", ...}
{name: "[正常系] 9.15.0 → 9", ...}
{name: "[異常系] 空文字列", ...}
{name: "[異常系] NUL文字を含む", ...}
```

## Assertion Patterns

### Struct comparison with go-cmp

```go
if d := cmp.Diff(tt.want, got); d != "" {
	t.Errorf("FunctionName() mismatch (-want +got):\n%s", d)
}
```

### Error type comparison with reflect.TypeOf

Compares error **type** only, not message or cause content:

```go
if reflect.TypeOf(gotErr) != reflect.TypeOf(tt.wantErr) {
	t.Errorf("FunctionName() error = %v, wantErr %v", gotErr, tt.wantErr)
}
```

The `wantErr` field uses the domain error interface type (e.g., `lockfile_err.LockfileErrorIF`) and stores a zero-value pointer like `&lockfile_err.LockfileNotFoundError{}`.

### Simple value comparison

```go
if got != tt.want {
	t.Errorf("FunctionName(%q) = %v; want %v", tt.input, got, tt.want)
}
```

### Multiple return values

```go
if got != tt.want || gotBool != tt.wantBool {
	t.Errorf(
		"FunctionName(%q) = (%q, %v); want (%q, %v)",
		tt.s, got, gotBool, tt.want, tt.wantBool,
	)
}
```

## Filesystem Testing with afero

Tests that involve filesystem operations use a `setupFs` callback in test cases:

```go
tests := []struct {
	name    string
	setupFs func() afero.Fs
	path    string
	want    *lockfile.Lockfile
	wantErr lockfile_err.LockfileErrorIF
}{
	{
		name: "[正常系] ファイルが存在する",
		setupFs: func() afero.Fs {
			fs := afero.NewMemMapFs()
			_ = afero.WriteFile(fs, "/pnpm-lock.yaml",
				[]byte("lockfileVersion: '9.0'"), 0o644)
			return fs
		},
		path: "/pnpm-lock.yaml",
		want: &lockfile.Lockfile{LockfileVersion: "9.0"},
	},
	{
		name: "[異常系] ファイルが存在しない",
		setupFs: func() afero.Fs {
			return afero.NewMemMapFs()
		},
		path:    "/pnpm-lock.yaml",
		wantErr: &lockfile_err.LockfileNotFoundError{},
	},
}

for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
		fs := tt.setupFs()
		got, gotErr := lockfile.Load(fs, tt.path)
		// assertions...
	})
}
```

### Common afero operations in tests

```go
// Create file
afero.WriteFile(fs, "/path", []byte("content"), 0o644)

// Create directory
fs.Mkdir("/path", 0o755)

// Create executable file
afero.WriteFile(fs, "/bin/pnpm", []byte{}, 0o755)
```

## Environment Variable Testing

Use `t.Setenv()` (auto-restores after test):

```go
t.Run(tt.name, func(t *testing.T) {
	t.Setenv("PATH", tt.pathEnvVar)
	// test logic
})
```

Note: `t.Setenv()` is incompatible with `t.Parallel()` in the same subtest.

## OS-Specific Tests

Use build tags to separate platform-specific tests:

```go
//go:build unix

package path_test  // external test package for exported API

func Test_IsPath_Unix(t *testing.T) { ... }
```

```go
//go:build windows

package path  // internal test package for unexported helpers

func Test_extractDriveLetter(t *testing.T) { ... }
```

## Imports for Test Files

```go
import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/lockfile"
	lockfile_err "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/lockfile/errors"
)
```

Import alias convention: `pkgname_err` for error subpackages (e.g., `lockfile_err`, `pnpm_err`).
