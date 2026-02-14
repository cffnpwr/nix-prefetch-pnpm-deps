package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/mattn/go-isatty"
)

type LogLevel = slog.Level

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Fatal(msg string, args ...any)

	Debugf(tmpl string, args ...any)
	Infof(tmpl string, args ...any)
	Warnf(tmpl string, args ...any)
	Errorf(tmpl string, args ...any)
	Fatalf(tmpl string, args ...any)

	StepLogger(logLevel LogLevel, msg string) StepLogger
	CommandLogger(logLevel LogLevel, name string) CommandLogger

	Close() error
}

type StepLogger interface {
	Done()
	Fail(err error)
}

type CommandLogger interface {
	io.Writer
	Done()
	Fail(exitCode int)
}

type fdWriter interface {
	Fd() uintptr
}

func New(level LogLevel) Logger {
	return newLogger(level, os.Stdout)
}

func newLogger(level LogLevel, w io.Writer) Logger {
	ci, isCI := detectCI()
	// Use TUI Logger in non-CI environments and when `w` is TTY.
	if isTTY(w) && !isCI {
		return newTUILogger(level, w)
	}

	return newTextLogger(level, w, ci)
}

func isTTY(w io.Writer) bool {
	if f, ok := w.(fdWriter); ok {
		if isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd()) {
			return true
		}
	}
	return false
}
