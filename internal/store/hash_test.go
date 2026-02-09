package store_test

import (
	"encoding/base64"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/store"
	store_err "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/store/errors"
)

const sha256DigestLen = 32

func assertSRIFormat(t *testing.T, got string) {
	t.Helper()

	if !strings.HasPrefix(got, "sha256-") {
		t.Errorf("Hash() = %q, want prefix \"sha256-\"", got)
		return
	}

	b64Part := strings.TrimPrefix(got, "sha256-")

	decoded, err := base64.StdEncoding.DecodeString(b64Part)
	if err != nil {
		t.Errorf("Hash() base64 decode error: %v", err)
		return
	}

	if len(decoded) != sha256DigestLen {
		t.Errorf("Hash() digest length = %d, want %d", len(decoded), sha256DigestLen)
	}
}

func computeHash(t *testing.T, setup func(afero.Fs)) string {
	t.Helper()

	afs := afero.NewMemMapFs()
	setup(afs)

	got, err := store.Hash(afs, "/store")
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}

	return got
}

func Test_Hash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setupFs func() afero.Fs
		path    string
		wantErr store_err.StoreErrorIF
		verify  func(t *testing.T, got string)
	}{
		{
			name: "[正常系] ディレクトリのハッシュがSRI形式で返される",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				fs.MkdirAll("/store", 0o555)
				afero.WriteFile(fs, "/store/file.txt", []byte("hello"), 0o444)
				return fs
			},
			path:   "/store",
			verify: assertSRIFormat,
		},
		{
			name: "[正常系] 同じ内容で同じハッシュが返される",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				fs.MkdirAll("/store", 0o555)
				afero.WriteFile(fs, "/store/file.txt", []byte("hello"), 0o444)
				return fs
			},
			path: "/store",
			verify: func(t *testing.T, got string) {
				t.Helper()
				got2 := computeHash(t, func(fs afero.Fs) {
					fs.MkdirAll("/store", 0o555)
					afero.WriteFile(fs, "/store/file.txt", []byte("hello"), 0o444)
				})
				if got != got2 {
					t.Errorf("Hash() not deterministic: first=%q, second=%q", got, got2)
				}
			},
		},
		{
			name: "[正常系] 異なる内容で異なるハッシュが返される",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				fs.MkdirAll("/store", 0o555)
				afero.WriteFile(fs, "/store/file.txt", []byte("hello"), 0o444)
				return fs
			},
			path: "/store",
			verify: func(t *testing.T, got string) {
				t.Helper()
				got2 := computeHash(t, func(fs afero.Fs) {
					fs.MkdirAll("/store", 0o555)
					afero.WriteFile(fs, "/store/file.txt", []byte("world"), 0o444)
				})
				if got == got2 {
					t.Errorf("Hash() should differ for different content, both=%q", got)
				}
			},
		},
		{
			name: "[正常系] 実行可能フラグがハッシュに影響する",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				fs.MkdirAll("/store", 0o555)
				afero.WriteFile(fs, "/store/script", []byte("#!/bin/sh"), 0o444)
				return fs
			},
			path: "/store",
			verify: func(t *testing.T, got string) {
				t.Helper()
				got2 := computeHash(t, func(fs afero.Fs) {
					fs.MkdirAll("/store", 0o555)
					afero.WriteFile(fs, "/store/script", []byte("#!/bin/sh"), 0o555)
				})
				if got == got2 {
					t.Errorf("Hash() should differ for different executable flag, both=%q", got)
				}
			},
		},
		{
			name: "[正常系] ネストされたディレクトリ構造のハッシュが計算できる",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				fs.MkdirAll("/store/sub/deep", 0o555)
				afero.WriteFile(fs, "/store/a.txt", []byte("a"), 0o444)
				afero.WriteFile(fs, "/store/sub/b.txt", []byte("b"), 0o444)
				afero.WriteFile(fs, "/store/sub/deep/c.txt", []byte("c"), 0o444)
				return fs
			},
			path:   "/store",
			verify: assertSRIFormat,
		},
		{
			name:    "[異常系] 存在しないパス",
			setupFs: afero.NewMemMapFs,
			path:    "/nonexistent",
			wantErr: &store_err.FailedToHashError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			afs := tt.setupFs()
			got, gotErr := store.Hash(afs, tt.path)

			if reflect.TypeOf(gotErr) != reflect.TypeOf(tt.wantErr) {
				t.Errorf("Hash() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}

			if tt.verify != nil {
				tt.verify(t, got)
			}
		})
	}
}
