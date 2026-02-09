package lockfile_test

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/lockfile"
	lockfile_err "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/lockfile/errors"
)

func Test_Parse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    []byte
		want    *lockfile.Lockfile
		wantErr lockfile_err.LockfileErrorIF
	}{
		{
			name: "[正常系] バージョン9.0",
			data: []byte("lockfileVersion: '9.0'"),
			want: &lockfile.Lockfile{LockfileVersion: "9.0"},
		},
		{
			name: "[正常系] バージョン6.1",
			data: []byte("lockfileVersion: '6.1'"),
			want: &lockfile.Lockfile{LockfileVersion: "6.1"},
		},
		{
			name: "[正常系] 他のフィールドがあっても無視",
			data: []byte("lockfileVersion: '9.0'\nsettings:\n  autoInstallPeers: true"),
			want: &lockfile.Lockfile{LockfileVersion: "9.0"},
		},
		{
			name:    "[異常系] 無効なYAML",
			data:    []byte("{invalid yaml"),
			wantErr: &lockfile_err.FailedToParseError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, gotErr := lockfile.Parse(tt.data)
			if d := cmp.Diff(tt.want, got); d != "" {
				t.Errorf("Parse() mismatch (-want +got):\n%s", d)
			}
			if reflect.TypeOf(gotErr) != reflect.TypeOf(tt.wantErr) {
				t.Errorf("Parse() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func Test_Load(t *testing.T) {
	t.Parallel()

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
				_ = afero.WriteFile(fs, "/pnpm-lock.yaml", []byte("lockfileVersion: '9.0'"), 0o644)
				return fs
			},
			path: "/pnpm-lock.yaml",
			want: &lockfile.Lockfile{LockfileVersion: "9.0"},
		},
		{
			name:    "[異常系] ファイルが存在しない",
			setupFs: afero.NewMemMapFs,
			path:    "/pnpm-lock.yaml",
			wantErr: &lockfile_err.LockfileNotFoundError{},
		},
		{
			name: "[異常系] 読み込み失敗（ディレクトリ）",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				_ = fs.Mkdir("/pnpm-lock.yaml", 0o755)
				return fs
			},
			path:    "/pnpm-lock.yaml",
			wantErr: &lockfile_err.FailedToLoadError{},
		},
		{
			name: "[異常系] パース失敗",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				_ = afero.WriteFile(fs, "/pnpm-lock.yaml", []byte("{invalid yaml"), 0o644)
				return fs
			},
			path:    "/pnpm-lock.yaml",
			wantErr: &lockfile_err.FailedToParseError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := tt.setupFs()
			got, gotErr := lockfile.Load(fs, tt.path)
			if d := cmp.Diff(tt.want, got); d != "" {
				t.Errorf("Load() mismatch (-want +got):\n%s", d)
			}
			if reflect.TypeOf(gotErr) != reflect.TypeOf(tt.wantErr) {
				t.Errorf("Load() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}
