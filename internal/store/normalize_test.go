package store_test

import (
	"reflect"
	"testing"

	"github.com/spf13/afero"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/store"
	store_err "github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/store/errors"
)

func verifyDeleted(t *testing.T, afs afero.Fs, paths []string) {
	t.Helper()

	for _, p := range paths {
		exists, _ := afero.Exists(afs, p)
		if exists {
			t.Errorf("expected %s to be deleted, but it still exists", p)
		}
	}
}

func verifyFileContent(t *testing.T, afs afero.Fs, path string, want string) {
	t.Helper()

	data, err := afero.ReadFile(afs, path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}

	if string(data) != want {
		t.Errorf("content mismatch for %s\nwant: %q\ngot:  %q", path, want, string(data))
	}
}

type permCheck struct {
	path     string
	wantPerm uint32
}

func verifyPermissions(t *testing.T, afs afero.Fs, checks []permCheck) {
	t.Helper()

	for _, c := range checks {
		info, err := afs.Stat(c.path)
		if err != nil {
			t.Fatalf("Stat(%s) error: %v", c.path, err)
		}

		if uint32(info.Mode().Perm()) != c.wantPerm {
			t.Errorf("%s perm = %o, want %o", c.path, info.Mode().Perm(), c.wantPerm)
		}
	}
}

func Test_Normalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setupFs func() afero.Fs
		opts    store.NormalizeOptions
		wantErr store_err.StoreErrorIF
		verify  func(t *testing.T, afs afero.Fs)
	}{
		{
			name: "[正常系] tmpとprojectsディレクトリが削除される",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				fs.MkdirAll("/store/v3/tmp/some-dir", 0o755)
				afero.WriteFile(fs, "/store/v3/tmp/file.txt", []byte("temp"), 0o644)
				fs.MkdirAll("/store/v10/tmp", 0o755)
				fs.MkdirAll("/store/v3/projects/my-project", 0o755)
				fs.MkdirAll("/store/v10/projects", 0o755)
				return fs
			},
			opts: store.NormalizeOptions{StorePath: "/store", FetcherVersion: 1},
			verify: func(t *testing.T, afs afero.Fs) {
				t.Helper()
				verifyDeleted(t, afs, []string{
					"/store/v3/tmp",
					"/store/v10/tmp",
					"/store/v3/projects",
					"/store/v10/projects",
				})
			},
		},
		{
			name: "[正常系] JSONファイルのcheckedAtが削除されキーがソートされる",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				fs.MkdirAll("/store/v3", 0o755)
				afero.WriteFile(
					fs,
					"/store/v3/pkg.json",
					[]byte(`{"z":1,"a":2,"nested":{"checkedAt":123,"b":3},"checkedAt":456}`),
					0o644,
				)
				return fs
			},
			opts: store.NormalizeOptions{StorePath: "/store", FetcherVersion: 1},
			verify: func(t *testing.T, afs afero.Fs) {
				t.Helper()
				verifyFileContent(t, afs, "/store/v3/pkg.json",
					"{\n  \"a\": 2,\n  \"nested\": {\n    \"b\": 3\n  },\n  \"z\": 1\n}\n")
			},
		},
		{
			name: "[正常系] ネストされた配列内のcheckedAtも再帰的に削除される",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				fs.MkdirAll("/store/v3", 0o755)
				afero.WriteFile(
					fs,
					"/store/v3/deep.json",
					[]byte(`{"items":[{"checkedAt":1,"name":"a"},{"name":"b"}]}`),
					0o644,
				)
				return fs
			},
			opts: store.NormalizeOptions{StorePath: "/store", FetcherVersion: 1},
			verify: func(t *testing.T, afs afero.Fs) {
				t.Helper()
				verifyFileContent(
					t,
					afs,
					"/store/v3/deep.json",
					"{\n  \"items\": [\n    {\n      \"name\": \"a\"\n    },\n    {\n      \"name\": \"b\"\n    }\n  ]\n}\n",
				)
			},
		},
		{
			name: "[正常系] fetcherVersion 2以上でパーミッションが設定される",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				fs.MkdirAll("/store/v10/files", 0o755)
				afero.WriteFile(fs, "/store/v10/files/data.txt", []byte("data"), 0o644)
				afero.WriteFile(fs, "/store/v10/files/run-exec", []byte("exec"), 0o755)
				return fs
			},
			opts: store.NormalizeOptions{StorePath: "/store", FetcherVersion: 2},
			verify: func(t *testing.T, afs afero.Fs) {
				t.Helper()
				verifyPermissions(t, afs, []permCheck{
					{"/store/v10/files", 0o555},
					{"/store/v10/files/data.txt", 0o444},
					{"/store/v10/files/run-exec", 0o555},
				})
			},
		},
		{
			name: "[正常系] fetcherVersion 1ではパーミッション設定がスキップされる",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				fs.MkdirAll("/store/v3", 0o755)
				afero.WriteFile(fs, "/store/v3/data.txt", []byte("data"), 0o644)
				return fs
			},
			opts: store.NormalizeOptions{StorePath: "/store", FetcherVersion: 1},
			verify: func(t *testing.T, afs afero.Fs) {
				t.Helper()
				info, err := afs.Stat("/store/v3/data.txt")
				if err != nil {
					t.Fatalf("Stat() error: %v", err)
				}
				if info.Mode().Perm() != 0o644 {
					t.Errorf("file perm = %o, want 0644 (unchanged)", info.Mode().Perm())
				}
			},
		},
		{
			name: "[正常系] tmpやprojectsが存在しなくてもエラーにならない",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				fs.MkdirAll("/store", 0o755)
				return fs
			},
			opts: store.NormalizeOptions{StorePath: "/store", FetcherVersion: 1},
		},
		{
			name: "[異常系] 不正なJSONファイルがある場合",
			setupFs: func() afero.Fs {
				fs := afero.NewMemMapFs()
				fs.MkdirAll("/store/v3", 0o755)
				afero.WriteFile(fs, "/store/v3/invalid.json", []byte("{invalid"), 0o644)
				return fs
			},
			opts:    store.NormalizeOptions{StorePath: "/store", FetcherVersion: 1},
			wantErr: &store_err.FailedToNormalizeJSONError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			afs := tt.setupFs()
			gotErr := store.Normalize(afs, tt.opts)

			if reflect.TypeOf(gotErr) != reflect.TypeOf(tt.wantErr) {
				t.Errorf("Normalize() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}

			if tt.verify != nil {
				tt.verify(t, afs)
			}
		})
	}
}
