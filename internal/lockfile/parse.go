package lockfile

import (
	"github.com/spf13/afero"
	"go.yaml.in/yaml/v4"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/common"
	lockfile_err "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/lockfile/errors"
)

type Lockfile struct {
	LockfileVersion string `yaml:"lockfileVersion"`
}

func Parse(data []byte) (*Lockfile, lockfile_err.LockfileErrorIF) {
	var l Lockfile
	err := yaml.Unmarshal(data, &l)
	if err != nil {
		return nil, lockfile_err.NewLockfileError(
			&lockfile_err.FailedToParseError{},
			"",
			err,
		)
	}

	return &l, nil
}

func Load(fs afero.Fs, path string) (*Lockfile, lockfile_err.LockfileErrorIF) {
	// Check if path exists and is not a directory
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

	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, lockfile_err.NewLockfileError(
			&lockfile_err.FailedToLoadError{},
			"",
			err,
		)
	}

	return Parse(data)
}

func (l *Lockfile) MajorVersion() (int, lockfile_err.LockfileErrorIF) {
	major, err := common.MajorVersion(l.LockfileVersion)
	if err != nil {
		return 0, lockfile_err.NewLockfileError(
			&lockfile_err.FailedToParseError{},
			"invalid lockfile version format: "+l.LockfileVersion,
			err,
		)
	}

	return major, nil
}
