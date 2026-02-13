package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

type textLogger struct {
	logger *slog.Logger
	level  LogLevel
	w      io.Writer
	ci     CI
}

func newTextLogger(level LogLevel, w io.Writer, ci CI) Logger {
	handler := newTextHandler(w, level)
	return &textLogger{
		logger: slog.New(handler),
		level:  level,
		w:      w,
		ci:     ci,
	}
}

func (l *textLogger) Debug(msg string, args ...any) { l.logger.Debug(msg, args...) }
func (l *textLogger) Info(msg string, args ...any)  { l.logger.Info(msg, args...) }
func (l *textLogger) Warn(msg string, args ...any)  { l.logger.Warn(msg, args...) }
func (l *textLogger) Error(msg string, args ...any) { l.logger.Error(msg, args...) }

func (l *textLogger) Fatal(msg string, args ...any) {
	l.logger.Error(msg, args...)
	os.Exit(1)
}

func (l *textLogger) Debugf(tmpl string, args ...any) { l.logger.Debug(fmt.Sprintf(tmpl, args...)) }
func (l *textLogger) Infof(tmpl string, args ...any)  { l.logger.Info(fmt.Sprintf(tmpl, args...)) }
func (l *textLogger) Warnf(tmpl string, args ...any)  { l.logger.Warn(fmt.Sprintf(tmpl, args...)) }
func (l *textLogger) Errorf(tmpl string, args ...any) { l.logger.Error(fmt.Sprintf(tmpl, args...)) }

func (l *textLogger) Fatalf(tmpl string, args ...any) {
	l.logger.Error(fmt.Sprintf(tmpl, args...))
	os.Exit(1)
}

func (l *textLogger) StepLogger(logLevel LogLevel, msg string) StepLogger {
	start := time.Now()
	l.log(logLevel, "start "+msg)

	return &textStepLogger{
		logger:   l,
		logLevel: logLevel,
		msg:      msg,
		start:    start.Round(0), // Remove monotonic time
	}
}

func (l *textLogger) CommandLogger(logLevel LogLevel, name string) CommandLogger {
	start := time.Now()
	cl := &textCommandLogger{
		logger:       l,
		scopedLogger: l.logger.With("scope", name),
		logLevel:     logLevel,
		name:         name,
		start:        start.Round(0), // Remove monotonic time
	}
	cl.writeFoldStart()
	return cl
}

func (l *textLogger) Close() error {
	return nil
}

func (l *textLogger) log(level LogLevel, msg string, args ...any) {
	l.logger.Log(context.TODO(), level, msg, args...)
}

// textStepLogger implements StepLogger for text mode.
type textStepLogger struct {
	logger   *textLogger
	logLevel LogLevel
	msg      string
	start    time.Time
}

func (s *textStepLogger) Done() {
	elapsed := time.Since(s.start)
	s.logger.log(s.logLevel, fmt.Sprintf("%s completed in %s", s.msg, elapsed))
}

func (s *textStepLogger) Fail(err error) {
	elapsed := time.Since(s.start)
	s.logger.logger.Error(fmt.Sprintf("%s failed in %s", s.msg, elapsed), "error", err)
}

// textCommandLogger implements CommandLogger for text mode.
type textCommandLogger struct {
	logger       *textLogger
	scopedLogger *slog.Logger
	logLevel     LogLevel
	name         string
	start        time.Time
}

func (c *textCommandLogger) Write(p []byte) (n int, err error) {
	for line := range strings.SplitSeq(strings.Trim(string(p), "\n"), "\n") {
		c.scopedLogger.Info(line)
	}
	return len(p), nil
}

func (c *textCommandLogger) Done() {
	elapsed := time.Since(c.start)
	c.logger.log(c.logLevel, fmt.Sprintf("%s completed in %s", c.name, elapsed))
	c.writeFoldEnd()
}

func (c *textCommandLogger) Fail(exitCode int) {
	elapsed := time.Since(c.start)
	c.logger.logger.Error(
		fmt.Sprintf("%s failed with exit code %d in %s", c.name, exitCode, elapsed),
	)
	c.writeFoldEnd()
}

// foldID returns a sanitized identifier for CI folding syntax.
func (c *textCommandLogger) foldID() string {
	return strings.ReplaceAll(c.name, " ", "_")
}

func (c *textCommandLogger) writeFoldStart() {
	ci := c.logger.ci
	switch ci {
	case gitHubActions:
		fmt.Fprintf(c.logger.w, "::group::%s\n", c.name)
	case gitLabCI:
		fmt.Fprintf(
			c.logger.w,
			"\x1b[0Ksection_start:%d:%s\r\x1b[0K%s\n",
			time.Now().Unix(),
			c.foldID(),
			c.name,
		)
	case azurePipelines:
		fmt.Fprintf(c.logger.w, "##[group]%s\n", c.name)
	case teamCity:
		fmt.Fprintf(c.logger.w, "##teamcity[blockOpened name='%s']\n", c.name)
	case buildkite:
		fmt.Fprintf(c.logger.w, "--- %s\n", c.name)
	case travisCI:
		fmt.Fprintf(c.logger.w, "travis_fold:start:%s\n%s\n", c.foldID(), c.name)
	default:
		c.logger.log(c.logLevel, fmt.Sprintf("start %s", c.name))
	}
}

func (c *textCommandLogger) writeFoldEnd() {
	ci := c.logger.ci
	switch ci {
	case gitHubActions:
		fmt.Fprintln(c.logger.w, "::endgroup::")
	case gitLabCI:
		fmt.Fprintf(
			c.logger.w,
			"\x1b[0Ksection_end:%d:%s\r\x1b[0K\n",
			time.Now().Unix(),
			c.foldID(),
		)
	case azurePipelines:
		fmt.Fprintln(c.logger.w, "##[endgroup]")
	case teamCity:
		fmt.Fprintf(c.logger.w, "##teamcity[blockClosed name='%s']\n", c.name)
	case travisCI:
		fmt.Fprintf(c.logger.w, "travis_fold:end:%s\n", c.foldID())
	}
}
