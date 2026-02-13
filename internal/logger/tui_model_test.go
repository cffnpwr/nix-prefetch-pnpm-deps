package logger

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-cmp/cmp"
)

// updateModel is a test helper that calls Update and type-asserts the result.
func updateModel(m tuiModel, msg tea.Msg) tuiModel {
	updated, _ := m.Update(msg)
	return updated.(tuiModel)
}

func Test_appendRing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		buf  []string
		line string
		want []string
	}{
		{
			name: "[正常系] バッファが上限未満なら追加される",
			buf:  []string{"a", "b"},
			line: "c",
			want: []string{"a", "b", "c"},
		},
		{
			name: "[正常系] バッファが上限に達すると先頭が押し出される",
			buf:  func() []string { return []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"} }(),
			line: "new",
			want: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "new"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := appendRing(tt.buf, tt.line)
			if d := cmp.Diff(tt.want, got); d != "" {
				t.Errorf("appendRing() mismatch (-want +got):\n%s", d)
			}
		})
	}
}

//nolint:cyclop // table-driven test with per-case assert functions
func Test_tuiModel_Update_setActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		setup  func() tuiModel
		msg    tea.Msg
		assert func(t *testing.T, m tuiModel)
	}{
		{
			name:  "[正常系] stepStartMsgでactiveEntryが設定される",
			setup: func() tuiModel { return newTUIModel() },
			msg:   stepStartMsg{msg: "loading lockfile"},
			assert: func(t *testing.T, m tuiModel) {
				t.Helper()
				if m.active == nil {
					t.Fatal("active is nil")
				}
				if m.active.msg != "loading lockfile" {
					t.Errorf("active.msg = %q, want %q", m.active.msg, "loading lockfile")
				}
				if m.active.isCommand {
					t.Error("active.isCommand should be false")
				}
			},
		},
		{
			name:  "[正常系] cmdStartMsgでコマンド用のactiveEntryが設定される",
			setup: func() tuiModel { return newTUIModel() },
			msg:   cmdStartMsg{name: "pnpm install"},
			assert: func(t *testing.T, m tuiModel) {
				t.Helper()
				if m.active == nil {
					t.Fatal("active is nil")
				}
				if m.active.msg != "pnpm install" {
					t.Errorf("active.msg = %q, want %q", m.active.msg, "pnpm install")
				}
				if !m.active.isCommand {
					t.Error("active.isCommand should be true")
				}
				if len(m.active.ringBuf) != 0 {
					t.Errorf("ringBuf length = %d, want 0", len(m.active.ringBuf))
				}
			},
		},
		{
			name: "[正常系] cmdLineMsgでリングバッファに行が追加される",
			setup: func() tuiModel {
				m := newTUIModel()
				m = updateModel(m, cmdStartMsg{name: "test"})
				return m
			},
			msg: cmdLineMsg{line: "output line"},
			assert: func(t *testing.T, m tuiModel) {
				t.Helper()
				want := []string{"output line"}
				if d := cmp.Diff(want, m.active.ringBuf); d != "" {
					t.Errorf("ringBuf mismatch (-want +got):\n%s", d)
				}
				if m.active.totalLines != 1 {
					t.Errorf("totalLines = %d, want 1", m.active.totalLines)
				}
			},
		},
		{
			name: "[正常系] cmdLineMsgでリングバッファが上限を超えると古い行が消える",
			setup: func() tuiModel {
				m := newTUIModel()
				m = updateModel(m, cmdStartMsg{name: "test"})
				for i := range 10 {
					m = updateModel(m, cmdLineMsg{line: fmt.Sprintf("old%d", i)})
				}
				return m
			},
			msg: cmdLineMsg{line: "new"},
			assert: func(t *testing.T, m tuiModel) {
				t.Helper()
				if m.active.totalLines != 11 {
					t.Errorf("totalLines = %d, want 11", m.active.totalLines)
				}
				if m.active.ringBuf[0] != "old1" {
					t.Errorf("ringBuf[0] = %q, want %q", m.active.ringBuf[0], "old1")
				}
				if m.active.ringBuf[9] != "new" {
					t.Errorf("ringBuf[9] = %q, want %q", m.active.ringBuf[9], "new")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := tt.setup()
			result := updateModel(m, tt.msg)
			tt.assert(t, result)
		})
	}
}

func Test_tuiModel_Update_completeLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setup        func() tuiModel
		msg          tea.Msg
		wantContains []string
	}{
		{
			name:         "[正常系] logLineMsgでログ行がlinesに追加される",
			setup:        func() tuiModel { return newTUIModel() },
			msg:          logLineMsg{line: "hello world"},
			wantContains: []string{"hello world"},
		},
		{
			name: "[正常系] stepDoneMsgでactiveがnilになりlinesに完了行が追加される",
			setup: func() tuiModel {
				m := newTUIModel()
				m = updateModel(m, stepStartMsg{msg: "loading lockfile"})
				return m
			},
			msg:          stepDoneMsg{},
			wantContains: []string{"loading lockfile"},
		},
		{
			name: "[正常系] stepFailMsgでactiveがnilになりlinesにエラー行が追加される",
			setup: func() tuiModel {
				m := newTUIModel()
				m = updateModel(m, stepStartMsg{msg: "loading lockfile"})
				return m
			},
			msg:          stepFailMsg{err: errors.New("file not found")},
			wantContains: []string{"loading lockfile", "file not found"},
		},
		{
			name: "[正常系] cmdDoneMsgでactiveがnilになりlinesに完了行が追加される",
			setup: func() tuiModel {
				m := newTUIModel()
				m = updateModel(m, cmdStartMsg{name: "pnpm install"})
				return m
			},
			msg:          cmdDoneMsg{elapsed: 3 * time.Second},
			wantContains: []string{"pnpm install", "exit successfully"},
		},
		{
			name: "[正常系] cmdFailMsgでactiveがnilになりlinesにエラー行とバッファ内容が追加される",
			setup: func() tuiModel {
				m := newTUIModel()
				m = updateModel(m, cmdStartMsg{name: "pnpm install"})
				m = updateModel(m, cmdLineMsg{line: "resolving..."})
				m = updateModel(m, cmdLineMsg{line: "ERROR: not found"})
				return m
			},
			msg: cmdFailMsg{exitCode: 1, elapsed: 5 * time.Second},
			wantContains: []string{
				"pnpm install", "failed with exit code",
				"resolving...", "ERROR: not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := tt.setup()
			result := updateModel(m, tt.msg)

			if result.active != nil {
				t.Error("active should be nil")
			}
			if len(result.lines) != 1 {
				t.Fatalf("lines length = %d, want 1", len(result.lines))
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(result.lines[0], want) {
					t.Errorf("lines[0] = %q, want to contain %q", result.lines[0], want)
				}
			}
		})
	}
}

func Test_tuiModel_Update_ignored(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  tea.Msg
	}{
		{
			name: "[正常系] activeがnilのときstepDoneMsgは無視される",
			msg:  stepDoneMsg{},
		},
		{
			name: "[正常系] activeがnilのときstepFailMsgは無視される",
			msg:  stepFailMsg{err: errors.New("err")},
		},
		{
			name: "[正常系] activeがnilのときcmdLineMsgは無視される",
			msg:  cmdLineMsg{line: "output"},
		},
		{
			name: "[正常系] activeがnilのときcmdDoneMsgは無視される",
			msg:  cmdDoneMsg{elapsed: time.Second},
		},
		{
			name: "[正常系] activeがnilのときcmdFailMsgは無視される",
			msg:  cmdFailMsg{exitCode: 1, elapsed: time.Second},
		},
		{
			name: "[正常系] 未知のメッセージは状態を変更しない",
			msg:  struct{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newTUIModel()
			result := updateModel(m, tt.msg)

			if result.active != nil {
				t.Error("active should remain nil")
			}
			if len(result.lines) != 0 {
				t.Errorf("lines should be empty, got %v", result.lines)
			}
		})
	}
}

func Test_tuiModel_Update_QuitMsg(t *testing.T) {
	t.Parallel()

	m := newTUIModel()
	updated, cmd := m.Update(tea.QuitMsg{})
	result := updated.(tuiModel)

	if !result.quitting {
		t.Error("quitting should be true")
	}
	if cmd == nil {
		t.Error("cmd should not be nil")
	}
}

func Test_tuiModel_Update_interruptMsg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func() tuiModel
		wantLines  int
		wantActive bool
	}{
		{
			name:       "[正常系] activeがnilのときinterruptMsgでquittingが設定される",
			setup:      func() tuiModel { return newTUIModel() },
			wantLines:  1,
			wantActive: false,
		},
		{
			name: "[正常系] step実行中にinterruptMsgでactiveがクリアされる",
			setup: func() tuiModel {
				m := newTUIModel()
				m = updateModel(m, stepStartMsg{msg: "loading lockfile"})
				return m
			},
			wantLines:  1,
			wantActive: false,
		},
		{
			name: "[正常系] command実行中にinterruptMsgでactiveがクリアされる",
			setup: func() tuiModel {
				m := newTUIModel()
				m = updateModel(m, cmdStartMsg{name: "pnpm install"})
				m = updateModel(m, cmdLineMsg{line: "resolving..."})
				return m
			},
			wantLines:  1,
			wantActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := tt.setup()
			updated, cmd := m.Update(interruptMsg{})
			result := updated.(tuiModel)

			if !result.quitting {
				t.Error("quitting should be true")
			}
			if cmd == nil {
				t.Error("cmd should not be nil (tea.Quit)")
			}
			if (result.active != nil) != tt.wantActive {
				t.Errorf("active != nil is %v, want %v", result.active != nil, tt.wantActive)
			}
			if len(result.lines) != tt.wantLines {
				t.Fatalf("lines length = %d, want %d", len(result.lines), tt.wantLines)
			}
			if !strings.Contains(result.lines[len(result.lines)-1], "interrupted") {
				t.Errorf(
					"last line = %q, want to contain %q",
					result.lines[len(result.lines)-1],
					"interrupted",
				)
			}
		})
	}
}

func Test_tuiModel_View(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setup        func() tuiModel
		wantContains []string
		wantAbsent   []string
	}{
		{
			name: "[正常系] quitting時は完了行のみ出力される",
			setup: func() tuiModel {
				m := newTUIModel()
				m = updateModel(m, logLineMsg{line: "line1"})
				m = updateModel(m, logLineMsg{line: "line2"})
				m = updateModel(m, tea.QuitMsg{})
				return m
			},
			wantContains: []string{"line1", "line2"},
		},
		{
			name: "[正常系] activeがnilのときはlinesのみ出力される",
			setup: func() tuiModel {
				m := newTUIModel()
				m = updateModel(m, logLineMsg{line: "done line"})
				return m
			},
			wantContains: []string{"done line"},
			wantAbsent:   []string{"Executing"},
		},
		{
			name: "[正常系] step実行中はメッセージが表示される",
			setup: func() tuiModel {
				m := newTUIModel()
				m = updateModel(m, stepStartMsg{msg: "loading lockfile"})
				return m
			},
			wantContains: []string{"loading lockfile"},
			wantAbsent:   []string{"Executing"},
		},
		{
			name: "[正常系] command実行中はExecutingとバッファが表示される",
			setup: func() tuiModel {
				m := newTUIModel()
				m = updateModel(m, cmdStartMsg{name: "pnpm install"})
				m = updateModel(m, cmdLineMsg{line: "resolving..."})
				m = updateModel(m, cmdLineMsg{line: "fetching..."})
				return m
			},
			wantContains: []string{
				"Executing `pnpm install`",
				"resolving...",
				"fetching...",
			},
		},
		{
			name: "[正常系] command実行中にoverflowがある場合は件数が表示される",
			setup: func() tuiModel {
				m := newTUIModel()
				m = updateModel(m, cmdStartMsg{name: "test"})
				for i := range 15 {
					m = updateModel(m, cmdLineMsg{
						line: fmt.Sprintf("line%d", i),
					})
				}
				return m
			},
			wantContains: []string{
				"Executing `test`",
				"+5 lines",
			},
			wantAbsent: []string{
				"| line0", "| line1\n", "| line2\n", "| line3\n", "| line4\n",
			},
		},
		{
			name: "[正常系] interruptMsg後はinterrupted行のみ表示されスピナーは表示されない",
			setup: func() tuiModel {
				m := newTUIModel()
				m = updateModel(m, logLineMsg{line: "line1"})
				m = updateModel(m, cmdStartMsg{name: "pnpm install"})
				m = updateModel(m, cmdLineMsg{line: "resolving..."})
				m = updateModel(m, interruptMsg{})
				return m
			},
			wantContains: []string{"line1", "interrupted"},
			wantAbsent:   []string{"Executing", "pnpm install", "resolving..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := tt.setup()
			view := m.View()

			for _, want := range tt.wantContains {
				if !strings.Contains(view, want) {
					t.Errorf(
						"View() = %q, want to contain %q",
						view, want,
					)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(view, absent) {
					t.Errorf(
						"View() = %q, should not contain %q",
						view, absent,
					)
				}
			}
		})
	}
}
