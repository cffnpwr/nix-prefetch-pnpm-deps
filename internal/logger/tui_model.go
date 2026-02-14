package logger

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const ringBufferSize = 10

// TUI message types.
type (
	logLineMsg   struct{ line string }
	stepStartMsg struct{ msg string }
	stepDoneMsg  struct{}
	stepFailMsg  struct{ err error }
	cmdStartMsg  struct{ name string }
	cmdLineMsg   struct{ line string }
	cmdDoneMsg   struct{ elapsed time.Duration }
	cmdFailMsg   struct {
		exitCode int
		elapsed  time.Duration
	}
	interruptMsg struct{}
)

const cmdBoxMarginLeft = 2

var (
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

	doneStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	warnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	failStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	cmdBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			PaddingLeft(1).
			MarginLeft(cmdBoxMarginLeft)

	overflowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// tuiModel is the Bubble Tea model.
type tuiModel struct {
	spinner  spinner.Model
	lines    []string // completed log lines
	active   *activeEntry
	quitting bool
}

type activeEntry struct {
	msg        string
	isCommand  bool
	ringBuf    []string
	totalLines int
}

func newTUIModel() tuiModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle
	return tuiModel{
		spinner: s,
	}
}

func (m tuiModel) Init() tea.Cmd {
	return m.spinner.Tick
}

//nolint:cyclop,funlen // switch on message types is inherently branchy
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.QuitMsg:
		m.quitting = true
		return m, tea.Quit

	case interruptMsg:
		m.active = nil
		m.lines = append(m.lines, warnStyle.Render("!")+" interrupted")
		m.quitting = true
		return m, tea.Quit

	case logLineMsg:
		m.lines = append(m.lines, msg.line)
		return m, nil

	case stepStartMsg:
		m.active = &activeEntry{msg: msg.msg}
		return m, nil

	case stepDoneMsg:
		if m.active != nil {
			m.lines = append(m.lines, doneStyle.Render("✓")+" "+m.active.msg)
			m.active = nil
		}
		return m, nil

	case stepFailMsg:
		if m.active != nil {
			m.lines = append(m.lines,
				failStyle.Render("✖")+" "+m.active.msg+": "+msg.err.Error(),
			)
			m.active = nil
		}
		return m, nil

	case cmdStartMsg:
		m.active = &activeEntry{
			msg:       msg.name,
			isCommand: true,
			ringBuf:   make([]string, 0, ringBufferSize),
		}
		return m, nil

	case cmdLineMsg:
		if m.active != nil && m.active.isCommand {
			m.active.totalLines++
			m.active.ringBuf = appendRing(m.active.ringBuf, msg.line)
		}
		return m, nil

	case cmdDoneMsg:
		if m.active != nil {
			m.lines = append(m.lines,
				doneStyle.Render("✓")+
					fmt.Sprintf(
						" `%s` exit successfully in %s",
						m.active.msg, msg.elapsed,
					),
			)
			m.active = nil
		}
		return m, nil

	case cmdFailMsg:
		if m.active != nil {
			header := failStyle.Render("✖") +
				fmt.Sprintf(
					" `%s` failed with exit code %d in %s",
					m.active.msg, msg.exitCode, msg.elapsed,
				)
			if len(m.active.ringBuf) > 0 {
				cmdOutput := strings.Join(m.active.ringBuf, "\n")
				header += "\n" + cmdBoxStyle.Render(cmdOutput)
			}
			m.lines = append(m.lines, header)
			m.active = nil
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	default:
		return m, nil
	}
}

func (m tuiModel) View() string {
	if m.quitting {
		return strings.Join(m.lines, "\n") + "\n"
	}

	var b strings.Builder
	for _, l := range m.lines {
		b.WriteString(l)
		b.WriteByte('\n')
	}

	if m.active == nil {
		return b.String()
	}

	b.WriteString(m.spinner.View())
	b.WriteByte(' ')
	if m.active.isCommand {
		fmt.Fprintf(&b, "Executing `%s` ...", m.active.msg)
		b.WriteByte('\n')
		cmdLines := append([]string{}, m.active.ringBuf...)
		overflow := m.active.totalLines - len(m.active.ringBuf)
		if overflow > 0 {
			cmdLines = append(
				cmdLines,
				overflowStyle.Render(fmt.Sprintf("(+%d lines)", overflow)),
			)
		}
		if len(cmdLines) > 0 {
			b.WriteString(cmdBoxStyle.Render(strings.Join(cmdLines, "\n")))
			b.WriteByte('\n')
		}
	} else {
		b.WriteString(m.active.msg)
		b.WriteByte('\n')
	}
	return b.String()
}

func appendRing(buf []string, line string) []string {
	if len(buf) >= ringBufferSize {
		copy(buf, buf[1:])
		buf[len(buf)-1] = line
		return buf
	}
	return append(buf, line)
}
