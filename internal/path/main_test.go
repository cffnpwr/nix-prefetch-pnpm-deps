package path_test

import (
	"testing"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/path"
)

func Test_IsPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "[正常系] カレントディレクトリ",
			s:    ".",
			want: true,
		},
		{
			name: "[正常系] 親ディレクトリ",
			s:    "..",
			want: true,
		},
		{
			name: "[正常系] 相対パス (./)",
			s:    "./foo/bar",
			want: true,
		},
		{
			name: "[正常系] 相対パス (../)",
			s:    "../foo/bar",
			want: true,
		},
		{
			name: "[正常系] 単純なファイル名",
			s:    "file.txt",
			want: true,
		},
		{
			name: "[正常系] 空白を含むパス",
			s:    "path with spaces/file",
			want: true,
		},
		{
			name: "[異常系] 空文字列",
			s:    "",
			want: false,
		},
		{
			name: "[異常系] NUL文字を含む",
			s:    "/path\x00/file",
			want: false,
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
