package logger

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"
)

func Test_textHandler_Handle(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.FixedZone("", 7*60*60))

	tests := []struct {
		name     string
		preAttrs []slog.Attr
		record   func() slog.Record
		want     string
	}{
		{
			name: "[正常系] 基本的なメッセージが出力される",
			record: func() slog.Record {
				r := slog.NewRecord(fixedTime, slog.LevelInfo, "hello", 0)
				return r
			},
			want: "2006-01-02T15:04:05+07:00 INFO hello\n",
		},
		{
			name: "[正常系] scope属性がメッセージの前に出力される",
			preAttrs: []slog.Attr{
				slog.String("scope", "pnpm install"),
			},
			record: func() slog.Record {
				r := slog.NewRecord(fixedTime, slog.LevelInfo, "installing packages", 0)
				return r
			},
			want: "2006-01-02T15:04:05+07:00 INFO scope=\"pnpm install\" installing packages\n",
		},
		{
			name: "[正常系] レコード属性がメッセージの後に出力される",
			record: func() slog.Record {
				r := slog.NewRecord(fixedTime, slog.LevelError, "failed", 0)
				r.AddAttrs(slog.String("error", "timeout"))
				return r
			},
			want: "2006-01-02T15:04:05+07:00 ERROR failed error=timeout\n",
		},
		{
			name: "[正常系] scope属性はメッセージ前でその他preAttrsはメッセージ後に出力される",
			preAttrs: []slog.Attr{
				slog.String("scope", "build"),
				slog.String("env", "production"),
			},
			record: func() slog.Record {
				r := slog.NewRecord(fixedTime, slog.LevelWarn, "warning", 0)
				return r
			},
			want: "2006-01-02T15:04:05+07:00 WARN scope=\"build\" warning env=production\n",
		},
		{
			name: "[正常系] preAttrsとレコード属性の両方がメッセージ後に出力される",
			preAttrs: []slog.Attr{
				slog.String("scope", "test"),
				slog.String("module", "auth"),
			},
			record: func() slog.Record {
				r := slog.NewRecord(fixedTime, slog.LevelInfo, "ok", 0)
				r.AddAttrs(slog.String("user", "alice"))
				return r
			},
			want: "2006-01-02T15:04:05+07:00 INFO scope=\"test\" ok module=auth user=alice\n",
		},
		{
			name: "[正常系] DEBUGレベルが正しく出力される",
			record: func() slog.Record {
				r := slog.NewRecord(fixedTime, slog.LevelDebug, "debug msg", 0)
				return r
			},
			want: "2006-01-02T15:04:05+07:00 DEBUG debug msg\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := &bytes.Buffer{}
			h := newTextHandler(w, slog.LevelDebug)
			if tt.preAttrs != nil {
				h = h.WithAttrs(tt.preAttrs).(*textHandler)
			}

			r := tt.record()
			err := h.Handle(context.Background(), r)

			if err != nil {
				t.Fatalf("Handle() returned unexpected error: %v", err)
			}

			got := w.String()
			if got != tt.want {
				t.Errorf("Handle() output = %q, want %q", got, tt.want)
			}
		})
	}
}
