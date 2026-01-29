//go:build unix

package path_test

import (
	"testing"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/path"
)

func Test_IsPath_Unix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "[正常系] 絶対パス",
			s:    "/path/to/file",
			want: true,
		},
		{
			name: "[正常系] ルートディレクトリ",
			s:    "/",
			want: true,
		},
		{
			name: "[正常系] スラッシュ複数連続",
			s:    "///foo/bar",
			want: true,
		},
		{
			name: "[正常系] 山括弧を含む",
			s:    "/path<with>file",
			want: true,
		},
		{
			name: "[正常系] コロンを含む",
			s:    "/path:file",
			want: true,
		},
		{
			name: "[正常系] パイプを含む",
			s:    "/path|file",
			want: true,
		},
		{
			name: "[正常系] クエスチョンを含む",
			s:    "/path?file",
			want: true,
		},
		{
			name: "[正常系] アスタリスクを含む",
			s:    "/path*file",
			want: true,
		},
		{
			name: "[正常系] ダブルクォートを含む",
			s:    "/path\"file",
			want: true,
		},
		{
			name: "[正常系] バックスラッシュを含む",
			s:    "/path\\file",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := path.IsPath(tt.s)
			if got != tt.want {
				t.Errorf("IsPath(%q) = %v; want %v", tt.s, got, tt.want)
			}
		})
	}
}
