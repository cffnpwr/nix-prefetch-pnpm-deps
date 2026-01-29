package pnpm

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	pnpm_err "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/pnpm/errors"
)

func Test_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		wantPath string
		wantErr  pnpm_err.PnpmErrorIF
	}{
		{
			name:     "[正常系] 正しいパス文字列",
			path:     "/path/to/pnpm/bin/pnpm",
			wantPath: "/path/to/pnpm/bin/pnpm",
		},
		{
			name:    "[異常系] 空のパス文字列",
			path:    "",
			wantErr: &pnpm_err.PnpmNotFoundError{},
		},
		{
			name:    "[異常系] pnpmの実行ファイルが存在しないパス文字列",
			path:    "/invalid/path/to/pnpm/bin/pnpm",
			wantErr: &pnpm_err.PnpmNotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			if tt.wantErr == nil {
				// Create a dummy pnpm executable file in the in-memory filesystem
				err := afero.WriteFile(fs, tt.path, []byte{}, 0755)
				if err != nil {
					t.Fatalf("failed to create dummy pnpm executable: %v", err)
				}
			}

			got, gotErr := New(fs, tt.path)
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
		pathEnvVar string
		wantPath   string
		wantErr    pnpm_err.PnpmErrorIF
	}{
		{
			name:       "[正常系] PATHにpnpmが含まれている場合",
			pathEnvVar: "/path/to/pnpm/bin",
			wantPath:   "/path/to/pnpm/bin/pnpm",
		},
		{
			name:       "[正常系] PATHに複数のパスが含まれている場合",
			pathEnvVar: "/some/other/path:/path/to/pnpm/bin",
			wantPath:   "/path/to/pnpm/bin/pnpm",
		},
		{
			name:       "[異常系] PATHにpnpmが含まれていない場合",
			pathEnvVar: "/some/other/path",
			wantErr:    &pnpm_err.PnpmNotFoundError{},
		},
		{
			name:       "[異常系] 不正なパス文字列が含まれている場合",
			pathEnvVar: "",
			wantErr:    &pnpm_err.PnpmNotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PATH", tt.pathEnvVar)
			fs := afero.NewMemMapFs()
			if tt.wantPath != "" {
				// Create a dummy pnpm executable file in the in-memory filesystem
				err := afero.WriteFile(fs, tt.wantPath, []byte{}, 0755)
				if err != nil {
					t.Fatalf("failed to create dummy pnpm executable: %v", err)
				}
			}

			got, gotErr := WithPathEnvVar(fs)
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
