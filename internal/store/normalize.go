package store

import (
	"bytes"
	"encoding/json"
	"errors"
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
	if err := normalizeJSONFiles(afs, opts.StorePath); err != nil {
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
	//nolint:mnd // fetcherVersion 2+ requires permission normalization
	if opts.FetcherVersion >= 2 {
		if err := setPermissions(afs, opts.StorePath); err != nil {
			return err
		}
	}

	return nil
}

// normalizeJSONFiles walks storePath, finds all .json files, and normalizes each one.
// Stops and returns the error on the first failure.
func normalizeJSONFiles(afs afero.Fs, storePath string) store_err.StoreErrorIF {
	normalizeErr := afero.Walk(
		afs,
		storePath,
		func(path string, info fs.FileInfo, walkErr error) error {
			if walkErr != nil {
				return store_err.NewStoreError(
					&store_err.FailedToNormalizeJSONError{},
					"",
					walkErr,
				)
			}

			if info.IsDir() || !strings.HasSuffix(path, ".json") {
				return nil
			}

			if e := normalizeJSONFile(afs, path, info.Mode()); e != nil {
				return e
			}

			return nil
		},
	)

	if normalizeErr != nil {
		var storeErr store_err.StoreErrorIF
		if errors.As(normalizeErr, &storeErr) {
			return storeErr
		}

		return store_err.NewStoreError(
			&store_err.FailedToNormalizeJSONError{},
			"",
			normalizeErr,
		)
	}

	return nil
}

// normalizeJSONFile normalizes a single JSON file for reproducible Nix hash computation.
// pnpm writes JSON with non-deterministic key order and includes "checkedAt" timestamps
// that change on every install. This function removes all "checkedAt" keys and writes
// the file back with sorted keys so the same pnpm-lock.yaml always produces the same hash.
// UseNumber preserves original number formatting, and SetEscapeHTML(false) matches jq output.
func normalizeJSONFile(afs afero.Fs, path string, mode fs.FileMode) store_err.StoreErrorIF {
	data, readErr := afero.ReadFile(afs, path)
	if readErr != nil {
		return store_err.NewStoreError(
			&store_err.FailedToNormalizeJSONError{},
			path,
			readErr,
		)
	}

	// Decode with UseNumber to preserve original number literals (avoid float64 precision loss)
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var v any
	if decodeErr := dec.Decode(&v); decodeErr != nil {
		return store_err.NewStoreError(
			&store_err.FailedToNormalizeJSONError{},
			path,
			decodeErr,
		)
	}

	v = removeCheckedAt(v)

	// Re-encode with sorted keys (json.Marshal sorts map keys) and 2-space indent to match jq output.
	// SetEscapeHTML(false) prevents Go from escaping <, >, & which jq does not escape.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if encodeErr := enc.Encode(v); encodeErr != nil {
		return store_err.NewStoreError(
			&store_err.FailedToNormalizeJSONError{},
			path,
			encodeErr,
		)
	}

	if writeErr := afero.WriteFile(afs, path, buf.Bytes(), mode); writeErr != nil {
		return store_err.NewStoreError(
			&store_err.FailedToNormalizeJSONError{},
			path,
			writeErr,
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
	walkErr := afero.Walk(afs, storePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return store_err.NewStoreError(
				&store_err.FailedToSetPermissionsError{},
				"",
				err,
			)
		}

		var mode fs.FileMode
		switch {
		case info.IsDir():
			mode = 0o555
		case strings.HasSuffix(info.Name(), "-exec"):
			mode = 0o555
		default:
			mode = 0o444
		}

		if chmodErr := afs.Chmod(path, mode); chmodErr != nil {
			return store_err.NewStoreError(
				&store_err.FailedToSetPermissionsError{},
				path,
				chmodErr,
			)
		}

		return nil
	})

	if walkErr != nil {
		var storeErr store_err.StoreErrorIF
		if errors.As(walkErr, &storeErr) {
			return storeErr
		}

		return store_err.NewStoreError(
			&store_err.FailedToSetPermissionsError{},
			"",
			walkErr,
		)
	}

	return nil
}
