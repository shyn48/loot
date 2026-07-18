package core

import (
	"testing"
	"time"
)

func TestClearCompleted(t *testing.T) {
	body := makeBody(8192)
	srv := rangeServer(body)
	defer srv.Close()

	m := newTestManager(t)
	id, err := m.Add(srv.URL + "/f.bin")
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { return stateOf(m, id) == StateDone }, 10*time.Second)

	if !fileExists(metaPath(m.stateDir, id)) {
		t.Fatal("meta should exist before clear")
	}

	m.ClearCompleted()

	if _, ok := statusOf(m, id); ok {
		t.Fatal("completed job still present after clear")
	}
	if fileExists(metaPath(m.stateDir, id)) {
		t.Fatal("meta not deleted by clear")
	}
}
