package core

import (
	"fmt"
	"testing"
	"time"
)

func TestConcurrencyQueue(t *testing.T) {
	body := makeBody(200000)
	srv := slowRangeServer(body, 512, 8*time.Millisecond)
	defer srv.Close()

	m := newTestManager(t)
	m.maxActive = 2

	var ids []string
	for i := 0; i < 4; i++ {
		id, err := m.Add(srv.URL + fmt.Sprintf("/f%d.bin", i))
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}

	sawQueued := false
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		active, queued, done := 0, 0, 0
		for _, s := range m.Snapshot() {
			switch s.State {
			case StateDownloading, StateMerging:
				active++
			case StateQueued:
				queued++
			case StateDone:
				done++
			}
		}
		if active > m.maxActive {
			t.Fatalf("active downloads exceeded cap: %d > %d", active, m.maxActive)
		}
		if queued > 0 {
			sawQueued = true
		}
		if done == 4 {
			break
		}
		time.Sleep(15 * time.Millisecond)
	}

	if !sawQueued {
		t.Fatal("expected some downloads to be queued behind the cap")
	}
	for _, id := range ids {
		if st := stateOf(m, id); st != StateDone {
			t.Fatalf("job %s did not complete: %s", id, st)
		}
	}
}
