package core

import (
	"path/filepath"
	"testing"
)

func TestMetaRoundTrip(t *testing.T) {
	dir := t.TempDir()
	m := meta{ID: "abc", URL: "http://x/f", Filename: "f", TargetPath: "/d/f",
		Size: 100, TotalSection: 4, Sections: [][2]int64{{0, 24}, {25, 49}, {50, 74}, {75, 99}}, AcceptRanges: true}
	if err := writeMeta(dir, m); err != nil {
		t.Fatal(err)
	}
	got, err := readMeta(filepath.Join(dir, "abc.meta.json"))
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != m.ID || got.Size != m.Size || len(got.Sections) != 4 || !got.AcceptRanges {
		t.Fatalf("round trip mismatch: %+v", got)
	}
}

func TestListMetaFiles(t *testing.T) {
	dir := t.TempDir()
	if err := writeMeta(dir, meta{ID: "a"}); err != nil {
		t.Fatal(err)
	}
	if err := writeMeta(dir, meta{ID: "b"}); err != nil {
		t.Fatal(err)
	}
	files, err := listMetaFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("want 2 meta files, got %d: %v", len(files), files)
	}
}
