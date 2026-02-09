package store

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/nix-community/go-nix/pkg/nar"
	"github.com/spf13/afero"

	store_err "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/store/errors"
)

// Hash computes the NAR hash of the store directory and returns it in SRI format (sha256-<base64>).
// The store must be normalized before hashing to produce a reproducible result.
// This is equivalent to running "nix hash path --type sha256" on the store directory.
func Hash(afs afero.Fs, storePath string) (string, store_err.StoreErrorIF) {
	h := sha256.New()

	nw, err := nar.NewWriter(h)
	if err != nil {
		return "", store_err.NewStoreError(
			&store_err.FailedToHashError{},
			"",
			err,
		)
	}

	if hashErr := writeNarEntry(afs, nw, storePath, "/"); hashErr != nil {
		return "", hashErr
	}

	if closeErr := nw.Close(); closeErr != nil {
		return "", store_err.NewStoreError(
			&store_err.FailedToHashError{},
			"",
			closeErr,
		)
	}

	digest := h.Sum(nil)
	sriHash := "sha256-" + base64.StdEncoding.EncodeToString(digest)

	return sriHash, nil
}

// writeNarEntry writes a single filesystem entry (file or directory) to the NAR writer.
// For directories, it reads entries sorted by name (afero.ReadDir returns sorted results)
// and recursively writes each child, matching the NAR spec's requirement for lexicographic order.
func writeNarEntry(
	afs afero.Fs,
	nw *nar.Writer,
	fsPath string,
	narPath string,
) store_err.StoreErrorIF {
	info, err := afs.Stat(fsPath)
	if err != nil {
		return store_err.NewStoreError(
			&store_err.FailedToHashError{},
			fsPath,
			err,
		)
	}

	if info.IsDir() {
		return writeNarDir(afs, nw, fsPath, narPath)
	}

	return writeNarFile(afs, nw, fsPath, narPath, info)
}

// writeNarDir writes a directory and all its children to the NAR writer.
func writeNarDir(
	afs afero.Fs,
	nw *nar.Writer,
	fsPath string,
	narPath string,
) store_err.StoreErrorIF {
	if err := nw.WriteHeader(&nar.Header{
		Path: narPath,
		Type: nar.TypeDirectory,
	}); err != nil {
		return store_err.NewStoreError(
			&store_err.FailedToHashError{},
			fsPath,
			err,
		)
	}

	// afero.ReadDir returns entries sorted by name, satisfying NAR's lexicographic order requirement
	entries, err := afero.ReadDir(afs, fsPath)
	if err != nil {
		return store_err.NewStoreError(
			&store_err.FailedToHashError{},
			fsPath,
			err,
		)
	}

	for _, entry := range entries {
		childFsPath := filepath.Join(fsPath, entry.Name())

		childNarPath := narPath + "/" + entry.Name()

		if entry.IsDir() {
			if dirErr := writeNarDir(afs, nw, childFsPath, childNarPath); dirErr != nil {
				return dirErr
			}
		} else {
			if fileErr := writeNarFile(afs, nw, childFsPath, childNarPath, entry); fileErr != nil {
				return fileErr
			}
		}
	}

	return nil
}

// writeNarFile writes a regular file's header and contents to the NAR writer.
func writeNarFile(
	afs afero.Fs,
	nw *nar.Writer,
	fsPath string,
	narPath string,
	info fs.FileInfo,
) store_err.StoreErrorIF {
	if err := nw.WriteHeader(&nar.Header{
		Path:       narPath,
		Type:       nar.TypeRegular,
		Size:       info.Size(),
		Executable: info.Mode()&0o111 != 0,
	}); err != nil {
		return store_err.NewStoreError(
			&store_err.FailedToHashError{},
			fsPath,
			err,
		)
	}

	f, err := afs.Open(fsPath)
	if err != nil {
		return store_err.NewStoreError(
			&store_err.FailedToHashError{},
			fsPath,
			err,
		)
	}
	defer f.Close()

	if _, copyErr := io.Copy(nw, f); copyErr != nil {
		return store_err.NewStoreError(
			&store_err.FailedToHashError{},
			fsPath,
			copyErr,
		)
	}

	return nil
}
