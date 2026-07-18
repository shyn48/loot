package core

import (
	"os"
	"testing"
	"time"
)

func statusOf(m *Manager, id string) (JobStatus, bool) {
	for _, st := range m.Snapshot() {
		if st.ID == id {
			return st, true
		}
	}
	return JobStatus{}, false
}

func TestSpeedSampling(t *testing.T) {
	m := newTestManager(t)
	j := &Job{ID: "s", Filename: "f", Size: 1000, AcceptRanges: true,
		TotalSection: 2, Sections: computeSections(1000, 2), state: StateDownloading}
	m.mu.Lock()
	m.jobs = append(m.jobs, j)
	m.mu.Unlock()

	// Baseline: 0 bytes one second ago.
	j.mu.Lock()
	j.lastBytes = 0
	j.lastSampleTime = time.Now().Add(-1 * time.Second)
	j.mu.Unlock()

	// 500 bytes now present → ~500 B/s.
	if err := os.WriteFile(j.sectionFile(m.stateDir, 0), make([]byte, 500), 0o644); err != nil {
		t.Fatal(err)
	}

	m.sample()

	st, _ := statusOf(m, "s")
	if st.Speed <= 0 {
		t.Fatalf("expected positive speed, got %f", st.Speed)
	}
	if st.ETASeconds <= 0 {
		t.Fatalf("expected positive ETA, got %d", st.ETASeconds)
	}
}
