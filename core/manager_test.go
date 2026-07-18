package core

import (
	"testing"
	"time"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	m, err := newManager(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m.bytesPerSection = 4096 // segment even the small bodies used in tests
	t.Cleanup(func() { m.Close() })
	return m
}

func stateOf(m *Manager, id string) JobState {
	for _, st := range m.Snapshot() {
		if st.ID == id {
			return st.State
		}
	}
	return ""
}

func waitFor(t *testing.T, cond func() bool, d time.Duration) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("condition not met in time")
}

func TestManagerAddCompletes(t *testing.T) {
	body := makeBody(8192)
	srv := rangeServer(body)
	defer srv.Close()

	m := newTestManager(t)
	id, err := m.Add(srv.URL + "/f.bin")
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { return stateOf(m, id) == StateDone }, 5*time.Second)
}
