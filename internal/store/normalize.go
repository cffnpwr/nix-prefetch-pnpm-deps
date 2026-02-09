package store

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	store_err "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/store/errors"
)

var storeVersionDirs = []string{"v3", "v10"}

type NormalizeOptions struct {
	StorePath      string
	FetcherVersion int
}

func Normalize(afs afero.Fs, opts NormalizeOptions) store_err.StoreErrorIF {
	// Step 1: Remove temporary directories
	for _, dir := range storeVersionDirs {
		tmpPath := filepath.Join(opts.StorePath, dir, "tmp")
		if err := afs.RemoveAll(tmpPath); err != nil {
			return store_err.NewStoreError(
				&store_err.FailedToCleanupError{},
				tmpPath,
				err,
			)
		}
	}

	// Step 2: Normalize JSON files
	if err := normalizeJsonFiles(afs, opts.StorePath); err != nil {
		return err
	}

	// Step 3: Remove projects directories
	for _, dir := range storeVersionDirs {
		projectsPath := filepath.Join(opts.StorePath, dir, "projects")
		if err := afs.RemoveAll(projectsPath); err != nil {
			return store_err.NewStoreError(
				&store_err.FailedToCleanupError{},
				projectsPath,
				err,
			)
		}
	}

	// Step 4: Set permissions (fetcherVersion >= 2 only)
	if opts.FetcherVersion >= 2 {
		if err := setPermissions(afs, opts.StorePath); err != nil {
			return err
		}
	}

	return nil
}

// normalizeJsonFiles walks storePath, finds all .json files, and normalizes each one.
// Stops and returns the error on the first failure.
func normalizeJsonFiles(afs afero.Fs, storePath string) store_err.StoreErrorIF {
	var normalizeErr store_err.StoreErrorIF

	_ = afero.Walk(afs, storePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			normalizeErr = store_err.NewStoreError(
				&store_err.FailedToNormalizeJsonError{},
				"",
				err,
			)
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		if e := normalizeJsonFile(afs, path, info.Mode()); e != nil {
			normalizeErr = e
			return e
		}

		return nil
	})

	return normalizeErr
}

// normalizeJsonFile normalizes a single JSON file for reproducible Nix hash computation.
// pnpm writes JSON with non-deterministic key order and includes "checkedAt" timestamps
// that change on every install. This function removes all "checkedAt" keys and writes
// the file back with sorted keys so the same pnpm-lock.yaml always produces the same hash.
// UseNumber preserves original number formatting, and SetEscapeHTML(false) matches jq output.
func normalizeJsonFile(afs afero.Fs, path string, mode fs.FileMode) store_err.StoreErrorIF {
	data, err := afero.ReadFile(afs, path)
	if err != nil {
		return store_err.NewStoreError(
			&store_err.FailedToNormalizeJsonError{},
			path,
			err,
		)
	}

	// Decode with UseNumber to preserve original number literals (avoid float64 precision loss)
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var v any
	if err := dec.Decode(&v); err != nil {
		return store_err.NewStoreError(
			&store_err.FailedToNormalizeJsonError{},
			path,
			err,
		)
	}

	v = removeCheckedAt(v)

	// Re-encode with sorted keys (json.Marshal sorts map keys) and 2-space indent to match jq output.
	// SetEscapeHTML(false) prevents Go from escaping <, >, & which jq does not escape.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(v); err != nil {
		return store_err.NewStoreError(
			&store_err.FailedToNormalizeJsonError{},
			path,
			err,
		)
	}

	if err := afero.WriteFile(afs, path, buf.Bytes(), mode); err != nil {
		return store_err.NewStoreError(
			&store_err.FailedToNormalizeJsonError{},
			path,
			err,
		)
	}

	return nil
}

// removeCheckedAt recursively removes all "checkedAt" keys from a decoded JSON value.
// Equivalent to jq's "del(.. | .checkedAt?)".
func removeCheckedAt(v any) any {
	switch val := v.(type) {
	case map[string]any:
		result := make(map[string]any, len(val))
		for k, v := range val {
			if k == "checkedAt" {
				continue
			}
			result[k] = removeCheckedAt(v)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, v := range val {
			result[i] = removeCheckedAt(v)
		}
		return result
	default:
		return v
	}
}

// setPermissions walks storePath and sets fixed permissions on all entries so that
// nix hash produces the same result regardless of the build environment.
//   - directories:        0555 (r-xr-xr-x)
//   - files named *-exec: 0555 (r-xr-xr-x)
//   - other files:        0444 (r--r--r--)
func setPermissions(afs afero.Fs, storePath string) store_err.StoreErrorIF {
	err := afero.Walk(afs, storePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return store_err.NewStoreError(
				&store_err.FailedToSetPermissionsError{},
				"",
				err,
			)
		}

		var mode fs.FileMode
		if info.IsDir() {
			mode = 0o555
		} else if strings.HasSuffix(info.Name(), "-exec") {
			mode = 0o555
		} else {
			mode = 0o444
		}

		if err := afs.Chmod(path, mode); err != nil {
			return store_err.NewStoreError(
				&store_err.FailedToSetPermissionsError{},
				path,
				err,
			)
		}

		return nil
	})

	if err != nil {
		return err.(store_err.StoreErrorIF)
	}

	return nil
}
