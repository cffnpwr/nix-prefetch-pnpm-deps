package logger

import (
	"bytes"
	"io"
	"log/slog"
	"reflect"
	"testing"

	"github.com/aymanbagabas/go-pty"
)

// fakeFdWriter implements io.Writer and fdWriter but is not a real TTY.
type fakeFdWriter struct {
	bytes.Buffer
}

func (f *fakeFdWriter) Fd() uintptr {
	return 0
}

func newPTY(t *testing.T) pty.Pty {
	t.Helper()
	p, err := pty.New()
	if err != nil {
		t.Fatalf("failed to open pty: %v", err)
	}
	t.Cleanup(func() { _ = p.Close() })
	return p
}

func Test_isTTY(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T) io.Writer
		want  bool
	}{
		{
			name: "[正常系] io.WriterがfdWriterを満たさない場合はfalseを返す",
			setup: func(_ *testing.T) io.Writer {
				return &bytes.Buffer{}
			},
			want: false,
		},
		{
			name: "[正常系] fdWriterだがTTYでない場合はfalseを返す",
			setup: func(_ *testing.T) io.Writer {
				return &fakeFdWriter{}
			},
			want: false,
		},
		{
			name: "[正常系] TTYの場合はtrueを返す",
			setup: func(t *testing.T) io.Writer {
				t.Helper()
				return newPTY(t)
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := tt.setup(t)

			got := isTTY(w)
			if got != tt.want {
				t.Errorf("isTTY() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newLogger(t *testing.T) {
	tests := []struct {
		name     string
		envs     map[string]string
		setup    func(t *testing.T) io.Writer
		wantType reflect.Type
	}{
		{
			name: "[正常系] io.WriterがfdWriterを満たさない場合はtextLoggerを返す",
			setup: func(_ *testing.T) io.Writer {
				return &bytes.Buffer{}
			},
			wantType: reflect.TypeFor[*textLogger](),
		},
		{
			name: "[正常系] fdWriterだがTTYでない場合はtextLoggerを返す",
			setup: func(_ *testing.T) io.Writer {
				return &fakeFdWriter{}
			},
			wantType: reflect.TypeFor[*textLogger](),
		},
		{
			name: "[正常系] TTYの場合はtuiLoggerを返す",
			setup: func(t *testing.T) io.Writer {
				t.Helper()
				return newPTY(t)
			},
			wantType: reflect.TypeFor[*tuiLogger](),
		},
		{
			name: "[正常系] CI環境かつTTYの場合はtextLoggerを返す",
			envs: map[string]string{"CI": "true"},
			setup: func(t *testing.T) io.Writer {
				t.Helper()
				return newPTY(t)
			},
			wantType: reflect.TypeFor[*textLogger](),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearCIEnvVars(t)
			for k, v := range tt.envs {
				t.Setenv(k, v)
			}
			w := tt.setup(t)

			got := newLogger(slog.LevelInfo, w)
			t.Cleanup(func() { got.Close() })

			gotType := reflect.TypeOf(got)
			if gotType != tt.wantType {
				t.Errorf("newLogger() returned %s, want %s", gotType, tt.wantType)
			}
		})
	}
}
