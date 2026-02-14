package logger

import (
	"bytes"
	"errors"
	"log/slog"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func Test_newTextLogger(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
	}{
		{
			name:  "[正常系] Debugレベルで初期化される",
			level: slog.LevelDebug,
		},
		{
			name:  "[正常系] Infoレベルで初期化される",
			level: slog.LevelInfo,
		},
		{
			name:  "[正常系] Warnレベルで初期化される",
			level: slog.LevelWarn,
		},
		{
			name:  "[正常系] Errorレベルで初期化される",
			level: slog.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearCIEnvVars(t)

			w := &bytes.Buffer{}
			got := newTextLogger(tt.level, w, CI(""))

			tl, ok := got.(*textLogger)
			if !ok {
				t.Fatalf("newTextLogger() returned %T, want *textLogger", got)
			}
			if tl.level != tt.level {
				t.Errorf("newTextLogger().level = %v, want %v", tl.level, tt.level)
			}
			if tl.w != w {
				t.Errorf("newTextLogger().w does not match the provided writer")
			}
			if tl.logger == nil {
				t.Errorf("newTextLogger().logger is nil")
			}
		})
	}
}

func Test_textLogger_StepLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		logLevel LogLevel
		msg      string
		wantType reflect.Type
		wantLog  string
	}{
		{
			name:     "[正常系] StepLoggerが生成されログが出力される",
			logLevel: slog.LevelInfo,
			msg:      "fetching dependencies",
			wantType: reflect.TypeFor[*textStepLogger](),
			wantLog:  "fetching dependencies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := &bytes.Buffer{}
			l := newTextLogger(slog.LevelDebug, w, CI("")).(*textLogger)

			got := l.StepLogger(tt.logLevel, tt.msg)

			gotType := reflect.TypeOf(got)
			if gotType != tt.wantType {
				t.Errorf("StepLogger() returned %s, want %s", gotType, tt.wantType)
			}

			output := w.String()
			if !strings.Contains(output, tt.wantLog) {
				t.Errorf("StepLogger() output = %q, want to contain %q", output, tt.wantLog)
			}
		})
	}
}

func Test_textStepLogger_Done(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		msg     string
		wantLog string
	}{
		{
			name:    "[正常系] 完了ログが出力される",
			msg:     "fetching dependencies",
			wantLog: "fetching dependencies completed in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := &bytes.Buffer{}
			l := newTextLogger(slog.LevelDebug, w, CI("")).(*textLogger)
			sl := l.StepLogger(slog.LevelInfo, tt.msg)

			w.Reset()
			sl.Done()

			output := w.String()
			if !strings.Contains(output, tt.wantLog) {
				t.Errorf("Done() output = %q, want to contain %q", output, tt.wantLog)
			}
		})
	}
}

func Test_textStepLogger_Fail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		msg     string
		err     error
		wantMsg string
		wantErr string
	}{
		{
			name:    "[正常系] 失敗ログがエラー付きで出力される",
			msg:     "fetching dependencies",
			err:     errors.New("network timeout"),
			wantMsg: "fetching dependencies failed",
			wantErr: "network timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := &bytes.Buffer{}
			l := newTextLogger(slog.LevelDebug, w, CI("")).(*textLogger)
			sl := l.StepLogger(slog.LevelInfo, tt.msg)

			w.Reset()
			sl.Fail(tt.err)

			output := w.String()
			if !strings.Contains(output, tt.wantMsg) {
				t.Errorf("Fail() output = %q, want to contain %q", output, tt.wantMsg)
			}
			if !strings.Contains(output, tt.wantErr) {
				t.Errorf("Fail() output = %q, want to contain error %q", output, tt.wantErr)
			}
		})
	}
}

func Test_textLogger_CommandLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ci        CI
		logLevel  LogLevel
		cmdName   string
		wantType  reflect.Type
		wantLog   string
		wantRegex string // 正規表現でマッチさせる場合（GitLab CI等）
	}{
		{
			name:     "[正常系] 非CI環境ではCommandLoggerが生成され開始ログが出力される",
			ci:       CI(""),
			logLevel: slog.LevelInfo,
			cmdName:  "pnpm install",
			wantType: reflect.TypeFor[*textCommandLogger](),
			wantLog:  "start pnpm install",
		},
		{
			name:     "[正常系] GitHub Actionsでは折りたたみ開始構文が出力される",
			ci:       gitHubActions,
			logLevel: slog.LevelInfo,
			cmdName:  "pnpm install",
			wantType: reflect.TypeFor[*textCommandLogger](),
			wantLog:  "::group::pnpm install",
		},
		{
			name:      "[正常系] GitLab CIでは折りたたみ開始構文が出力される",
			ci:        gitLabCI,
			logLevel:  slog.LevelInfo,
			cmdName:   "pnpm install",
			wantType:  reflect.TypeFor[*textCommandLogger](),
			wantRegex: `\x1b\[0Ksection_start:\d+:pnpm_install\r\x1b\[0Kpnpm install`,
		},
		{
			name:     "[正常系] Azure Pipelinesでは折りたたみ開始構文が出力される",
			ci:       azurePipelines,
			logLevel: slog.LevelInfo,
			cmdName:  "pnpm install",
			wantType: reflect.TypeFor[*textCommandLogger](),
			wantLog:  "##[group]pnpm install",
		},
		{
			name:     "[正常系] TeamCityでは折りたたみ開始構文が出力される",
			ci:       teamCity,
			logLevel: slog.LevelInfo,
			cmdName:  "pnpm install",
			wantType: reflect.TypeFor[*textCommandLogger](),
			wantLog:  "##teamcity[blockOpened name='pnpm install']",
		},
		{
			name:     "[正常系] Buildkiteでは折りたたみ開始構文が出力される",
			ci:       buildkite,
			logLevel: slog.LevelInfo,
			cmdName:  "pnpm install",
			wantType: reflect.TypeFor[*textCommandLogger](),
			wantLog:  "--- pnpm install",
		},
		{
			name:     "[正常系] Travis CIでは折りたたみ開始構文が出力される",
			ci:       travisCI,
			logLevel: slog.LevelInfo,
			cmdName:  "pnpm install",
			wantType: reflect.TypeFor[*textCommandLogger](),
			wantLog:  "travis_fold:start:pnpm_install",
		},
		{
			name:     "[正常系] その他CI環境では折りたたみ構文なしで開始ログが出力される",
			ci:       others,
			logLevel: slog.LevelInfo,
			cmdName:  "pnpm install",
			wantType: reflect.TypeFor[*textCommandLogger](),
			wantLog:  "start pnpm install",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := &bytes.Buffer{}
			l := newTextLogger(slog.LevelDebug, w, tt.ci).(*textLogger)

			got := l.CommandLogger(tt.logLevel, tt.cmdName)

			gotType := reflect.TypeOf(got)
			if gotType != tt.wantType {
				t.Errorf("CommandLogger() returned %s, want %s", gotType, tt.wantType)
			}

			output := w.String()
			if tt.wantRegex != "" {
				if !regexp.MustCompile(tt.wantRegex).MatchString(output) {
					t.Errorf("CommandLogger() output = %q, want to match %q", output, tt.wantRegex)
				}
			} else if !strings.Contains(output, tt.wantLog) {
				t.Errorf("CommandLogger() output = %q, want to contain %q", output, tt.wantLog)
			}
		})
	}
}

func Test_textCommandLogger_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cmdName  string
		input    []byte
		wantN    int
		wantLogs []string
	}{
		{
			name:     "[正常系] バイト列がログに書き込まれ書き込みバイト数が返る",
			cmdName:  "pnpm install",
			input:    []byte("installing packages..."),
			wantN:    len("installing packages..."),
			wantLogs: []string{"installing packages..."},
		},
		{
			name:     "[正常系] 改行を含む入力は行ごとに別々のログエントリとして出力される",
			cmdName:  "pnpm install",
			input:    []byte("line1\nline2\nline3"),
			wantN:    len("line1\nline2\nline3"),
			wantLogs: []string{"line1", "line2", "line3"},
		},
		{
			name:     "[正常系] 末尾の改行による空行は除外される",
			cmdName:  "pnpm install",
			input:    []byte("line1\nline2\n"),
			wantN:    len("line1\nline2\n"),
			wantLogs: []string{"line1", "line2"},
		},
		{
			name:     "[正常系] 先頭の改行による空行は除外される",
			cmdName:  "pnpm install",
			input:    []byte("\nline1\nline2"),
			wantN:    len("\nline1\nline2"),
			wantLogs: []string{"line1", "line2"},
		},
		{
			name:     "[正常系] 先頭と末尾の改行による空行は除外されるが中間の空行は保持される",
			cmdName:  "pnpm install",
			input:    []byte("\nline1\n\nline2\n"),
			wantN:    len("\nline1\n\nline2\n"),
			wantLogs: []string{"line1", "", "line2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := &bytes.Buffer{}
			l := newTextLogger(slog.LevelDebug, w, CI("")).(*textLogger)
			cl := l.CommandLogger(slog.LevelInfo, tt.cmdName)

			w.Reset()
			n, err := cl.Write(tt.input)

			if err != nil {
				t.Fatalf("Write() returned unexpected error: %v", err)
			}
			if n != tt.wantN {
				t.Errorf("Write() = %d, want %d", n, tt.wantN)
			}

			output := w.String()
			// 各ログエントリは slog.TextHandler により "msg=..." を含む1行として出力される
			var logMsgs []string
			for line := range strings.SplitSeq(strings.TrimRight(output, "\n"), "\n") {
				logMsgs = append(logMsgs, line)
			}
			if len(logMsgs) != len(tt.wantLogs) {
				t.Errorf(
					"Write() produced %d log entries, want %d\noutput: %q",
					len(logMsgs),
					len(tt.wantLogs),
					output,
				)
			}
			for _, wantLog := range tt.wantLogs {
				if !strings.Contains(output, wantLog) {
					t.Errorf("Write() output = %q, want to contain %q", output, wantLog)
				}
			}
		})
	}
}

func Test_textCommandLogger_Done(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ci        CI
		cmdName   string
		wantLog   string
		wantRegex string
	}{
		{
			name:      "[正常系] 非CI環境では完了ログが経過時間付きで出力される",
			ci:        CI(""),
			cmdName:   "pnpm install",
			wantRegex: `pnpm install completed in \S+`,
		},
		{
			name:    "[正常系] GitHub Actionsでは完了ログの後に折りたたみ終了構文が出力される",
			ci:      gitHubActions,
			cmdName: "pnpm install",
			wantLog: "::endgroup::",
		},
		{
			name:      "[正常系] GitLab CIでは完了ログの後に折りたたみ終了構文が出力される",
			ci:        gitLabCI,
			cmdName:   "pnpm install",
			wantRegex: `\x1b\[0Ksection_end:\d+:pnpm_install\r\x1b\[0K`,
		},
		{
			name:    "[正常系] Azure Pipelinesでは完了ログの後に折りたたみ終了構文が出力される",
			ci:      azurePipelines,
			cmdName: "pnpm install",
			wantLog: "##[endgroup]",
		},
		{
			name:    "[正常系] TeamCityでは完了ログの後に折りたたみ終了構文が出力される",
			ci:      teamCity,
			cmdName: "pnpm install",
			wantLog: "##teamcity[blockClosed name='pnpm install']",
		},
		{
			name:      "[正常系] Buildkiteでは折りたたみ終了構文なしで完了ログが出力される",
			ci:        buildkite,
			cmdName:   "pnpm install",
			wantRegex: `pnpm install completed in \S+`,
		},
		{
			name:    "[正常系] Travis CIでは完了ログの後に折りたたみ終了構文が出力される",
			ci:      travisCI,
			cmdName: "pnpm install",
			wantLog: "travis_fold:end:pnpm_install",
		},
		{
			name:      "[正常系] その他CI環境では折りたたみ構文なしで完了ログが出力される",
			ci:        others,
			cmdName:   "pnpm install",
			wantRegex: `pnpm install completed in \S+`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := &bytes.Buffer{}
			l := newTextLogger(slog.LevelDebug, w, tt.ci).(*textLogger)
			cl := l.CommandLogger(slog.LevelInfo, tt.cmdName)

			w.Reset()
			cl.Done()

			output := w.String()
			if tt.wantRegex != "" {
				if !regexp.MustCompile(tt.wantRegex).MatchString(output) {
					t.Errorf("Done() output = %q, want to match %q", output, tt.wantRegex)
				}
			} else if !strings.Contains(output, tt.wantLog) {
				t.Errorf("Done() output = %q, want to contain %q", output, tt.wantLog)
			}
		})
	}
}

func Test_textCommandLogger_Fail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ci        CI
		cmdName   string
		exitCode  int
		wantLog   string
		wantRegex string
	}{
		{
			name:      "[正常系] 非CI環境では失敗ログが終了コードと経過時間付きで出力される",
			ci:        CI(""),
			cmdName:   "pnpm install",
			exitCode:  1,
			wantRegex: `pnpm install failed with exit code 1 in \S+`,
		},
		{
			name:     "[正常系] GitHub Actionsでは失敗ログの後に折りたたみ終了構文が出力される",
			ci:       gitHubActions,
			cmdName:  "pnpm install",
			exitCode: 1,
			wantLog:  "::endgroup::",
		},
		{
			name:      "[正常系] GitLab CIでは失敗ログの後に折りたたみ終了構文が出力される",
			ci:        gitLabCI,
			cmdName:   "pnpm install",
			exitCode:  1,
			wantRegex: `\x1b\[0Ksection_end:\d+:pnpm_install\r\x1b\[0K`,
		},
		{
			name:     "[正常系] Azure Pipelinesでは失敗ログの後に折りたたみ終了構文が出力される",
			ci:       azurePipelines,
			cmdName:  "pnpm install",
			exitCode: 1,
			wantLog:  "##[endgroup]",
		},
		{
			name:     "[正常系] TeamCityでは失敗ログの後に折りたたみ終了構文が出力される",
			ci:       teamCity,
			cmdName:  "pnpm install",
			exitCode: 1,
			wantLog:  "##teamcity[blockClosed name='pnpm install']",
		},
		{
			name:      "[正常系] Buildkiteでは折りたたみ終了構文なしで失敗ログが出力される",
			ci:        buildkite,
			cmdName:   "pnpm install",
			exitCode:  1,
			wantRegex: `pnpm install failed with exit code 1 in \S+`,
		},
		{
			name:     "[正常系] Travis CIでは失敗ログの後に折りたたみ終了構文が出力される",
			ci:       travisCI,
			cmdName:  "pnpm install",
			exitCode: 1,
			wantLog:  "travis_fold:end:pnpm_install",
		},
		{
			name:      "[正常系] その他CI環境では折りたたみ構文なしで失敗ログが出力される",
			ci:        others,
			cmdName:   "pnpm install",
			exitCode:  1,
			wantRegex: `pnpm install failed with exit code 1 in \S+`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := &bytes.Buffer{}
			l := newTextLogger(slog.LevelDebug, w, tt.ci).(*textLogger)
			cl := l.CommandLogger(slog.LevelInfo, tt.cmdName)

			w.Reset()
			cl.Fail(tt.exitCode)

			output := w.String()
			if tt.wantRegex != "" {
				if !regexp.MustCompile(tt.wantRegex).MatchString(output) {
					t.Errorf("Fail() output = %q, want to match %q", output, tt.wantRegex)
				}
			} else if !strings.Contains(output, tt.wantLog) {
				t.Errorf("Fail() output = %q, want to contain %q", output, tt.wantLog)
			}
		})
	}
}
