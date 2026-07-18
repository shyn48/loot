package tui

import (
	"fmt"

	"simple-gui/core"
)

// formatHMS renders seconds as HH:MM:SS, or an em dash when unknown (negative).
func formatHMS(sec int) string {
	if sec < 0 {
		return "—"
	}
	return fmt.Sprintf("%02d:%02d:%02d", sec/3600, (sec%3600)/60, sec%60)
}

// stats are the aggregate counts shown in the header bar.
type stats struct {
	active, completed, queued, errors int
	totalSpeed                        float64
}

// sortRows groups in-progress downloads first and completed ones last,
// preserving original order within each group so the cursor stays stable.
func sortRows(rows []core.JobStatus) []core.JobStatus {
	active := make([]core.JobStatus, 0, len(rows))
	done := make([]core.JobStatus, 0, len(rows))
	for _, r := range rows {
		if r.State == core.StateDone {
			done = append(done, r)
		} else {
			active = append(active, r)
		}
	}
	return append(active, done...)
}

// activeCount is the number of leading non-completed rows — where the dashed
// group separator is drawn.
func activeCount(rows []core.JobStatus) int {
	n := 0
	for _, r := range rows {
		if r.State != core.StateDone {
			n++
		}
	}
	return n
}

func computeStats(rows []core.JobStatus) stats {
	var s stats
	for _, r := range rows {
		switch r.State {
		case core.StateDownloading, core.StateMerging:
			s.active++
			s.totalSpeed += r.Speed
		case core.StateDone:
			s.completed++
		case core.StateQueued:
			s.queued++
		case core.StateFailed:
			s.errors++
		}
	}
	return s
}
