package core

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLoadPersistedResumes simulates an app restart: a first manager downloads
// partway and pauses, then a fresh manager over the same dirs reloads the job
// and resumes it to completion.
func TestLoadPersistedResumes(t *testing.T) {
	body := makeBody(100000)
	rs := newRecordingSlowServer(body, 512, 15*time.Millisecond)
	defer rs.Close()

	dl, st := t.TempDir(), t.TempDir()

	m1, err := newManager(dl, st)
	if err != nil {
		t.Fatal(err)
	}
	id, err := m1.Add(rs.URL + "/f.bin")
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		s, _ := statusOf(m1, id)
		return s.Downloaded > 0 && s.Downloaded < s.Size
	}, 3*time.Second)
	m1.Pause(id)
	waitFor(t, func() bool { return stateOf(m1, id) == StatePaused }, 3*time.Second)
	m1.Close() // simulate process exit

	// Fresh manager over the same directories.
	m2, err := newManager(dl, st)
	if err != nil {
		t.Fatal(err)
	}
	defer m2.Close()
	if err := m2.LoadPersisted(); err != nil {
		t.Fatal(err)
	}

	st2, ok := statusOf(m2, id)
	if !ok {
		t.Fatal("persisted job not reloaded")
	}
	if st2.State != StatePaused {
		t.Fatalf("reloaded state = %s, want paused", st2.State)
	}
	if st2.Downloaded == 0 {
		t.Fatal("reloaded job shows no progress")
	}

	m2.Resume(id)
	waitFor(t, func() bool { return stateOf(m2, id) == StateDone }, 10*time.Second)

	got, _ := os.ReadFile(filepath.Join(dl, st2.Name))
	if sha256.Sum256(got) != sha256.Sum256(body) {
		t.Fatal("restart-resumed content mismatch")
	}
}

func TestRemoveDeletesArtifacts(t *testing.T) {
	body := makeBody(50000)
	rs := slowRangeServer(body, 512, 10*time.Millisecond)
	defer rs.Close()

	m := newTestManager(t)
	id, err := m.Add(rs.URL + "/f.bin")
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { s, _ := statusOf(m, id); return s.Downloaded > 0 }, 3*time.Second)
	m.Pause(id)
	waitFor(t, func() bool { return stateOf(m, id) == StatePaused }, 3*time.Second)

	if !fileExists(metaPath(m.stateDir, id)) {
		t.Fatal("meta should exist before remove")
	}

	m.Remove(id)

	if _, ok := statusOf(m, id); ok {
		t.Fatal("job still present after remove")
	}
	if fileExists(metaPath(m.stateDir, id)) {
		t.Fatal("meta not deleted")
	}
	temps, _ := filepath.Glob(filepath.Join(m.stateDir, "section-*"))
	if len(temps) != 0 {
		t.Fatalf("temp files remain: %v", temps)
	}
}
