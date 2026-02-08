---
name: write-tests
description: Workflow for writing tests in this Go project. Use when creating new test files, adding test cases to existing tests, or writing tests for new functions/methods. Covers table-driven tests, parallel execution, afero filesystem testing, go-cmp assertions, and Japanese naming conventions.
---

# Write Tests

## Reference

Full pattern details: `.claude/docs/reference/testing.md`

## Steps

### Step 1: Determine test file and package

- File: `<source_file>_test.go` in the same directory
- External test package (`package foo_test`) for exported API
- Internal test package (`package foo`) only for unexported functions
- OS-specific: add `//go:build unix` or `//go:build windows`

### Step 2: Write test function

```go
func Test_FunctionName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		// input and expected output fields
	}{
		// test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// assertions
		})
	}
}
```

### Step 3: Define test cases

Name with Japanese and `[正常系]`/`[異常系]` prefix:

```go
{name: "[正常系] 正しい入力", input: "valid", want: expected},
{name: "[異常系] 空文字列",   input: "",      wantErr: &pkg_err.SomeError{}},
```

For filesystem tests, use `setupFs func() afero.Fs`:

```go
{
	name: "[正常系] ファイルが存在する",
	setupFs: func() afero.Fs {
		fs := afero.NewMemMapFs()
		_ = afero.WriteFile(fs, "/file", []byte("data"), 0o644)
		return fs
	},
},
```

### Step 4: Write assertions

| What | Pattern |
|---|---|
| Structs | `cmp.Diff(tt.want, got)` → `t.Errorf("mismatch (-want +got):\n%s", d)` |
| Error types | `reflect.TypeOf(gotErr) != reflect.TypeOf(tt.wantErr)` |
| Simple values | `got != tt.want` → `t.Errorf` |

### Step 5: Add imports

```go
import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/<package>"
	pkg_err "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/<package>/errors"
)
```

Error subpackage alias: `<pkg>_err` (e.g., `lockfile_err`, `pnpm_err`).

## Checklist

- [ ] `t.Parallel()` on test function AND each subtest
- [ ] Table-driven with `[]struct` and `for _, tt := range tests`
- [ ] Names in Japanese with `[正常系]`/`[異常系]` prefix
- [ ] Both normal and error cases covered
- [ ] Structs compared with `cmp.Diff`, errors with `reflect.TypeOf`
- [ ] `afero.NewMemMapFs()` for filesystem tests (not real FS)
- [ ] Import order: stdlib, external, local
- [ ] `t.Setenv()` for env vars (not with `t.Parallel()` in same subtest)
