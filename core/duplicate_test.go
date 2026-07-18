package core

import (
	"crypto/sha256"
	"os"
	"testing"
	"time"
)

// Adding the same URL twice must produce two independent downloads: distinct
// ids, distinct target files, distinct temp files, and both files intact.
func TestAddDuplicateURL(t *testing.T) {
	body := makeBody(20000)
	srv := rangeServer(body)
	defer srv.Close()

	m := newTestManager(t)
	id1, err := m.Add(srv.URL + "/dup.bin")
	if err != nil {
		t.Fatal(err)
	}
	id2, err := m.Add(srv.URL + "/dup.bin")
	if err != nil {
		t.Fatal(err)
	}
	if id1 == id2 {
		t.Fatal("duplicate add returned the same id")
	}

	waitFor(t, func() bool {
		return stateOf(m, id1) == StateDone && stateOf(m, id2) == StateDone
	}, 10*time.Second)

	j1, j2 := m.findJob(id1), m.findJob(id2)
	if j1.TargetPath == j2.TargetPath {
		t.Fatalf("both downloads share a target path: %s", j1.TargetPath)
	}
	for _, j := range []*Job{j1, j2} {
		got, err := os.ReadFile(j.TargetPath)
		if err != nil {
			t.Fatalf("read %s: %v", j.TargetPath, err)
		}
		if sha256.Sum256(got) != sha256.Sum256(body) {
			t.Fatalf("file %s is corrupt (len %d, want %d)", j.TargetPath, len(got), len(body))
		}
	}
}
