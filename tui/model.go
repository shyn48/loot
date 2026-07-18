package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"simple-gui/core"
)

// Model is the Bubble Tea model. It holds no download state of its own — it
// renders core.Controller snapshots and forwards key actions to the controller.
type Model struct {
	ctrl     core.Controller
	rows     []core.JobStatus
	cursor   int
	adding   bool
	input    textinput.Model
	showHelp bool
	errMsg   string
	clockStr string
	w, h     int
}

type tickMsg time.Time

func newModel(ctrl core.Controller) Model {
	ti := textinput.New()
	ti.Placeholder = "https://example.com/file.zip"
	ti.CharLimit = 2048
	ti.Width = 60
	return Model{ctrl: ctrl, rows: sortRows(ctrl.Snapshot()), input: ti}
}

func (m Model) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		return m, nil
	case tickMsg:
		m.rows = sortRows(m.ctrl.Snapshot())
		m.clockStr = time.Time(msg).Format("15:04:05")
		m.clampCursor()
		return m, tick()
	case tea.KeyMsg:
		if m.adding {
			return m.updateAdding(msg)
		}
		return m.updateNormal(msg)
	}
	return m, nil
}

func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.ctrl.PauseAll()
		return m, tea.Quit
	case "j", "down":
		if m.cursor < len(m.rows)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "a":
		m.adding = true
		m.errMsg = ""
		m.input.Reset()
		m.input.Focus()
		return m, textinput.Blink
	case "p":
		if id, ok := m.selectedID(); ok {
			m.ctrl.Pause(id)
		}
	case "r":
		if id, ok := m.selectedID(); ok {
			m.ctrl.Resume(id)
		}
	case "d":
		if id, ok := m.selectedID(); ok {
			m.ctrl.Remove(id)
		}
	case "o":
		m.ctrl.OpenFolder()
	case "?":
		m.showHelp = !m.showHelp
	}
	return m, nil
}

func (m Model) updateAdding(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if url := strings.TrimSpace(m.input.Value()); url != "" {
			if _, err := m.ctrl.Add(url); err != nil {
				m.errMsg = err.Error()
			} else {
				m.errMsg = ""
			}
		}
		m.input.Reset()
		m.adding = false
		return m, nil
	case "esc":
		m.input.Reset()
		m.adding = false
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) selectedID() (string, bool) {
	if m.cursor >= 0 && m.cursor < len(m.rows) {
		return m.rows[m.cursor].ID, true
	}
	return "", false
}

func (m Model) selected() (core.JobStatus, bool) {
	if m.cursor >= 0 && m.cursor < len(m.rows) {
		return m.rows[m.cursor], true
	}
	return core.JobStatus{}, false
}

func (m *Model) clampCursor() {
	if m.cursor >= len(m.rows) {
		m.cursor = len(m.rows) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}
