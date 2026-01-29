//go:build windows

package path

import (
	"testing"
)

func Test_IsPath_Windows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "[正常系] ドライブレター (バックスラッシュ)",
			s:    `C:\Users\test`,
			want: true,
		},
		{
			name: "[正常系] ドライブレター (スラッシュ)",
			s:    "C:/Users/test",
			want: true,
		},
		{
			name: "[正常系] ドライブ相対パス",
			s:    `C:foo\bar`,
			want: true,
		},
		{
			name: "[正常系] 小文字ドライブレター",
			s:    `c:\users`,
			want: true,
		},
		{
			name: "[正常系] UNCパス (バックスラッシュ)",
			s:    `\\server\share\path`,
			want: true,
		},
		{
			name: "[正常系] UNCパス (スラッシュ)",
			s:    "//server/share/path",
			want: true,
		},
		{
			name: "[正常系] デバイスパス (dot)",
			s:    `\\.\COM1`,
			want: true,
		},
		{
			name: "[正常系] デバイスパス (question)",
			s:    `\\?\very\long\path`,
			want: true,
		},
		{
			name: "[正常系] デバイスパス + ドライブレター",
			s:    `\\?\C:\very\long\path`,
			want: true,
		},
		{
			name: "[正常系] 予約名 CON",
			s:    "CON",
			want: true,
		},
		{
			name: "[正常系] 予約名をパスに含む",
			s:    `C:\folder\CON`,
			want: true,
		},
		{
			name: "[正常系] 相対パス (バックスラッシュ)",
			s:    `.\foo\bar`,
			want: true,
		},
		{
			name: "[異常系] 禁止文字 <",
			s:    `C:\path<file`,
			want: false,
		},
		{
			name: "[異常系] 禁止文字 >",
			s:    `C:\path>file`,
			want: false,
		},
		{
			name: "[異常系] 禁止文字 \"",
			s:    `C:\path"file`,
			want: false,
		},
		{
			name: "[異常系] 禁止文字 |",
			s:    `C:\path|file`,
			want: false,
		},
		{
			name: "[異常系] 禁止文字 ? (パス内)",
			s:    `C:\path?\file`,
			want: false,
		},
		{
			name: "[異常系] 禁止文字 *",
			s:    `C:\path*file`,
			want: false,
		},
		{
			name: "[異常系] 制御文字 (0x01)",
			s:    `C:\path\x01file`,
			want: false,
		},
		{
			name: "[異常系] 制御文字 (0x1F)",
			s:    `C:\path\x1Ffile`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsPath(tt.s)
			if got != tt.want {
				t.Errorf("IsPath(%q) = %v; want %v", tt.s, got, tt.want)
			}
		})
	}
}

func Test_extractDriveLetter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		s        string
		want     string
		wantBool bool
	}{
		{
			name:     "[正常系] 有効なドライブレター",
			s:        `C:\path\to\file`,
			want:     "C:",
			wantBool: true,
		},
		{
			name:     "[正常系] 小文字ドライブレター",
			s:        "d:/another/path",
			want:     "d:",
			wantBool: true,
		},
		{
			name:     "[異常系] 空文字",
			s:        "",
			want:     "",
			wantBool: false,
		},
		{
			name:     "[異常系] ドライブレターなし",
			s:        `relative\path`,
			want:     "",
			wantBool: false,
		},
		{
			name:     "[異常系] 不正なドライブレター (数字)",
			s:        `1:\invalid\drive`,
			want:     "",
			wantBool: false,
		},
		{
			name:     "[異常系] 不正なドライブレター (記号)",
			s:        `%:\invalid\drive`,
			want:     "",
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, gotBool := extractDriveLetter(tt.s)
			if got != tt.want || gotBool != tt.wantBool {
				t.Errorf(
					"ExtractDriveLetter(%q) = (%q, %v); want (%q, %v)",
					tt.s,
					got,
					gotBool,
					tt.want,
					tt.wantBool,
				)
			}
		})
	}
}

func Test_extractDevicePathPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		s        string
		want     string
		wantBool bool
	}{
		{
			name:     "[正常系] デバイスパス (dot)",
			s:        `\\.\COM1`,
			want:     `\\.\`,
			wantBool: true,
		},
		{
			name:     "[正常系] デバイスパス (question)",
			s:        `\\?\C:\very\long\path`,
			want:     `\\?\`,
			wantBool: true,
		},
		{
			name:     "[異常系] 空文字",
			s:        "",
			want:     "",
			wantBool: false,
		},
		{
			name:     "[異常系] デバイスパス接頭辞なし",
			s:        `C:\normal\path`,
			want:     "",
			wantBool: false,
		},
		{
			name:     "[異常系] デバイスパスなし",
			s:        `relative\path`,
			want:     "",
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, gotBool := extractDevicePathPrefix(tt.s)
			if got != tt.want || gotBool != tt.wantBool {
				t.Errorf(
					"ExtractDevicePathPrefix(%q) = (%q, %v); want (%q, %v)",
					tt.s,
					got,
					gotBool,
					tt.want,
					tt.wantBool,
				)
			}
		})
	}
}
