package core

import (
	"context"
	"sync"
	"time"
)

type JobState string

const (
	StateQueued      JobState = "queued"
	StateDownloading JobState = "downloading"
	StatePaused      JobState = "paused"
	StateMerging     JobState = "merging"
	StateDone        JobState = "done"
	StateFailed      JobState = "failed"
)

// Job is one download. It owns its section goroutines; bytes-downloaded is not
// stored here but derived from temp file sizes on disk (see downloadedBytes).
type Job struct {
	ID, URL, Filename, TargetPath string
	Size                          int64
	AcceptRanges                  bool
	Sections                      [][2]int64
	TotalSection                  int

	mu             sync.Mutex
	state          JobState
	err            error
	cancel         context.CancelFunc
	speed          float64 // EWMA bytes/sec
	lastBytes      int64
	lastSampleTime time.Time
}

// JobStatus is an immutable snapshot of a Job for rendering by the front-ends.
type JobStatus struct {
	ID, Name         string
	Size, Downloaded int64
	Percent, Speed   float64
	ETASeconds       int
	State            JobState
	Resumable        bool
	Err              string
}

// statusWithDownloaded builds the render DTO. Caller supplies the current
// downloaded byte count (derived from disk) to keep this pure and testable.
func (j *Job) statusWithDownloaded(downloaded int64) JobStatus {
	var pct float64
	if j.Size > 0 {
		pct = float64(downloaded) / float64(j.Size) * 100
	}
	errStr := ""
	if j.err != nil {
		errStr = j.err.Error()
	}
	return JobStatus{
		ID:         j.ID,
		Name:       j.Filename,
		Size:       j.Size,
		Downloaded: downloaded,
		Percent:    pct,
		Speed:      j.speed,
		ETASeconds: etaSeconds(j.Size-downloaded, j.speed),
		State:      j.state,
		Resumable:  j.AcceptRanges && j.Size > 0,
		Err:        errStr,
	}
}
