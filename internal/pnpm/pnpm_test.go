package pnpm

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/logger"
	pnpm_err "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/pnpm/errors"
)

func Test_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setupFs  func() afero.Fs
		path     string
		wantPath string
		wantErr  pnpm_err.PnpmErrorIF
	}{
		{
			name: "[正常系] 正しいパス文字列",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				afero.WriteFile(fs, "/path/to/pnpm/bin/pnpm", []byte{}, 0755)
				return fs
			},
			path:     "/path/to/pnpm/bin/pnpm",
			wantPath: "/path/to/pnpm/bin/pnpm",
		},
		{
			name:    "[異常系] 空のパス文字列",
			setupFs: afero.NewMemMapFs,
			path:    "",
			wantErr: &pnpm_err.PnpmNotFoundError{},
		},
		{
			name:    "[異常系] pnpmの実行ファイルが存在しないパス文字列",
			setupFs: afero.NewMemMapFs,
			path:    "/invalid/path/to/pnpm/bin/pnpm",
			wantErr: &pnpm_err.PnpmNotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := tt.setupFs()

			l := logger.New(slog.LevelError)
			t.Cleanup(func() { l.Close() })

			got, gotErr := New(fs, l, tt.path)
			if got != nil {
				if d := cmp.Diff(tt.wantPath, got.path); d != "" {
					t.Errorf("New() path mismatch (-want +got):\n%s", d)
				}
			}
			if reflect.TypeOf(gotErr) != reflect.TypeOf(tt.wantErr) {
				t.Errorf("New() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func Test_WithPathEnvVar(t *testing.T) {
	tests := []struct {
		name       string
		setupFs    func() afero.Fs
		pathEnvVar string
		wantPath   string
		wantErr    pnpm_err.PnpmErrorIF
	}{
		{
			name: "[正常系] PATHにpnpmが含まれている場合",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				afero.WriteFile(fs, "/path/to/pnpm/bin/pnpm", []byte{}, 0755)
				return fs
			},
			pathEnvVar: "/path/to/pnpm/bin",
			wantPath:   "/path/to/pnpm/bin/pnpm",
		},
		{
			name: "[正常系] PATHに複数のパスが含まれている場合",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				afero.WriteFile(fs, "/path/to/pnpm/bin/pnpm", []byte{}, 0755)
				return fs
			},
			pathEnvVar: "/some/other/path:/path/to/pnpm/bin",
			wantPath:   "/path/to/pnpm/bin/pnpm",
		},
		{
			name:       "[異常系] PATHにpnpmが含まれていない場合",
			setupFs:    afero.NewMemMapFs,
			pathEnvVar: "/some/other/path",
			wantErr:    &pnpm_err.PnpmNotFoundError{},
		},
		{
			name:       "[異常系] PATHが空の場合",
			setupFs:    afero.NewMemMapFs,
			pathEnvVar: "",
			wantErr:    &pnpm_err.PnpmNotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PATH", tt.pathEnvVar)
			fs := tt.setupFs()

			l := logger.New(slog.LevelError)
			t.Cleanup(func() { l.Close() })

			got, gotErr := WithPathEnvVar(fs, l)
			if got != nil {
				if d := cmp.Diff(tt.wantPath, got.path); d != "" {
					t.Errorf("WithPathEnvVar() mismatch (-want +got):\n%s", d)
				}
			}
			if reflect.TypeOf(gotErr) != reflect.TypeOf(tt.wantErr) {
				t.Errorf("WithPathEnvVar() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}
