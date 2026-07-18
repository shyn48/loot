package core

import "testing"

func TestComputeSectionsCoversRange(t *testing.T) {
	cases := []struct {
		size int64
		n    int
	}{{100, 20}, {101, 20}, {19, 20}, {1000, 3}}
	for _, c := range cases {
		secs := computeSections(c.size, c.n)
		if secs[0][0] != 0 || secs[len(secs)-1][1] != c.size-1 {
			t.Fatalf("size=%d n=%d bad ends: %v", c.size, c.n, secs)
		}
		for i := 1; i < len(secs); i++ {
			if secs[i][0] != secs[i-1][1]+1 {
				t.Fatalf("gap/overlap at %d: %v", i, secs)
			}
		}
	}
}

func TestRemainingRange(t *testing.T) {
	s, e, ok := remainingRange([2]int64{10, 19}, 4) // have 4 of 10
	if !ok || s != 14 || e != 19 {
		t.Fatalf("got %d-%d ok=%v", s, e, ok)
	}
	if _, _, ok := remainingRange([2]int64{10, 19}, 10); ok {
		t.Fatal("complete section should report ok=false")
	}
}

func TestSectionsForSize(t *testing.T) {
	const per = 2 * 1024 * 1024
	cases := map[int64]int{
		0:                 1,
		500 * 1024:        1,  // <2MB → single
		2 * 1024 * 1024:   1,  // exactly 2MB → 1
		10 * 1024 * 1024:  5,  // ~5
		100 * 1024 * 1024: 20, // capped
		4_000_000_000:     20, // capped
	}
	for size, want := range cases {
		if got := sectionsForSize(size, per); got != want {
			t.Errorf("sectionsForSize(%d) = %d, want %d", size, got, want)
		}
	}
}

func TestETA(t *testing.T) {
	if etaSeconds(1000, 0) != -1 {
		t.Fatal("zero speed → -1")
	}
	if etaSeconds(1000, 500) != 2 {
		t.Fatal("1000/500 → 2s")
	}
}
