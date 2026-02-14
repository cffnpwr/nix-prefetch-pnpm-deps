package logger

import (
	"fmt"
	"strings"
	"time"
)

// tuiCommandLogger implements CommandLogger for TUI mode.
type tuiCommandLogger struct {
	logger *tuiLogger
	start  time.Time
}

func (c *tuiCommandLogger) Write(p []byte) (int, error) {
	for line := range strings.SplitSeq(strings.TrimRight(string(p), "\n"), "\n") {
		c.logger.send(cmdLineMsg{line: line})
	}
	return len(p), nil
}

func (c *tuiCommandLogger) Done() {
	elapsed := time.Since(c.start)
	c.logger.send(cmdDoneMsg{elapsed: elapsed})
}

func (c *tuiCommandLogger) Fail(exitCode int) {
	elapsed := time.Since(c.start)
	c.logger.send(cmdFailMsg{exitCode: exitCode, elapsed: elapsed})
}

// noopCommandLogger is returned when log level is below threshold.
type noopCommandLogger struct{}

func (c *noopCommandLogger) Write(p []byte) (int, error) { return len(p), nil }
func (c *noopCommandLogger) Done()                       {}
func (c *noopCommandLogger) Fail(_ int)                  {}

// formatKV formats a message with key-value pairs for TUI display.
func formatKV(msg string, args ...any) string {
	if len(args) == 0 {
		return msg
	}
	var b strings.Builder
	b.WriteString(msg)
	for i := 0; i+1 < len(args); i += 2 {
		fmt.Fprintf(&b, " %v=%v", args[i], args[i+1])
	}
	return b.String()
}
