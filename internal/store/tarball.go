package store

import (
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/afero"

	store_err "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/store/errors"
)

// SOURCE_DATE_EPOCH used by Nix for reproducible builds (1980-01-01 00:00:00 UTC).
const sourceDateEpoch = 315532800

type fileEntry struct {
	path    string
	relPath string
	info    fs.FileInfo
}

// readSymlinkTarget reads the target of a symlink using the filesystem's ReadlinkIfPossible method.
func readSymlinkTarget(afs afero.Fs, path string) (string, error) {
	type linkReader interface {
		ReadlinkIfPossible(string) (string, error)
	}

	lr, ok := afs.(linkReader)
	if !ok {
		return "", fs.ErrInvalid
	}

	return lr.ReadlinkIfPossible(path)
}

// entryTarPath computes the tar path for a file entry.
func entryTarPath(entry fileEntry) string {
	if entry.relPath == "." {
		return "./"
	}

	p := "./" + entry.relPath
	if entry.info.IsDir() && !strings.HasSuffix(p, "/") {
		p += "/"
	}

	return p
}

// resolveEntryType determines the tar type flag and symlink target for a file entry.
func resolveEntryType(
	afs afero.Fs,
	entry fileEntry,
) (byte, string, store_err.StoreErrorIF) {
	info := entry.info

	switch {
	case info.IsDir():
		return tarTypeDir, "", nil
	case info.Mode()&fs.ModeSymlink != 0:
		target, err := readSymlinkTarget(afs, entry.path)
		if err != nil {
			return 0, "", store_err.NewStoreError(
				&store_err.FailedToCreateTarballError{},
				entry.path,
				err,
			)
		}

		return tarTypeSymlink, target, nil
	default:
		return tarTypeFile, "", nil
	}
}

// openFileReader opens a regular file for reading if it has content.
func openFileReader(afs afero.Fs, entry fileEntry) (int64, io.Reader, store_err.StoreErrorIF) {
	info := entry.info

	if !info.Mode().IsRegular() || info.Size() <= 0 {
		return 0, nil, nil
	}

	file, openErr := afs.Open(entry.path)
	if openErr != nil {
		return 0, nil, store_err.NewStoreError(
			&store_err.FailedToCreateTarballError{},
			entry.path,
			openErr,
		)
	}

	return info.Size(), &closingReader{file: file}, nil
}

// writeStoreEntry writes a single store entry to the tar writer.
func writeStoreEntry(
	afs afero.Fs,
	tw *gnuTarWriter,
	entry fileEntry,
) store_err.StoreErrorIF {
	typeflag, linkTarget, typeErr := resolveEntryType(afs, entry)
	if typeErr != nil {
		return typeErr
	}

	fileSize, dataReader, readerErr := openFileReader(afs, entry)
	if readerErr != nil {
		return readerErr
	}

	if err := tw.writeEntry(entryTarPath(entry), tarEntryInfo{
		mode:       int64(entry.info.Mode().Perm()),
		size:       fileSize,
		mtime:      sourceDateEpoch,
		typeflag:   typeflag,
		linkname:   linkTarget,
		dataReader: dataReader,
	}); err != nil {
		return store_err.NewStoreError(
			&store_err.FailedToCreateTarballError{},
			entry.path,
			err,
		)
	}

	return nil
}

// closingReader wraps afero.File to close after reading all content.
type closingReader struct {
	file afero.File
	done bool
}

func (r *closingReader) Read(p []byte) (int, error) {
	n, err := r.file.Read(p)
	if err != nil && !r.done {
		r.done = true
		_ = r.file.Close()
	}

	return n, err
}

// collectSortedEntries walks storePath and returns file entries sorted by relative path.
func collectSortedEntries(afs afero.Fs, storePath string) ([]fileEntry, store_err.StoreErrorIF) {
	var entries []fileEntry

	walkErr := afero.Walk(afs, storePath, func(path string, info fs.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, relErr := filepath.Rel(storePath, path)
		if relErr != nil {
			return relErr
		}

		entries = append(entries, fileEntry{
			path:    path,
			relPath: relPath,
			info:    info,
		})

		return nil
	})

	if walkErr != nil {
		return nil, store_err.NewStoreError(
			&store_err.FailedToCreateTarballError{},
			"",
			walkErr,
		)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].relPath < entries[j].relPath
	})

	return entries, nil
}

// writeTarball writes sorted entries as a zstd-compressed tarball to the output file.
func writeTarball(
	afs afero.Fs,
	entries []fileEntry,
	outputPath string,
	outFile afero.File,
) store_err.StoreErrorIF {
	// Use CGo-linked C zstd library with CLI-compatible parameters
	// (compression level 3, content checksum enabled) for byte-identical output.
	zw, zstdErr := newZstdWriter(outFile)
	if zstdErr != nil {
		return store_err.NewStoreError(
			&store_err.FailedToCreateTarballError{},
			outputPath,
			zstdErr,
		)
	}

	tw := newGNUTarWriter(zw)

	for _, entry := range entries {
		if err := writeStoreEntry(afs, tw, entry); err != nil {
			_ = zw.Close()

			return err
		}
	}

	if closeErr := tw.close(); closeErr != nil {
		_ = zw.Close()

		return store_err.NewStoreError(
			&store_err.FailedToCreateTarballError{},
			outputPath,
			closeErr,
		)
	}

	if closeErr := zw.Close(); closeErr != nil {
		return store_err.NewStoreError(
			&store_err.FailedToCreateTarballError{},
			outputPath,
			closeErr,
		)
	}

	return nil
}

// CreateTarball creates a reproducible zstd-compressed tarball from storePath.
// Produces output byte-identical to GNU tar with:
//
//	tar --sort=name --mtime="@315532800" --owner=0 --group=0 --numeric-owner \
//	  --pax-option=exthdr.name=%d/PaxHeaders/%f,delete=atime,delete=ctime \
//	  --zstd -cf - -C storePath .
//
// Uses CGo-linked C zstd library for compression compatibility.
func CreateTarball(afs afero.Fs, storePath string, outputPath string) store_err.StoreErrorIF {
	entries, collectErr := collectSortedEntries(afs, storePath)
	if collectErr != nil {
		return collectErr
	}

	outFile, createErr := afs.Create(outputPath)
	if createErr != nil {
		return store_err.NewStoreError(
			&store_err.FailedToCreateTarballError{},
			outputPath,
			createErr,
		)
	}
	defer outFile.Close()

	return writeTarball(afs, entries, outputPath, outFile)
}
