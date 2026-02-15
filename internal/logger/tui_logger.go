package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// tuiLogger implements Logger with Bubble Tea TUI.
type tuiLogger struct {
	program   *tea.Program
	level     LogLevel
	mu        sync.Mutex
	done      chan struct{}
	closeOnce sync.Once
}

func newTUILogger(level LogLevel, w io.Writer) Logger {
	m := newTUIModel()
	p := tea.NewProgram(m, tea.WithOutput(w), tea.WithInput(nil), tea.WithoutSignalHandler())
	done := make(chan struct{})
	l := &tuiLogger{
		program: p,
		level:   level,
		done:    done,
	}
	go func() { _, _ = p.Run() }()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		select {
		case <-sigCh:
			signal.Stop(sigCh)
			l.send(interruptMsg{})
			l.program.Wait()
			os.Exit(130) //nolint:mnd // 128 + SIGINT(2)
		case <-done:
			signal.Stop(sigCh)
		}
	}()

	return l
}

func (l *tuiLogger) Debug(msg string, args ...any) {
	if !l.enabled(slog.LevelDebug) {
		return
	}
	l.send(logLineMsg{line: formatKV(msg, args...)})
}

func (l *tuiLogger) Info(msg string, args ...any) {
	if !l.enabled(slog.LevelInfo) {
		return
	}
	l.send(logLineMsg{line: doneStyle.Render("✓") + " " + formatKV(msg, args...)})
}

func (l *tuiLogger) Warn(msg string, args ...any) {
	if !l.enabled(slog.LevelWarn) {
		return
	}
	l.send(logLineMsg{line: warnStyle.Render("!") + " " + formatKV(msg, args...)})
}

func (l *tuiLogger) Error(msg string, args ...any) {
	if !l.enabled(slog.LevelError) {
		return
	}
	l.send(logLineMsg{line: failStyle.Render("✖") + " " + formatKV(msg, args...)})
}

func (l *tuiLogger) Fatal(msg string, args ...any) {
	l.send(logLineMsg{line: failStyle.Render("✖") + " " + formatKV(msg, args...)})
	_ = l.Close()
	os.Exit(1)
}

func (l *tuiLogger) Debugf(tmpl string, args ...any) {
	if !l.enabled(slog.LevelDebug) {
		return
	}
	l.send(logLineMsg{line: fmt.Errorf(tmpl, args...).Error()})
}

func (l *tuiLogger) Infof(tmpl string, args ...any) {
	if !l.enabled(slog.LevelInfo) {
		return
	}
	l.send(logLineMsg{line: doneStyle.Render("✓") + " " + fmt.Errorf(tmpl, args...).Error()})
}

func (l *tuiLogger) Warnf(tmpl string, args ...any) {
	if !l.enabled(slog.LevelWarn) {
		return
	}
	l.send(logLineMsg{line: warnStyle.Render("!") + " " + fmt.Errorf(tmpl, args...).Error()})
}

func (l *tuiLogger) Errorf(tmpl string, args ...any) {
	if !l.enabled(slog.LevelError) {
		return
	}
	l.send(logLineMsg{line: failStyle.Render("✖") + " " + fmt.Errorf(tmpl, args...).Error()})
}

func (l *tuiLogger) Fatalf(tmpl string, args ...any) {
	l.send(logLineMsg{line: failStyle.Render("✖") + " " + fmt.Errorf(tmpl, args...).Error()})
	_ = l.Close()
	os.Exit(1)
}

func (l *tuiLogger) StepLogger(logLevel LogLevel, msg string) StepLogger {
	if !l.enabled(logLevel) {
		return &noopStepLogger{}
	}
	l.send(stepStartMsg{msg: msg})
	return &tuiStepLogger{logger: l}
}

func (l *tuiLogger) CommandLogger(logLevel LogLevel, name string) CommandLogger {
	if !l.enabled(logLevel) {
		return &noopCommandLogger{}
	}
	l.send(cmdStartMsg{name: name})
	return &tuiCommandLogger{logger: l, start: time.Now().Round(0)}
}

func (l *tuiLogger) Close() error {
	l.closeOnce.Do(func() {
		close(l.done)
		l.program.Quit()
		l.program.Wait()
	})
	return nil
}

func (l *tuiLogger) send(msg tea.Msg) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.program.Send(msg)
}

func (l *tuiLogger) enabled(level LogLevel) bool {
	return level >= l.level
}
