# CLI Flag State & Reset Analysis

## Summary
Flag state reset issues exist in integration tests due to `sync.Once` usage in cobraflags package preventing re-binding on multiple `cli.Execute()` calls.

## 1. Cobraflags Package Architecture

### Location
- Module: `github.com/go-extras/cobraflags v0.0.0-20260116100222-f76efc9500d4`
- Cache: `/Users/cffnpwr/go/pkg/mod/github.com/go-extras/cobraflags@v0.0.0-20260116100222-f76efc9500d4/`

### Key Components

#### FlagBase[T] Structure (cobraflags.go:71-87)
```go
type FlagBase[T any] struct {
	Name         string        
	ViperKey     string        
	Shorthand    string        
	Usage        string        
	Required     bool          
	Persistent   bool          
	Value        T             
	ValidateFunc func(T) error 
	Validator    Validator     
	
	flag     *pflag.Flag
	bindOnce sync.Once        // <-- CRITICAL: prevents re-binding
	
	flagGetter
	flagGetterE
}
```

#### StringSliceFlag Implementation (flag_stringslice.go:46-121)
- `GetStringSlice()` (lines 83-91): Uses `s.bindOnce.Do()` to bind once to Viper
- `GetStringSliceE()` (lines 107-121): Also uses `s.bindOnce.Do()` for one-time binding
- Similar pattern in: flag_string.go, flag_int.go, flag_bool.go, flag_uint8.go

### The Core Problem: sync.Once Locking

**Problem**: Each flag instance has a `bindOnce sync.Once` field
```go
func (s *StringSliceFlag) GetStringSlice() []string {
	viperKey := pStringSliceFlag(s).getViperKey()
	
	s.bindOnce.Do(func() {
		noError(viper.BindPFlag(viperKey, s.flag))
	})
	
	return viper.GetStringSlice(viperKey)
}
```

Once the closure in `bindOnce.Do()` executes, it NEVER executes again for that flag instance. This means:
1. First `cli.Execute()` call: flag is bound to Viper
2. Second `cli.Execute()` call: flag is NOT re-bound (sync.Once prevents it)
3. Viper retains old values from first execution

## 2. CLI Package Exported API

### Location: internal/cli/

#### Exported Functions
- **`Execute() error`** (root.go:32-34)
  - Main entry point for CLI
  - Calls `rootCmd.Execute()`
  - No state reset mechanism

#### Global Flag Instances (flags.go:18-76)
Package-level variables:
- `fetcherVersionFlag` (*cobraflags.IntFlag) - required, range 1-3
- `pnpmPathFlag` (*cobraflags.StringFlag)
- `workspaceFlag` (*cobraflags.StringSliceFlag) - multi-value
- `pnpmFlagFlag` (*cobraflags.StringSliceFlag) - multi-value
- `hashFlag` (*cobraflags.StringFlag)
- `quietFlag` (*cobraflags.BoolFlag)

#### Init Function (root.go:23-30)
Registers all flags with rootCmd:
```go
func init() {
	fetcherVersionFlag.Register(rootCmd)
	pnpmPathFlag.Register(rootCmd)
	workspaceFlag.Register(rootCmd)
	pnpmFlagFlag.Register(rootCmd)
	hashFlag.Register(rootCmd)
	quietFlag.Register(rootCmd)
}
```

### Important Note: Flags are Global
All flags are defined as package-level globals, so they persist across multiple calls to `cli.Execute()`.

## 3. Viper Integration

### In Codebase
**NO** direct viper usage in `/internal/cli/`:
- Viper is used ONLY indirectly through cobraflags
- Each flag's `GetXxx()` calls internally use `viper.BindPFlag()` and `viper.GetXxx()`

### Viper State Management
Viper is a global singleton:
```go
// In cobraflags package
func (s *StringSliceFlag) GetStringSlice() []string {
	viperKey := pStringSliceFlag(s).getViperKey()
	s.bindOnce.Do(func() {
		noError(viper.BindPFlag(viperKey, s.flag))  // First call only
	})
	return viper.GetStringSlice(viperKey)  // Returns whatever Viper has
}
```

**Viper issue**: 
- Global singleton maintains values across Execute() calls
- Once bound, Viper will return the last-set value unless explicitly cleared

## 4. Root Cause Summary

### The Problem in Tests
Current integration test pattern (integration_test.go:96-139):
```go
func executeCLI(t *testing.T, args []string) (string, error) {
	os.Args = append([]string{"cmd"}, args...)
	execErr := cli.Execute()
	// Problem: os.Args is set per test
	// But cobraflags maintains sync.Once + Viper state across calls
}
```

When called multiple times in same test process:
1. **First call**: `bindOnce.Do()` executes, Viper binds, flag values read correctly
2. **Second call**: `bindOnce.Do()` SKIPPED (already executed), Viper still has old values
3. Result: Second test gets first test's flag values

### Why Viper Isn't Reset
- Viper is a global singleton (github.com/spf13/viper)
- No exposed reset/clear method in integration
- Values persist across cobra command executions
