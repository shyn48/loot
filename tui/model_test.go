package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"simple-gui/core"
)

type fakeController struct {
	rows                            []core.JobStatus
	added, paused, resumed, removed string
	opened, pausedAll               bool
}

func (f *fakeController) Add(url string) (string, error) { f.added = url; return "new", nil }
func (f *fakeController) Pause(id string)                { f.paused = id }
func (f *fakeController) Resume(id string)               { f.resumed = id }
func (f *fakeController) Remove(id string)               { f.removed = id }
func (f *fakeController) OpenFolder() error              { f.opened = true; return nil }
func (f *fakeController) Snapshot() []core.JobStatus     { return f.rows }
func (f *fakeController) PauseAll()                      { f.pausedAll = true }

func threeRows() []core.JobStatus {
	return []core.JobStatus{
		{ID: "id-0", Name: "a"},
		{ID: "id-1", Name: "b"},
		{ID: "id-2", Name: "c"},
	}
}

func updateKey(m Model, key string) (Model, tea.Cmd) {
	var km tea.KeyMsg
	switch key {
	case "enter":
		km = tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		km = tea.KeyMsg{Type: tea.KeyEsc}
	default:
		km = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
	tm, cmd := m.Update(km)
	return tm.(Model), cmd
}

func typeString(m Model, s string) Model {
	for _, r := range s {
		tm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = tm.(Model)
	}
	return m
}

func TestCursorNavigation(t *testing.T) {
	fc := &fakeController{rows: threeRows()}
	m := newModel(fc)
	m, _ = updateKey(m, "j")
	m, _ = updateKey(m, "j")
	if m.cursor != 2 {
		t.Fatalf("cursor=%d", m.cursor)
	}
	m, _ = updateKey(m, "j") // clamp at last
	if m.cursor != 2 {
		t.Fatalf("cursor overran: %d", m.cursor)
	}
	m, _ = updateKey(m, "k")
	if m.cursor != 1 {
		t.Fatalf("cursor up: %d", m.cursor)
	}
}

func TestPauseKeyCallsController(t *testing.T) {
	fc := &fakeController{rows: threeRows()}
	m := newModel(fc)
	m.cursor = 1
	updateKey(m, "p")
	if fc.paused != "id-1" {
		t.Fatalf("paused=%q", fc.paused)
	}
}

func TestResumeAndRemoveKeys(t *testing.T) {
	fc := &fakeController{rows: threeRows()}
	m := newModel(fc)
	m.cursor = 2
	updateKey(m, "r")
	updateKey(m, "d")
	if fc.resumed != "id-2" || fc.removed != "id-2" {
		t.Fatalf("resumed=%q removed=%q", fc.resumed, fc.removed)
	}
}

func TestAddModeFlow(t *testing.T) {
	fc := &fakeController{}
	m := newModel(fc)
	m, _ = updateKey(m, "a")
	if !m.adding {
		t.Fatal("expected add mode")
	}
	m = typeString(m, "http://x/f")
	m, _ = updateKey(m, "enter")
	if fc.added != "http://x/f" || m.adding {
		t.Fatalf("added=%q adding=%v", fc.added, m.adding)
	}
}

func TestQuitPausesAll(t *testing.T) {
	fc := &fakeController{rows: threeRows()}
	m := newModel(fc)
	updateKey(m, "q")
	if !fc.pausedAll {
		t.Fatal("quit should PauseAll")
	}
}

func TestOpenFolderKey(t *testing.T) {
	fc := &fakeController{rows: threeRows()}
	m := newModel(fc)
	updateKey(m, "o")
	if !fc.opened {
		t.Fatal("o should open folder")
	}
}
