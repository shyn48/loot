package tui

import (
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"simple-gui/core"
	"simple-gui/helper"
)

// readClipboard is a package var so tests can stub out the real system clipboard.
var readClipboard = clipboard.ReadAll

// splitURLs splits a batch-add input into individual URLs on any whitespace.
func splitURLs(s string) []string {
	return strings.Fields(s)
}

// Model is the Bubble Tea model. It holds no download state of its own — it
// renders core.Controller snapshots and forwards key actions to the controller.
type Model struct {
	ctrl      core.Controller
	rows      []core.JobStatus
	cursor    int
	adding    bool
	input     textinput.Model
	showHelp  bool
	filtering bool
	filter    string
	errMsg    string
	clockStr  string
	speedHist map[string][]float64 // per-job recent speed samples, for the sparkline
	w, h      int
}

const speedHistLen = 120

type tickMsg time.Time

func newModel(ctrl core.Controller) Model {
	ti := textinput.New()
	ti.Placeholder = "https://example.com/file.zip"
	ti.CharLimit = 2048
	ti.Width = 60
	return Model{ctrl: ctrl, rows: sortRows(ctrl.Snapshot()), input: ti, speedHist: map[string][]float64{}}
}

// recordSpeeds appends the current per-job speed to each job's bounded history
// and drops history for jobs that no longer exist.
func (m *Model) recordSpeeds() {
	seen := make(map[string]bool, len(m.rows))
	for _, r := range m.rows {
		seen[r.ID] = true
		h := append(m.speedHist[r.ID], r.Speed)
		if len(h) > speedHistLen {
			h = h[len(h)-speedHistLen:]
		}
		m.speedHist[r.ID] = h
	}
	for id := range m.speedHist {
		if !seen[id] {
			delete(m.speedHist, id)
		}
	}
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
		m.recordSpeeds()
		m.clampCursor()
		return m, tick()
	case tea.KeyMsg:
		if m.adding {
			return m.updateAdding(msg)
		}
		if m.filtering {
			return m.updateFiltering(msg)
		}
		return m.updateNormal(msg)
	}
	return m, nil
}

// visible returns the rows currently shown: all of them, or those whose name
// matches the active filter (case-insensitive substring).
func (m Model) visible() []core.JobStatus {
	if m.filter == "" {
		return m.rows
	}
	q := strings.ToLower(m.filter)
	out := make([]core.JobStatus, 0, len(m.rows))
	for _, r := range m.rows {
		if strings.Contains(strings.ToLower(r.Name), q) {
			out = append(out, r)
		}
	}
	return out
}

func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.ctrl.PauseAll()
		return m, tea.Quit
	case "j", "down":
		if m.cursor < len(m.visible())-1 {
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
		// Prefill from the clipboard if it looks like a URL (copy link → a → enter).
		if clip, err := readClipboard(); err == nil {
			clip = strings.TrimSpace(clip)
			if f := strings.Fields(clip); len(f) > 0 && helper.IsValidUrl(f[0]) {
				m.input.SetValue(clip)
				m.input.CursorEnd()
			}
		}
		m.input.Focus()
		return m, textinput.Blink
	case "/":
		m.filtering = true
		m.filter = ""
		m.input.Reset()
		m.input.Focus()
		return m, textinput.Blink
	case "esc":
		m.filter = "" // clear an applied filter
		m.clampCursor()
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
		var lastErr string
		added := 0
		for _, url := range splitURLs(m.input.Value()) {
			if _, err := m.ctrl.Add(url); err != nil {
				lastErr = err.Error()
			} else {
				added++
			}
		}
		if added == 0 && lastErr != "" {
			m.errMsg = lastErr
		} else {
			m.errMsg = ""
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

func (m Model) updateFiltering(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.filtering = false // keep the filter applied
		return m, nil
	case "esc":
		m.filtering = false
		m.filter = ""
		m.input.Reset()
		m.clampCursor()
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.filter = m.input.Value()
	m.clampCursor()
	return m, cmd
}

func (m Model) selectedID() (string, bool) {
	rows := m.visible()
	if m.cursor >= 0 && m.cursor < len(rows) {
		return rows[m.cursor].ID, true
	}
	return "", false
}

func (m Model) selected() (core.JobStatus, bool) {
	rows := m.visible()
	if m.cursor >= 0 && m.cursor < len(rows) {
		return rows[m.cursor], true
	}
	return core.JobStatus{}, false
}

func (m *Model) clampCursor() {
	n := len(m.visible())
	if m.cursor >= n {
		m.cursor = n - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}
