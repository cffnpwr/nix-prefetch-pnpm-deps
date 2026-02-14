package logger

import "testing"

func Test_formatKV(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  string
		args []any
		want string
	}{
		{
			name: "[正常系] 引数なしの場合はメッセージのみ返す",
			msg:  "hello",
			args: nil,
			want: "hello",
		},
		{
			name: "[正常系] キーバリューペアがフォーマットされる",
			msg:  "loaded",
			args: []any{"path", "/pnpm-lock.yaml"},
			want: "loaded path=/pnpm-lock.yaml",
		},
		{
			name: "[正常系] 複数のキーバリューペアがフォーマットされる",
			msg:  "status",
			args: []any{"resolved", 123, "downloaded", 45},
			want: "status resolved=123 downloaded=45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatKV(tt.msg, tt.args...)
			if got != tt.want {
				t.Errorf("formatKV() = %q, want %q", got, tt.want)
			}
		})
	}
}
