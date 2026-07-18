package core

import (
	"path/filepath"
	"testing"
)

func TestInboxAppendAndDrain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inbox")

	if err := appendToInboxFile(path, []string{"http://x/a", "  ", "http://x/b"}); err != nil {
		t.Fatal(err)
	}
	if err := appendToInboxFile(path, []string{"http://x/c"}); err != nil {
		t.Fatal(err)
	}

	got := drainInboxFile(path)
	want := []string{"http://x/a", "http://x/b", "http://x/c"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d = %q, want %q", i, got[i], want[i])
		}
	}

	// Inbox is cleared after draining.
	if again := drainInboxFile(path); len(again) != 0 {
		t.Fatalf("inbox not cleared: %v", again)
	}
}
