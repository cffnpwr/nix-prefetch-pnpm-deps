package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"
)

// textHandler is a custom slog.Handler that outputs logs in the format:
//
//	2006-01-02T15:04:05+07:00 LEVEL [scope="name" ]msg [key=value ...]
//
// Attributes named "scope" added via WithAttrs are placed before the message,
// while all other attributes appear after it.
type textHandler struct {
	w        io.Writer
	mu       *sync.Mutex
	level    slog.Level
	preAttrs []slog.Attr
}

func newTextHandler(w io.Writer, level slog.Level) *textHandler {
	return &textHandler{
		w:     w,
		mu:    &sync.Mutex{},
		level: level,
	}
}

func (h *textHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *textHandler) Handle(_ context.Context, r slog.Record) error {
	var buf []byte

	// Timestamp
	buf = append(buf, r.Time.Format(time.RFC3339)...)
	buf = append(buf, ' ')

	// Level
	buf = append(buf, r.Level.String()...)
	buf = append(buf, ' ')

	// Pre-attributes (scope) before message
	for _, a := range h.preAttrs {
		if a.Key == "scope" {
			buf = append(buf, fmt.Sprintf("scope=%q ", a.Value.String())...)
		}
	}

	// Message
	buf = append(buf, r.Message...)

	// Post-attributes: non-scope preAttrs + record attrs
	for _, a := range h.preAttrs {
		if a.Key != "scope" {
			buf = append(buf, ' ')
			buf = append(buf, formatAttr(a)...)
		}
	}
	r.Attrs(func(a slog.Attr) bool {
		buf = append(buf, ' ')
		buf = append(buf, formatAttr(a)...)
		return true
	})

	buf = append(buf, '\n')

	h.mu.Lock()
	defer h.mu.Unlock()

	_, err := h.w.Write(buf)

	return err
}

func (h *textHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.preAttrs), len(h.preAttrs)+len(attrs))
	copy(newAttrs, h.preAttrs)
	newAttrs = append(newAttrs, attrs...)

	return &textHandler{
		w:        h.w,
		mu:       h.mu,
		level:    h.level,
		preAttrs: newAttrs,
	}
}

func (h *textHandler) WithGroup(_ string) slog.Handler {
	// Group support is not needed for this project.
	return h
}

func formatAttr(a slog.Attr) string {
	return fmt.Sprintf("%s=%s", a.Key, a.Value)
}
