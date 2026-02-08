package common

import (
	"testing"
)

func Test_MajorVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		want    int
		wantErr bool
	}{
		{
			name:    "[正常系] 9.15.0 → 9",
			version: "9.15.0",
			want:    9,
		},
		{
			name:    "[正常系] 6.1 → 6",
			version: "6.1",
			want:    6,
		},
		{
			name:    "[正常系] 10.0 → 10",
			version: "10.0",
			want:    10,
		},
		{
			name:    "[正常系] 改行付き",
			version: "9.15.0\n",
			want:    9,
		},
		{
			name:    "[異常系] 空文字",
			version: "",
			wantErr: true,
		},
		{
			name:    "[異常系] 数値でない",
			version: "abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := MajorVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("MajorVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MajorVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
