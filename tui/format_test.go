package tui

import (
	"testing"

	"simple-gui/core"
)

func TestFormatHMS(t *testing.T) {
	cases := map[int]string{
		-1:    "—",
		0:     "00:00:00",
		52:    "00:00:52",
		97:    "00:01:37",
		3661:  "01:01:01",
		45296: "12:34:56",
	}
	for sec, want := range cases {
		if got := formatHMS(sec); got != want {
			t.Errorf("formatHMS(%d) = %q, want %q", sec, got, want)
		}
	}
}

func TestComputeStats(t *testing.T) {
	rows := []core.JobStatus{
		{State: core.StateDownloading, Speed: 1000},
		{State: core.StateDownloading, Speed: 2000},
		{State: core.StatePaused},
		{State: core.StateQueued},
		{State: core.StateQueued},
		{State: core.StateDone},
		{State: core.StateFailed},
	}
	s := computeStats(rows)
	if s.active != 2 || s.completed != 1 || s.queued != 2 || s.errors != 1 {
		t.Fatalf("counts wrong: %+v", s)
	}
	if s.totalSpeed != 3000 {
		t.Fatalf("totalSpeed = %f, want 3000", s.totalSpeed)
	}
}
